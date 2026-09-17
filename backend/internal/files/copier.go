package files

import (
	"github.com/otiai10/copy"
)

// Copier abstracts copying directory
type Copier interface {
	Copy(src, dest string) error
}

type directoryCopier struct{}

// NewCopier creates and new Copier and returns it
func NewCopier() Copier {
	return directoryCopier{}
}

func (directoryCopier) Copy(src, dst string) error {
	return copy.Copy(src, dst, copy.Options{
		Sync: true,
	})
}
