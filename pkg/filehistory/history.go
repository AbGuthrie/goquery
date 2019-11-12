// Package =filehistory provides a basic set of interfaces to keeping
// history in a file
package filehistory

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type FileHistory struct {
	historypath string
	historyfile *os.File
}

func New(historypath string) (*FileHistory, error) {
	h := &FileHistory{
		historypath: historypath,
	}

	// if filepath is blank, try to guess. I'm not sure if this is
	// a good idea here or not.
	if h.historypath == "" {
		if err := h.guessHistoryFile(); err != nil {
			return nil, errors.Wrap(err, "guessing history file location")
		}
	}

	// Try to read history. If that fails, trigger creation of a new one.
	if err := h.openOrCreate(); err != nil {
		return nil, errors.Wrap(err, "open or creating history file")
	}

	return h, nil
}

func (h *FileHistory) Close() error {
	return h.historyfile.Close()
}

func (h *FileHistory) GetRecent(count int) []string {
	// TODO, re-write this to use something to seek to the end,
	// reading the whole file into ram will explode if it get
	// large enough.

	historyBytes, err := ioutil.ReadFile(h.historypath)
	if err != nil {
		return nil
	}

	historyBytes = bytes.TrimSuffix(historyBytes, []byte("\n"))

	// TODO this will fail if commands contain a \n. But that's unix history files for you...
	fullhistory := strings.Split(string(historyBytes), "\n")

	if len(fullhistory) <= count {
		return fullhistory
	}
	return fullhistory[len(fullhistory)-count:]

}

func (h *FileHistory) Append(line string) error {
	if _, err := h.historyfile.WriteString(line + "\n"); err != nil {
		return errors.Wrap(err, "writing history")
	}

	// TODO: consider skipping this, it may add write
	// overhead. But, then we'd need something to ensure we close
	// on exist.
	if err := h.historyfile.Sync(); err != nil {
		return errors.Wrap(err, "syncing history")
	}
	return nil
}

func (h *FileHistory) guessHistoryFile() error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "looking up current user")
	}

	h.historypath = filepath.Join(usr.HomeDir, ".goquery", "history")
	return nil
}

func (h *FileHistory) openOrCreate() error {

	basedir := filepath.Dir(h.historypath)
	basedirStat, err := os.Stat("basedir")
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(basedir, os.ModePerm); err != nil {
			return errors.Wrap(err, "Making goquery directory")
		}
	case !basedirStat.IsDir():
		return errors.Errorf("%s exists and is not a directory", basedir)
	case err != nil:
		return errors.Wrap(err, "error stat'ing basedir")
	}

	if h.historyfile, err = os.OpenFile(h.historypath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return errors.Wrap(err, "opening history for writing")
	}
	return nil

}
