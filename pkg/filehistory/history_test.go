package filehistory

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRecent(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		fileLength   int
		recentLength int
		expected     []string
	}{
		{
			fileLength:   0,
			recentLength: 20,
			expected:     []string{""}, // FIXME this is probably a bug
		},
		{
			fileLength:   5,
			recentLength: 4,
			expected:     []string{"2", "3", "4", "5"},
		},
		{
			fileLength:   4,
			recentLength: 4,
			expected:     []string{"1", "2", "3", "4"},
		},
		{
			fileLength:   4,
			recentLength: 5,
			expected:     []string{"1", "2", "3", "4"},
		},
	}

	for _, tt := range tests {
		tmpFile := makeTestFile(t, tt.fileLength)
		h, err := New(tmpFile)
		defer os.RemoveAll(tmpFile)
		require.NoError(t, err, "new history")
		require.Equal(t, tt.expected, h.GetRecent(tt.recentLength))
	}

}

func TestHistory_NewDir(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "test_filehistory")
	require.NoError(t, err, "make temp dir")
	defer os.RemoveAll(dir)

	h, err := New(filepath.Join(dir, "newdir", "history"))
	defer h.Close()
	require.NoError(t, err, "make history")
	testAppend(t, h)
}

func TestHistory_NewFile(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "test_filehistory")
	require.NoError(t, err, "make temp dir")
	defer os.RemoveAll(dir)

	h, err := New(filepath.Join(dir, "history"))
	defer h.Close()
	require.NoError(t, err, "make history")
	testAppend(t, h)

}

func TestHistory_FileExists(t *testing.T) {
	t.Parallel()

	filename, err := ioutil.TempFile("", "test_filehistory")
	require.NoError(t, err, "make temp file")
	require.NoError(t, filename.Close(), "close")
	defer os.RemoveAll(filename.Name())

	h, err := New(filename.Name())
	defer h.Close()
	require.NoError(t, err, "make history")
	testAppend(t, h)
}

func testAppend(t *testing.T, h *FileHistory) {
	command := randomCommand() + " " + randomCommand()
	require.NoError(t, h.Append(command), "append")

	contents, err := ioutil.ReadFile(h.historyfile.Name())
	require.NoError(t, err, "reading history from disk")
	require.Equal(t, command+"\n", string(contents))
}

func makeTestFile(t *testing.T, count int) string {
	filename, err := ioutil.TempFile("", "test_filehistory")
	require.NoError(t, err, "make temp file")
	defer filename.Close()

	for i := 1; i <= count; i++ {
		_, err := filename.WriteString(strconv.Itoa(i) + "\n")
		require.NoError(t, err, "write")
	}
	require.NoError(t, filename.Close(), "close")
	return filename.Name()
}

func randomCommand() string {
	letterBytes := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 12)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
