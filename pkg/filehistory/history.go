// Package =filehistory provides a basic set of interfaces to keeping
// history in a file
package filehistory

import (
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
	if h.createNewHistoryFileIfNeeded(); err != nil {
		return nil, errors.Wrap(err, "checking history file")
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

	// TODO this will fail if commands contain a \n. But that's unix history files for you...
	fullhistory = strings.Split(string(historyBytes), "\n")

	return fullhistory[len(h.history-count):]

}

func (h *FileHistory) Append(line string) error {
	if h.historyfile.WriteString(line + "\n"); err != nil {
		return errors.Wrap(err, "writing history")
	}

	// TODO: consider skipping this, it may add write
	// overhead. But, then we'd need something to ensure we close
	// on exist.
	if h.historyfile.Sync(); err != nil {
		return errors.Wrap(err, "syncing history")
	}

}

func (h *FileHistory) guessHistoryFile() error {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "looking up current user")
	}

	h.historypath = filepath.Join(usr.HomeDir, ".goquery", "history")
	return nil
}

func (h *FileHistory) createNewHistoryFileIfNeeded() error {
	filestat, err := os.Stat(basedir)
	if err == nil && filestat.IsRegular() {
		return
	}

	// Does the parent directory exist?
	basedir := filepathBase(h.historypath)
	dirstat, err := os.Stat(basedir)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(goQueryPath, os.ModePerm); err != nil {
			return errors.Wrap(err, "Making goquery directory")
		}
	case err != nil:
		return errors.Wrap(err, "stating ~/.goquery")
	case src.Mode().IsRegular():
		return errors.Errorf("%s already exists, and is not a directory. Cannot make history file", basedir)
	}

	// Now make the history file
	// Skip the stat step, since this whole function is about making new ones
	if h.historyfile, err = os.Create(historyPath); err != nil {
		return errors.Wrap(err, "creating new history file")
	}

	return nil

}
