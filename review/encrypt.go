package review

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
)

// BlobVersion is the first byte of the uploaded blob. It lets the decryptor
// reject a format it doesn't understand instead of trying to decrypt garbage.
// Bump it only on an incompatible layout change.
const BlobVersion = 1

// Encrypt marshals the payload to JSON and encrypts it with a freshly generated
// 256-bit key using AES-256-GCM. It returns the upload blob and the key. The
// blob layout is: version(1) || nonce(12) || ciphertext+tag. The key is
// returned to the caller, which keeps it out of band (never uploaded), so the
// server receives ciphertext it cannot read. A decryptor on the rendering side
// reverses this exact layout.
func (p Payload) Encrypt() (blob, key []byte, err error) {
	plaintext, err := json.Marshal(p)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal review payload: %w", err)
	}

	key = make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	blob = make([]byte, 0, 1+len(nonce)+len(ciphertext))
	blob = append(blob, BlobVersion)
	blob = append(blob, nonce...)
	blob = append(blob, ciphertext...)
	return blob, key, nil
}
