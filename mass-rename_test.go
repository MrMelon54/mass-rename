package mass_rename

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	ioFs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createFileInline(fs afero.Fs, name string) error {
	_ = fs.MkdirAll(filepath.Dir(name), os.ModePerm)
	a, err := fs.Create(name)
	if err != nil {
		return err
	}
	_ = a.Close()
	return nil
}

func printTree(fs afero.Fs) string {
	var s strings.Builder
	_ = afero.Walk(fs, ".", func(path string, info ioFs.FileInfo, err error) error {
		s.WriteString(path + "\n")
		return nil
	})
	return s.String()
}

func TestMassRename(t *testing.T) {
	fs := afero.NewMemMapFs()
	assert.NoError(t, createFileInline(fs, "a.txt"))
	assert.NoError(t, createFileInline(fs, "1/2/c.txt"))
	assert.NoError(t, createFileInline(fs, "1/2/d.txt"))
	assert.NoError(t, createFileInline(fs, "1/4/e.txt"))
	assert.NoError(t, createFileInline(fs, "1/4/f.txt"))
	assert.Equal(t, `.
1
1/2
1/2/c.txt
1/2/d.txt
1/4
1/4/e.txt
1/4/f.txt
a.txt
`, printTree(fs))

	errs := MassRename(fs, []MappedName{
		{"a.txt", "b.txt"},
		{"1/2/c.txt", "1/3/c.txt"},
		{"1/4/e.txt", "1/5/e.txt"},
		{"1/4/f.txt", "1/5/f.txt"},
	})
	for _, i := range errs {
		if i != nil {
			assert.NoError(t, i)
		}
	}

	assert.Equal(t, `.
1
1/2
1/2/d.txt
1/3
1/3/c.txt
1/5
1/5/e.txt
1/5/f.txt
b.txt
`, printTree(fs))
}

func TestParseMassRenameMap(t *testing.T) {
	renameMap, err := ParseMassRenameMap(strings.NewReader(`a.txt => b.txt
1/2 => 1/2
1/2/c.txt => 1/3/c.txt
1/4/e.txt => 1/5/e.txt
1/4/f.txt => 1/5/f.txt`))
	assert.NoError(t, err)
	assert.Equal(t, []MappedName{
		{"a.txt", "b.txt"},
		{"1/2/c.txt", "1/3/c.txt"},
		{"1/4/e.txt", "1/5/e.txt"},
		{"1/4/f.txt", "1/5/f.txt"},
	}, renameMap)
}
