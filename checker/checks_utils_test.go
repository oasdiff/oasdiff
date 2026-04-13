package checker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterfaceToString_Nil(t *testing.T) {
	require.Equal(t, "undefined", interfaceToString(nil))
}

func TestInterfaceToString_String(t *testing.T) {
	require.Equal(t, "hello", interfaceToString("hello"))
}

func TestInterfaceToString_StringList(t *testing.T) {
	require.Equal(t, "a, b, c", interfaceToString([]string{"a", "b", "c"}))
}

func TestInterfaceToString_Int(t *testing.T) {
	require.Equal(t, "42", interfaceToString(42))
}

func TestInterfaceToString_Uint64(t *testing.T) {
	require.Equal(t, "100", interfaceToString(uint64(100)))
}

func TestInterfaceToString_Float64(t *testing.T) {
	require.Equal(t, "3.14", interfaceToString(3.14))
}

func TestInterfaceToString_Bool(t *testing.T) {
	require.Equal(t, "true", interfaceToString(true))
	require.Equal(t, "false", interfaceToString(false))
}

func TestInterfaceToString_Fallback(t *testing.T) {
	// A type not explicitly handled falls through to fmt.Sprintf("%v", arg)
	require.Equal(t, "[1 2 3]", interfaceToString([]int{1, 2, 3}))
}
