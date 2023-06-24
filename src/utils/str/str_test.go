package str

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringList(t *testing.T) {
	assert.Equal(t, []string{"hello"}, StringList("hello"))
	assert.Equal(t, []string{"hello", "world"}, StringList([]string{"hello", "world"}))
	assert.Equal(t, []string{"hello", "world", "!"}, StringList("hello", []string{"world"}, "!"))
	assert.Equal(t, []string{}, StringList())
	assert.Panics(t, func() {
		StringList(123)
	})
}

func TestRemoveDuplicatesAndSort(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Test 1",
			input:    []string{"apple", "banana", "apple", "orange", "banana", "orange", "apple", "banana"},
			expected: []string{"apple", "banana", "orange"},
		},
		{
			name:     "Test 2",
			input:    []string{"car", "bike", "car", "truck", "bike", "truck", "car", "bike"},
			expected: []string{"bike", "car", "truck"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RemoveDuplicatesAndSort(&tc.input)
			assert.Equal(t, tc.expected, tc.input)
		})
	}
}

func TestConvertToRelativePaths(t *testing.T) {
	// Test case 1: Absolute paths
	elements := []string{
		"/absolute/path/file1.txt",
		"/absolute/path/subdir/file3.txt",
		"/absolute/other/path/file2.txt",
	}

	basepath := "/absolute/path"

	err := ConvertToRelativePaths(&elements, basepath)
	assert.NoError(t, err)

	expected := []string{
		"file1.txt",
		"subdir/file3.txt",
		"../other/path/file2.txt",
	}

	assert.Equal(t, expected, elements)

	// Test case 2: Invalid base path
	elements = []string{
		"/absolute/path/file1.txt",
	}

	basepath = ""

	err = ConvertToRelativePaths(&elements, basepath)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to convert /absolute/path/file1.txt to a relative path")

	// Test case 3: Relative paths, no modifications
	elements = []string{
		"file1.txt",
		"rel2/file2.txt",
		"rel3/rel3/file3.txt",
	}

	basepath = "/absolute/path"

	err = ConvertToRelativePaths(&elements, basepath)
	assert.NoError(t, err)
	assert.Equal(t, []string{"file1.txt", "rel2/file2.txt", "rel3/rel3/file3.txt"}, elements)
}
