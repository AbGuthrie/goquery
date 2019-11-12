// Package =filehistory provides a basic set of interfaces to keeping
// history in a file
package filehistory

type FileHistory struct {
	filepath string
	history  []string
}

func New(filepath string) (*FileHistory, error) {
}

func (h *FileHistory) GetAll() []string {
}

func (h *FileHistory) GetRecent(int) []string {
}

func (h *FileHistory) Append(string) error {
}
