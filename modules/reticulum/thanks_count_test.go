// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseThanksCount(t *testing.T) {
	cases := []struct {
		name    string
		hex     string
		want    int
		wantErr bool
	}{
		{"zero", "81a5636f756e7400", 0, false},
		{"one", "81a5636f756e7401", 1, false},
		{"five", "81a5636f756e7405", 5, false},
		{"127", "81a5636f756e747f", 127, false},
		{"128", "81a5636f756e74cc80", 128, false},
		{"256", "81a5636f756e74cd0100", 256, false},
		{"1000", "81a5636f756e74cd03e8", 1000, false},
		{"65535", "81a5636f756e74cdffff", 65535, false},
		{"100000", "81a5636f756e74ce000186a0", 100000, false},
		{"empty", "", 0, true},
		{"garbage", "ff", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			if tc.hex != "" {
				data = make([]byte, len(tc.hex)/2)
				for i := 0; i < len(data); i++ {
					var b byte
					_, err := parseHexByte(tc.hex[i*2:i*2+2], &b)
					require.NoError(t, err)
					data[i] = b
				}
			}
			got, err := ParseThanksCount(data)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func parseHexByte(s string, out *byte) (int, error) {
	var v byte
	for i := 0; i < 2; i++ {
		c := s[i]
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= c - '0'
		case c >= 'a' && c <= 'f':
			v |= c - 'a' + 10
		case c >= 'A' && c <= 'F':
			v |= c - 'A' + 10
		default:
			return 0, assert.AnError
		}
	}
	*out = v
	return 2, nil
}
