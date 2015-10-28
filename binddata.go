package assets

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// DevelopmentMode represents development mode for bfs files
const DevelopmentMode = 0

// ProductionMode repesents a production assembly mode for bfs
const ProductionMode = 1

// BindFSConfig provides a configuration struct for BindFS
type BindFSConfig struct {
	InDir      string        //directory path use as source
	OutDir     string        //directory path to save file in
	Package    string        //package name for the file
	File       string        //file name of the file
	Gzipped    bool          // to enable gzipping
	Production bool          // to enable production mode as default
	ValidPath  PathValidator //use to filter allowed paths
	Mux        PathMux       //use to mutate path look
}

// BindFS provides the struct for creating and updating a go file containing static assets from a directory
type BindFS struct {
	config      *BindFSConfig
	listing     *DirListing
	mode        int64
	endpoint    string
	endpointDir string
	inputDir    string
	curDir      string
}

// NewBindFS returns a new BindFS instance or an error if it fails to located directory
func NewBindFS(config BindFSConfig) (*BindFS, error) {
	vali := config.ValidPath
	mux := config.Mux

	(&config).ValidPath = func(path string, in os.FileInfo) bool {
		if strings.Contains(path, ".git") {
			return false
		}

		if vali != nil {
			return vali(path, in)
		}

		return true
	}

	//clean out pathway or use custom muxer if provided
	(&config).Mux = func(path string, in os.FileInfo) string {
		if mux != nil {
			return mux(path, in)
		}
		if filepath.Clean(path) == "." {
			return "/"
		}
		return path
	}

	ls, err := DirListings(config.InDir, config.ValidPath, config.Mux)

	if err != nil {
		return nil, fmt.Errorf("---> BindFS: Unable to create dirlisting for %s -> %s", config.InDir, err)
	}

	bf := BindFS{
		config:  &config,
		listing: ls,
	}

	if config.Production {
		bf.ProductionMode()
	}

	return &bf, nil
}

// Mode returns the current mode of the BindFS
func (bfs *BindFS) Mode() int {
	return int(atomic.LoadInt64(&bfs.mode))
}

// DevMode switches BindFS operations into dev mode
func (bfs *BindFS) DevMode() {
	atomic.StoreInt64(&bfs.mode, DevelopmentMode)
}

// ProductionMode switches BindFS operations into production mode and embeds file data into their corresponding vfiles
func (bfs *BindFS) ProductionMode() {
	atomic.StoreInt64(&bfs.mode, ProductionMode)
}

// Record dumps all the files and dir listings with their corresponding data into a go file within the specified path
func (bfs *BindFS) Record() error {
	pwd, _ := os.Getwd()
	input := filepath.Join(pwd, bfs.config.InDir)
	endpoint := filepath.Join(pwd, bfs.config.OutDir, bfs.config.File+".go")
	endpointDir := filepath.Dir(endpoint)
	pkgHeader := fmt.Sprintf(packageDetails, bfs.config.Package, input, bfs.config.Package)

	err := os.MkdirAll(endpointDir, 0700)

	if err != nil && err != os.ErrExist {
		return err
	}

	//remove the file for safety and to reduce bloated ouput if file was added in list
	os.Remove(endpoint)

	boutput, err := os.Create(endpoint)

	if err != nil && err != os.ErrExist {
		return err
	}

	defer boutput.Close()

	// var output = boutput
	var output = bytes.NewBuffer([]byte{})

	//writes the library package header
	fmt.Fprint(output, pkgHeader)

	//writes the library imports
	if bfs.Mode() > 0 {
		// if !bfs.config.Gzipped {
		// fmt.Fprint(output, uncomImports)
		// } else {
		fmt.Fprint(output, comImports)
		// }
	} else {
		fmt.Fprint(output, debugImports)
	}

	//writing the libraries core
	fmt.Fprint(output, rootDir)
	fmt.Fprint(output, structBase)

	if bfs.Mode() > 0 {
		if !bfs.config.Gzipped {
			fmt.Fprint(output, uncomFunc)
		} else {
			fmt.Fprint(output, comFunc)
		}
	}

	// log.Printf("tree: %s", bfs.listing.Listings.Tree)

	//go through the directories listings
	bfs.listing.EachDir(func(dir *BasicAssetTree, path string) {
		// log.Printf("walking dir: %s", path)

		path = filepath.ToSlash(filepath.Clean(path))
		modDir := filepath.ToSlash(filepath.Clean(dir.ModDir))
		pathDir := filepath.ToSlash(filepath.Clean(dir.Dir))
		pathAbs := filepath.ToSlash(filepath.Clean(dir.AbsDir))

		//fill up the directory content
		dirContent := fmt.Sprintf(dirRegister, path, modDir, pathDir, pathAbs)

		var subs []string
		var data []string

		// go through the subdirectories list and added them
		dir.EachChild(func(child *BasicAssetTree) {
			//add the sub-directories
			// log.Printf("Walking Child: %s", child.AbsDir)
			childDir := filepath.ToSlash(filepath.Clean(child.ModDir))
			subs = append(subs, fmt.Sprintf(subRegister, filepath.Base(childDir), childDir))
		})

		//loadup the files
		dir.Tree.Each(func(modded, real string) {
			// log.Printf("Walking: %s -> File: %s %s", dir.AbsDir, modded, real)

			modded = filepath.ToSlash(filepath.Clean(modded))
			real = filepath.ToSlash(filepath.Clean(real))
			cleanPwd := filepath.ToSlash(filepath.Clean(pwd))

			var output string
			if bfs.Mode() == DevelopmentMode {
				stat, _ := os.Stat(filepath.Join(pwd, real))
				output = fmt.Sprintf(debugFile, cleanPwd, modded, real, stat.Size(), fileRead)
			} else {
				//production mode is active,we need to load the file contents

				file, err := os.Open(real)

				if err != nil {
					fmt.Printf("---> BindFS.error: failed to loadup %s file -> %s", real, err)
					return
				}

				var data bytes.Buffer
				var writer io.WriteCloser

				if bfs.config.Gzipped {
					writer = createCompressWriter(&data)
				} else {
					writer = createUnCompressWriter(&data)
				}

				n, _ := io.Copy(writer, file)
				file.Close()
				writer.Close()

				var bu []byte

				if bfs.config.Gzipped {
					bu = sanitize(data.Bytes())
				} else {
					bu = data.Bytes()
				}

				var format string
				if bfs.config.Gzipped {
					stringed := fmt.Sprintf("%q", bu)
					stringed = strings.Replace(stringed, `\\`, `\`, -1)
					format = fmt.Sprintf(prodRead, stringed)
				} else {
					format = fmt.Sprintf(prodRead, fmt.Sprintf("`%s`", bu))
				}

				output = fmt.Sprintf(debugFile, cleanPwd, modded, real, n, format)
			}

			data = append(data, output)
		})

		dirContent = strings.Replace(dirContent, "{{ subs }}", strings.Join(subs, "\n"), -1)
		dirContent = strings.Replace(dirContent, "{{ files }}", strings.Join(data, "\n"), -1)

		fmt.Fprint(output, fmt.Sprintf(rootInit, dirContent))
	})

	io.Copy(boutput, output)
	return nil
}
