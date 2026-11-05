package kgs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	groupsCount  = 4
	groupsLen    = 5
)

func TestGenerateKey(t *testing.T) {
	t.Run("test alphabet", func(t *testing.T) {
		tKGS := New(groupsCount, groupsLen)

		k, err := tKGS.GenerateKey()
		require.NoError(t, err)

		groups := strings.Split(k, "-")
		require.Len(t, groups, groupsCount)

		for _, group := range groups {
			require.Len(t, group, groupsLen)
			for _, ch := range group {
				require.True(t, strings.ContainsRune(testAlphabet, ch))
			}
		}
	})

	t.Run("test length", func(t *testing.T) {
		tKGS := New(groupsCount, groupsLen)

		k, err := tKGS.GenerateKey()
		require.NoError(t, err)

		expectedLen := groupsCount*groupsLen + (groupsCount - 1)
		require.Len(t, k, expectedLen)
	})
}

func BenchmarkGenerateKey(b *testing.B) {
	kgs := New(groupsCount, groupsLen)
	b.ResetTimer()

	for b.Loop() {
		_, _ = kgs.GenerateKey()
	}
}
