package assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//AssetMap provides a map of paths that contain assets of the specific filepaths
type AssetMap map[string]string

// ReloadAssetMap reloads the files into the map skipping the already found ones
func ReloadAssetMap(tree AssetMap, dir string, ext []string, skip []string) error {
	var stat os.FileInfo
	var err error

	//do the path exists
	if stat, err = os.Stat(dir); err != nil {
		return err
	}

	//do we have a directory?
	if !stat.IsDir() {

		if tree.Has(filepath.ToSlash(dir)) {
			return nil
		}

		var fext string
		var rel = filepath.Base(dir)

		if strings.Index(rel, ".") != -1 {
			fext = filepath.Ext(rel)
		}

		if len(ext) > 0 {
			if hasExt(ext, fext) {
				tree[rel] = filepath.ToSlash(dir)
			}
		} else {
			tree[rel] = filepath.ToSlash(dir)
		}

	} else {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

			//if info is nil or is a directory when we skip
			if info == nil || info.IsDir() {
				return nil
			}

			repath := filepath.ToSlash(path)

			if tree.Has(repath) {
				return nil
			}

			if strings.Contains(repath, ".git") {
				return nil
			}

			if strings.Index(repath, ".git") != -1 {
				return nil
			}

			if hasIn(skip, repath) {
				return nil
			}

			var rel string
			var rerr error

			//is this path relative to the current one, if not,return err
			if rel, rerr = filepath.Rel(dir, path); rerr != nil {
				return rerr
			}

			var fext string

			if strings.Index(rel, ".") != -1 {
				fext = filepath.Ext(rel)
			}

			if len(ext) > 0 {
				if hasExt(ext, fext) {
					tree[rel] = filepath.ToSlash(path)
				}
			} else {
				tree[rel] = filepath.ToSlash(path)
			}

			return nil
		})
	}
	return nil
}

// AssetTree provides a map tree of files across the given directory that match the filenames being used
func AssetTree(dir string, ext, skip []string) (AssetMap, error) {
	//do the path exists
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	var tree = make(AssetMap)

	err := ReloadAssetMap(tree, dir, ext, skip)

	return tree, err
}

// AssetLoader is a function type which returns the content in []byte of a specific asset
type AssetLoader func(string) ([]byte, error)

// Has returns true/false if the filename without its extension exists
func (am AssetMap) Has(name string) bool {
	_, ok := am[name]
	return ok
}

// Load returns the data of the specific file with the name
func (am AssetMap) Load(name string) ([]byte, error) {
	if !am.Has(name) {
		return nil, NewCustomError("AssetMap", fmt.Sprintf("%s unknown", name))
	}
	return ioutil.ReadFile(am[name])
}
