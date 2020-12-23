package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeSLIP(t *testing.T) {

	t.Run("multiple packets and rest", func(t *testing.T) {
		packets, rest, err := decodeSLIP([]byte{
			0xA, 0xB, slipEnd,
			0xC, 0xD, slipEnd,
			0xE, 0xF,
		})

		require.NoError(t, err)
		require.Equal(t, [][]byte{{0xA, 0xB}, {0xC, 0xD}}, packets)
		require.Equal(t, []byte{0xE, 0xF}, rest)
	})

	t.Run("escaped end", func(t *testing.T) {
		packets, rest, err := decodeSLIP([]byte{
			0xA, 0xB, slipEsc, slipEscEnd, 0xC, 0xD, slipEnd,
		})

		require.NoError(t, err)
		require.Equal(t, [][]byte{{0xA, 0xB, slipEnd, 0xC, 0xD}}, packets)
		require.Empty(t, rest)
	})

	t.Run("escaped escape", func(t *testing.T) {
		packets, rest, err := decodeSLIP([]byte{
			0xA, 0xB, slipEsc, slipEscEsc, 0xC, 0xD, slipEnd,
		})

		require.NoError(t, err)
		require.Equal(t, [][]byte{{0xA, 0xB, slipEsc, 0xC, 0xD}}, packets)
		require.Empty(t, rest)
	})

	t.Run("escaped other", func(t *testing.T) {
		_, _, err := decodeSLIP([]byte{
			0xA, 0xB, slipEsc, 0xC, 0xD, slipEnd,
		})

		require.Error(t, err)
	})

}
