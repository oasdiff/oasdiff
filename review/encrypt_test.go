package review

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// decryptBlob reverses the version(1) || nonce(12) || ciphertext+tag layout
// using the returned key, the test-side mirror of the browser decryptor. If
// this round-trips, a correct WebCrypto implementation will too.
func decryptBlob(t *testing.T, blob, key []byte) []byte {
	t.Helper()
	require.Equal(t, byte(BlobVersion), blob[0], "first byte must be the format version")
	block, err := aes.NewCipher(key)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := blob[1 : 1+gcm.NonceSize()]
	plaintext, err := gcm.Open(nil, nonce, blob[1+gcm.NonceSize():], nil)
	require.NoError(t, err, "ciphertext must decrypt with the returned key")
	return plaintext
}

func TestPayloadEncrypt_RoundTrip(t *testing.T) {
	p := Payload{RevisionSpec: "openapi: 3.0.0", Mode: "changelog", ToolVersion: "v1.25.1"}
	blob, key, err := p.Encrypt()
	require.NoError(t, err)
	require.Len(t, key, 32, "AES-256 needs a 256-bit key")

	// The blob must not contain the spec text anywhere -- it is ciphertext.
	require.NotContains(t, string(blob), "openapi: 3.0.0")

	var got Payload
	require.NoError(t, json.Unmarshal(decryptBlob(t, blob, key), &got))
	require.Equal(t, p.RevisionSpec, got.RevisionSpec)
	require.Equal(t, p.Mode, got.Mode)
	require.Equal(t, p.ToolVersion, got.ToolVersion, "the producing oasdiff version travels inside the bundle")

	// Two encryptions of the same payload must differ (fresh key + nonce), so a
	// server seeing two identical specs can't tell they're identical.
	blob2, key2, err := p.Encrypt()
	require.NoError(t, err)
	require.NotEqual(t, key, key2, "each upload must use a fresh key")
	require.NotEqual(t, blob, blob2, "ciphertext must not be deterministic")
}
