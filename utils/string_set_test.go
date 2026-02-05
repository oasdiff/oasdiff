package utils_test

import (
	"testing"

	"github.com/oasdiff/oasdiff/utils"
	"github.com/stretchr/testify/require"
)

func TestMinus_Self(t *testing.T) {
	s := utils.StringSet{}
	s.Add("x")
	s.Add("y")
	require.Empty(t, s.Minus(s))
}

func TestMinus_Partial(t *testing.T) {
	s1 := utils.StringSet{}
	s1.Add("x")
	s1.Add("y")

	s2 := utils.StringSet{}
	s2.Add("x")

	require.Equal(t, []string{"y"}, s1.Minus(s2).ToStringList())
	require.Empty(t, s2.Minus(s1))
}

func TestIntersection_Self(t *testing.T) {
	s := utils.StringSet{}
	s.Add("x")
	s.Add("y")
	require.Equal(t, s, s.Intersection(s))
}

func TestIntersection_Empty(t *testing.T) {
	s := utils.StringSet{}
	s.Add("x")
	s.Add("y")
	require.Empty(t, s.Intersection(utils.StringSet{}))
}

func TestStringSet_Plus(t *testing.T) {
	s := utils.StringSet{}
	s.Add("x")
	require.True(t, s.Equals(s.Plus(s)))
}

func TestStringSet_Equals(t *testing.T) {
	s := utils.StringSet{}
	s.Add("x")
	require.True(t, s.Equals(s))
}
