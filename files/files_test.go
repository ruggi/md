package files_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ruggi/md/files"
	"github.com/stretchr/testify/require"
)

func TestListToMap(t *testing.T) {
	tests := []struct {
		files    []string
		expected map[string][]string
	}{
		{
			files:    []string{},
			expected: map[string][]string{},
		},
		{
			files: []string{"a", "b", "c"},
			expected: map[string][]string{
				"": {"a", "b", "c"},
			},
		},
		{
			files: []string{"a", "b", "c", "d/e"},
			expected: map[string][]string{
				"":  {"a", "b", "c", "d/e"},
				"d": {"d/e"},
			},
		},
		{
			files: []string{"a", "b", "c", "d/e", "d/f", "d/g", "e/1", "e/2/3", "e/2/4", "e/3/4", "e/3/5"},
			expected: map[string][]string{
				"":    {"a", "b", "c", "d/e", "d/f", "d/g", "e/1", "e/2/3", "e/2/4", "e/3/4", "e/3/5"},
				"d":   {"d/e", "d/f", "d/g"},
				"e":   {"e/1", "e/2/3", "e/2/4", "e/3/4", "e/3/5"},
				"e/2": {"e/2/3", "e/2/4"},
				"e/3": {"e/3/4", "e/3/5"},
			},
		},
		{
			files: []string{
				"a",
				"b/c",
				"d",
				"d/e/f/g",
				"d/e/f/h",
				"d/e/z/i",
				"d/z/z/z",
			},
			expected: map[string][]string{
				"": {
					"a",
					"b/c",
					"d",
					"d/e/f/g",
					"d/e/f/h",
					"d/e/z/i",
					"d/z/z/z",
				},
				"b": {
					"b/c",
				},
				"d": {
					"d/e/f/g",
					"d/e/f/h",
					"d/e/z/i",
					"d/z/z/z",
				},
				"d/e": {
					"d/e/f/g",
					"d/e/f/h",
					"d/e/z/i",
				},
				"d/e/f": {
					"d/e/f/g",
					"d/e/f/h",
				},
				"d/e/z": {
					"d/e/z/i",
				},
				"d/z": {
					"d/z/z/z",
				},
				"d/z/z": {
					"d/z/z/z",
				},
			},
		},
	}
	for _, tt := range tests {
		jsonExp, err := json.Marshal(tt.expected)
		require.NoError(t, err)

		got := files.ListToMap(tt.files)
		jsonGot, err := json.MarshalIndent(got, "", " ")
		require.NoError(t, err)

		fmt.Println(string(jsonGot))

		require.JSONEq(t, string(jsonExp), string(jsonGot))
	}
}
