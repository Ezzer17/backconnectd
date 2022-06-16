package rpc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	tests := []struct {
		in  [][]byte
		exp [][]byte
	}{
		{
			[][]byte{[]byte("fadsfasdf")},
			[][]byte{},
		},
		{
			[][]byte{[]byte("fasdf\x00")},
			[][]byte{[]byte("fasdf")},
		},
		{
			[][]byte{[]byte("aaaaa\x00aaaaaa")},
			[][]byte{[]byte("aaaaa")},
		},
		{
			[][]byte{[]byte("aaaaa\x00bbbbbbbb\x00")},
			[][]byte{[]byte("aaaaa"), []byte("bbbbbbbb")},
		},
		{
			[][]byte{[]byte("aaaaa\x00bb"), []byte("bbbbbb\x00")},
			[][]byte{[]byte("aaaaa"), []byte("bbbbbbbb")},
		},
		{
			[][]byte{[]byte("aaaaa\x00bb"), []byte("bbbbbb\x00\x00")},
			[][]byte{[]byte("aaaaa"), []byte("bbbbbbbb")},
		},
		{
			[][]byte{[]byte("aaaaa\x00bb"), []byte("bbbbbb")},
			[][]byte{[]byte("aaaaa")},
		},
		{
			[][]byte{[]byte("aaaaa\x00\x00bb"), []byte("bbbbbb")},
			[][]byte{[]byte("aaaaa")},
		},
	}
	for testnum, test := range tests {
		result := [][]byte{}
		r := Reader()
		for _, in := range test.in {
			out := r(in)
			result = append(result, out...)
		}
		assert.Equal(t, test.exp, result, fmt.Sprintf("Test %d failed", testnum))
	}
}
