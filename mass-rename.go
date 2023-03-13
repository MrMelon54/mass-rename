package mass_rename

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidMappingLine = errors.New("invalid mapping line")

type MappedName struct {
	Old string
	New string
}

func MassRename(fs afero.Fs, m []MappedName) []error {
	errs := make([]error, len(m))
	oldDirs := make(map[string]struct{})

	// move files one by one
	for i, item := range m {
		// old and new path
		op, np := item.Old, item.New
		// save oldDirs for clean up
		oldDirs[filepath.Dir(op)] = struct{}{}
		// make new directory
		err := fs.MkdirAll(filepath.Dir(np), os.ModePerm)
		if err != nil {
			errs[i] = fmt.Errorf("failed to create directory tree: %w", err)
			continue
		}
		err = fs.Rename(op, np)
		if err != nil {
			errs[i] = fmt.Errorf("failed to rename file '%s' => '%s': %w", op, np, err)
			continue
		}
	}

	// clean up old directories
	for i := range oldDirs {
		dir, err := afero.ReadDir(fs, i)
		if err != nil {
			continue
		}
		if len(dir) == 0 {
			err := fs.Remove(i)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func ParseMassRenameMap(b io.Reader) ([]MappedName, error) {
	a := make([]MappedName, 0)
	buf := bufio.NewScanner(b)
	for buf.Scan() {
		s := strings.Split(buf.Text(), " => ")
		if len(s) == 2 {
			if s[0] != s[1] {
				a = append(a, MappedName{s[0], s[1]})
			}
		} else {
			return nil, ErrInvalidMappingLine
		}
	}
	return a, nil
}
