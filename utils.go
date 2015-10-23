package assets

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ByName Implement sort.Interface for []os.FileInfo based on Name()
type ByName []os.FileInfo

func (v ByName) Len() int           { return len(v) }
func (v ByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ByName) Less(i, j int) bool { return v[i].Name() < v[j].Name() }

func makeAbsolute(path string) string {
	as, _ := filepath.Abs(path)
	return as
}

// makeRelative removes the if found the beginning slash '/'
func makeRelative(path string) string {
	if path[0] == '/' {
		return path[1:]
	}
	return path
}

func hasIn(paths []string, dt string) bool {
	for _, so := range paths {
		if strings.Contains(so, dt) || so == dt {
			return true
		}
	}
	return false
}

func hasExt(paths []string, dt string) bool {
	for _, so := range paths {
		if so == dt {
			return true
		}
	}
	return false
}

func getDirListings(dir string) ([]os.FileInfo, error) {
	//open up the filepath since its a directory, read and sort
	cdir, err := os.Open(dir)

	if err != nil {
		return nil, err
	}

	defer cdir.Close()

	files, err := cdir.Readdir(0)

	if err != nil {
		return nil, err
	}

	sort.Sort(ByName(files))

	return files, nil
}
