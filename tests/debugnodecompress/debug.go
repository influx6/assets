//Auto-generated from github.com/influx6/assets
// DO NOT CHANGE

// Package debug provides an auto-generated static embeding of data files within the specific directory C:\Users\Flux\Lab\go\cmd\src\github.com\influx6\assets
package debug

  
import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)
	
// RootDirectory defines a directory root for these virtual files
var RootDirectory = NewDirCollector()


// httpFile represents a basic http.FileSystem valid file
type httpFile struct {
	*bytes.Reader
	*VFile
}

// VDir defines a virtual directory structure
type VDir struct {
	*VFile
	FileMutex sync.RWMutex
	Files     FileCollector
	SubMutex  sync.RWMutex
	Subs      DeferDirCollector
}

// NewVDir creates a new VirtualDirectory
func NewVDir(moddedPath, realPath, abs string) *VDir {
	vf := VFile{
		BaseDir:   abs,
		Dir:       moddedPath,
		ShadowDir: realPath,
		FileName:  filepath.Base(moddedPath),
		Mod:       time.Now(),
	}

	return &VDir{
		VFile: &vf,
		Files: NewFileCollector(),
		Subs:  NewDeferDirCollector(),
	}
}

// ErrNotFound is Returned When a File/Directory path is not found
var ErrNotFound = errors.New("File/Directory path is not found")

// IsDir returns true for VDir
func (vd *VDir) IsDir() bool {
	return true
}

// DeferVDir defines a function type that returns a VDir
type DeferVDir func() *VDir

// AddDirectory adds a directory listing into the virtual directory
func (vd *VDir) AddDirectory(path string, vf DeferVDir) {
	vd.SubMutex.Lock()
	defer vd.SubMutex.Unlock()
	vd.Subs.Set(path, vf)
}

// Readdir meets the Readdir interface requirements
func (vd *VDir) Readdir(count int) ([]os.FileInfo, error) {
	var total = count
	var files []os.FileInfo

	vd.Files.Each(func(v *VFile, _ string, stop func()) {
		if total <= 0 {
			stop()
			return
		}

		files = append(files, v)
		total--
	})

	return files, nil
}

// Open meets the http.FileSystem interface requirements
func (vd *VDir) Open(file string) (http.File, error) {
	vf, err := vd.GetFile(file)
	if err != nil {
		return nil, err
	}

	data, _ := vf.Data()
	return &httpFile{
		Reader: bytes.NewReader(data),
		VFile:  vf,
	}, nil
}

// GetFile gets the file set within its pathway or its sub-directories pathway
func (vd *VDir) GetFile(file string) (*VFile, error) {
	file = cleanPath(file)
	//grab the base name again,just incase we dealing with a directory like path eg doc/box/file.go
	basename := filepath.Base(file)
	// dirPath := filepath.Dir(file)

	// log.Printf("dir: %s name: %s", dirPath, basename)

	dir, err := vd.GetDir(file)

	if err != nil {
		return nil, err
	}

	if dir == vd {
		// log.Print("in self: %s", vd.Files)

		var file *VFile

		vd.FileMutex.RLock()
		if vd.Files.Has(basename) {
			file = vd.Files.Get(basename)
		}
		vd.FileMutex.RUnlock()

		if file == nil {
			return nil, os.ErrNotExist
		}

		return file, nil
	}

	return dir.GetFile(basename)
}

// GetDir loads the path if available and returns the VDir corresponding to that path
func (vd *VDir) GetDir(m string) (*VDir, error) {
	if m == "" {
		return nil, os.ErrNotExist
	}

	vd.SubMutex.RLock()
	defer vd.SubMutex.RUnlock()

	if vd.Subs.Has(m) {
		return vd.Subs.Get(m)(), nil
	}

	file := cleanPath(m)

	if file == "." || file == "/" {
		return vd, nil
	}

	if vd.Subs.Has(file) {
		return vd.Subs.Get(file)(), nil
	}

	//grab the base name again,just incase we dealing with a file like path eg doc/box/file.go
	// basename := filepath.Base(file)
	dirPath := cleanPath(filepath.Dir(file))

	// log.Printf("dir: %s -> %s", dirPath, file)

	if dirPath == "." {
		return vd, nil
	}

	//its not a current path, but a subpath,so get the first piece then pass down to that
	var parts = strings.Split(dirPath, "/")
	var first = parts[0]

	if vd.Subs.Has(first) {
		return vd.GetDir(strings.Join(parts[1:], "/"))
	}

	return nil, os.ErrNotExist
}

// AddFile adds a virtual file into the virtual directory
func (vd *VDir) AddFile(vf *VFile) {
	vd.FileMutex.Lock()
	defer vd.FileMutex.Unlock()
	vd.Files.Set(vf.Name(), vf)
}

// Close does nothing
func (vd *VDir) Close() error {
	return nil
}

// DataPack represents the function that returns the underline data
type DataPack func(*VFile) ([]byte, error)

// VFile or virtual file for provide a virtual file info
type VFile struct {
	// Compressed    bool
	Decompress    bool
	ShadowDir     string
	BaseDir       string
	Dir           string
	FileName      string
	Datasize      int64
	processedPack []byte
	DataPack      DataPack
	Mod           time.Time
	cache         bool
}

// NewVFile creates a new VirtualFile
func NewVFile(pwd, modded, real string, size int64, decompress bool, fx DataPack) *VFile {
	mdir := filepath.Dir(modded)
	rdir := filepath.Dir(real)
	vf := VFile{
		// Compressed: compressed,
		Decompress: decompress,
		BaseDir:    pwd,
		Dir:        mdir,
		ShadowDir:  rdir,
		FileName:   filepath.Base(modded),
		Mod:        time.Now(),
		Datasize:   size,
		DataPack:   fx,
	}

	return &vf
}

// RealPath returns the true path of the file/dir on the filesystem, this is usually the same with the Path() but if a path mutation occured this returns the original path
func (v *VFile) RealPath() string {
	return filepath.Join(v.BaseDir, v.ShadowDir, v.FileName)
}

// Path returns the path of the file/dir
func (v *VFile) Path() string {
	return filepath.Join(v.BaseDir, v.Dir, v.FileName)
}

// Name returns the name of the file/dir
func (v *VFile) Name() string {
	return v.FileName
}

// Stat returns itself
func (v *VFile) Stat() (os.FileInfo, error) {
	return v, nil
}

// Sys returns nil
func (v *VFile) Sys() interface{} {
	return nil
}

// Readdir meets the Readdir interface requirements
func (v *VFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

// Data returns the data captured within
func (v *VFile) Data() ([]byte, error) {
	if v.DataPack == nil {
		return v.processedPack, nil
	}

	if v.cache && v.processedPack != nil {
		// if v.processedPack != nil {
		return v.processedPack, nil
		// }
	}

	pack, err := v.DataPack(v)

	if err != nil {
		return nil, err
	}

	v.processedPack = pack

	return pack, nil
}

// Mode returns 0 as the filemode
func (v *VFile) Mode() os.FileMode {
	return 0
}

// Size returns the size of the data
func (v *VFile) Size() int64 {
	if v.processedPack != nil {
		return int64(len(v.processedPack))
	}
	return v.Datasize
}

// ModTime returns the modtime for the virtual file
func (v *VFile) ModTime() time.Time {
	return v.Mod
}

// Close does nothing
func (v *VFile) Close() error {
	return nil
}

// IsDir returns false
func (v *VFile) IsDir() bool {
	return false
}

// FileCollector defines a typ of map string
type FileCollector map[string]*VFile

// NewFileCollector returns a new FileCollector
func NewFileCollector() FileCollector {
	return make(FileCollector)
}

// Clone makes a new clone of this FileCollector
func (c FileCollector) Clone() FileCollector {
	col := make(FileCollector)
	col.Copy(c)
	return col
}

// Remove deletes a key:value pair
func (c FileCollector) Remove(k string) {
	if c.Has(k) {
		delete(c, k)
	}
}

// Keys return the keys of the FileCollector
func (c FileCollector) Keys() []string {
	var keys []string
	c.Each(func(_ *VFile, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

// Get returns the value with the key
func (c FileCollector) Get(k string) *VFile {
	return c[k]
}

// Has returns if a key exists
func (c FileCollector) Has(k string) bool {
	_, ok := c[k]
	return ok
}

// HasMatch checks if key and value exists and are matching
func (c FileCollector) HasMatch(k string, v *VFile) bool {
	if c.Has(k) {
		return c.Get(k) == v
	}
	return false
}

// Set puts a specific key:value into the FileCollector
func (c FileCollector) Set(k string, v *VFile) {
	c[k] = v
}

// Copy copies the map into the FileCollector
func (c FileCollector) Copy(m map[string]*VFile) {
	for v, k := range m {
		c.Set(v, k)
	}
}

// Each iterates through all items in the FileCollector
func (c FileCollector) Each(fx func(*VFile, string, func())) {
	var state bool
	for k, v := range c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
}

// Clear clears the FileCollector
func (c FileCollector) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// DirCollector defines a typ of map string
type DirCollector map[string]*VDir

// NewDirCollector returns a new FileCollector
func NewDirCollector() DirCollector {
	return make(DirCollector)
}

// Clone makes a new clone of this DirCollector
func (c DirCollector) Clone() DirCollector {
	col := make(DirCollector)
	col.Copy(c)
	return col
}

// GetFile gets the VFile for the specific file if existing
func (c DirCollector) GetFile(path string) (*VFile, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}

	dirPath, file := filepath.Split(cleanPath(path))

	dir, err := c.GetDir(dirPath)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return dir.GetFile(file)
}

// GetDir gets the given directory path and returns a VirtualDirectory
func (c DirCollector) GetDir(dir string) (*VDir, error) {
	if dir == "" {
		return nil, os.ErrNotExist
	}

	dir = cleanPath(dir)

	if dir == "." || dir == "/" {
		return c.Root(), nil
	}

	if c.Has(dir) {
		return c.Get(dir), nil
	}

	dirPath, _ := filepath.Split(dir)
	dirPath = cleanPath(dirPath)

	if c.Has(dirPath) {
		return c.Get(dirPath), nil
	}

	parts := strings.Split(dir, "/")
	first := parts[0]

	if c.Has(first) {
		return c.Get(first).GetDir(strings.Join(parts[1:], "/"))
	}

	return nil, os.ErrNotExist
}

// Root gets the root path found in the list,either a "." or a "/"
func (c DirCollector) Root() *VDir {
	if c.Has("/") {
		return c.Get("/")
	}
	return c.Get(".")
}

// Remove deletes a key:value pair
func (c DirCollector) Remove(k string) {
	if c.Has(k) {
		delete(c, k)
	}
}

// Keys return the keys of the DirCollector
func (c DirCollector) Keys() []string {
	var keys []string
	c.Each(func(_ *VDir, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

// Get returns the value with the key
func (c DirCollector) Get(k string) *VDir {
	return c[k]
}

// Has returns if a key exists
func (c DirCollector) Has(k string) bool {
	_, ok := c[k]
	return ok
}

// HasMatch checks if key and value exists and are matching
func (c DirCollector) HasMatch(k string, v *VDir) bool {
	if c.Has(k) {
		return c.Get(k) == v
	}
	return false
}

// Set puts a specific key:value into the DirCollector
func (c DirCollector) Set(k string, v *VDir) {
	c[k] = v
}

// Copy copies the map into the DirCollector
func (c DirCollector) Copy(m map[string]*VDir) {
	for v, k := range m {
		c.Set(v, k)
	}
}

// Each iterates through all items in the DirCollector
func (c DirCollector) Each(fx func(*VDir, string, func())) {
	var state bool
	for k, v := range c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
}

// Clear clears the DirCollector
func (c DirCollector) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// DeferDirCollector defines a typ of map string
type DeferDirCollector map[string]func() *VDir

// NewDeferDirCollector returns a new FileCollector
func NewDeferDirCollector() DeferDirCollector {
	return make(DeferDirCollector)
}

// Clone makes a new clone of this DeferDirCollector
func (c DeferDirCollector) Clone() DeferDirCollector {
	col := make(DeferDirCollector)
	col.Copy(c)
	return col
}

// Remove deletes a key:value pair
func (c DeferDirCollector) Remove(k string) {
	if c.Has(k) {
		delete(c, k)
	}
}

// Keys return the keys of the DeferDirCollector
func (c DeferDirCollector) Keys() []string {
	var keys []string
	c.Each(func(_ func() *VDir, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

// Get returns the value with the key
func (c DeferDirCollector) Get(k string) func() *VDir {
	return c[k]
}

// Has returns if a key exists
func (c DeferDirCollector) Has(k string) bool {
	_, ok := c[k]
	return ok
}

// Set puts a specific key:value into the DeferDirCollector
func (c DeferDirCollector) Set(k string, v func() *VDir) {
	c[k] = v
}

// Copy copies the map into the DeferDirCollector
func (c DeferDirCollector) Copy(m map[string]func() *VDir) {
	for v, k := range m {
		c.Set(v, k)
	}
}

// Each iterates through all items in the DeferDirCollector
func (c DeferDirCollector) Each(fx func(func() *VDir, string, func())) {
	var state bool
	for k, v := range c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
}

// Clear clears the DeferDirCollector
func (c DeferDirCollector) Clear() {
	for k := range c {
		delete(c, k)
	}
}

func cleanPath(dir string) string {
	dir = filepath.ToSlash(filepath.Clean(dir))

	if len(dir) == 1 {
		return dir
	}

	if dir[0] == '/' {
		dir = dir[1:]
	}

	size := len(dir)

	if size <= 1 {
		return dir
	}

	if dir[size-1] == '/' {
		dir = dir[:size-2]
	}

	return dir
}

func init(){

  RootDirectory.Set("fixtures",func() *VDir{
    var dir = NewVDir("fixtures","fixtures","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/fixtures")

    // register the sub-directories
    
	dir.AddDirectory("base",func() *VDir{
		return RootDirectory.Get("fixtures/base")
	})



	dir.AddDirectory("includes",func() *VDir{
		return RootDirectory.Get("fixtures/includes")
	})



	dir.AddDirectory("layouts",func() *VDir{
		return RootDirectory.Get("fixtures/layouts")
	})



    // register the files
    

    return dir
  }())

}


func init(){

  RootDirectory.Set("tests/debug",func() *VDir{
    var dir = NewVDir("tests/debug","tests/debug","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/tests/debug")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/debug/debug.go","tests/debug/debug.go",28217,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/debug/debug_test.go","tests/debug/debug_test.go",929,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("tests/debugnodecompress",func() *VDir{
    var dir = NewVDir("tests/debugnodecompress","tests/debugnodecompress","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/tests/debugnodecompress")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/debugnodecompress/debug_test.go","tests/debugnodecompress/debug_test.go",929,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/debugnodecompress/debug.go","tests/debugnodecompress/debug.go",0,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("tests/prodnodecompression",func() *VDir{
    var dir = NewVDir("tests/prodnodecompression","tests/prodnodecompression","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/tests/prodnodecompression")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/prodnodecompression/prod.go","tests/prodnodecompression/prod.go",2836107,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/prodnodecompression/prod_test.go","tests/prodnodecompression/prod_test.go",928,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("/",func() *VDir{
    var dir = NewVDir("/",".","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets")

    // register the sub-directories
    
	dir.AddDirectory("fixtures",func() *VDir{
		return RootDirectory.Get("fixtures")
	})



	dir.AddDirectory("tests",func() *VDir{
		return RootDirectory.Get("tests")
	})



    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","vim.md","vim.md",5,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","binddata_test.go","binddata_test.go",1768,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","functions.go","functions.go",421,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","pkg.go","pkg.go",1457,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","templates.go","templates.go",7311,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","utils.go","utils.go",3714,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","vfiles.go","vfiles.go",12494,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets",".travis.yml",".travis.yml",63,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","assetlists.go","assetlists.go",7178,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","readme.md","readme.md",2112,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","safetrees.go","safetrees.go",3349,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","vfiles_test.go","vfiles_test.go",3693,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","watchables.go","watchables.go",3780,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","assets.go","assets.go",2649,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","assets_test.go","assets_test.go",4574,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","binddata.go","binddata.go",7151,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","errors.go","errors.go",632,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","formats.go","formats.go",15296,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","stringwriter.go","stringwriter.go",452,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("fixtures/base",func() *VDir{
    var dir = NewVDir("fixtures/base","fixtures/base","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/fixtures/base")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","fixtures/base/basic.tmpl","fixtures/base/basic.tmpl",364,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","fixtures/base/index.tmpl","fixtures/base/index.tmpl",181,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("fixtures/includes",func() *VDir{
    var dir = NewVDir("fixtures/includes","fixtures/includes","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/fixtures/includes")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","fixtures/includes/index.tmpl","fixtures/includes/index.tmpl",80,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("fixtures/layouts",func() *VDir{
    var dir = NewVDir("fixtures/layouts","fixtures/layouts","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/fixtures/layouts")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","fixtures/layouts/basic.tmpl","fixtures/layouts/basic.tmpl",364,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}


func init(){

  RootDirectory.Set("tests",func() *VDir{
    var dir = NewVDir("tests","tests","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/tests")

    // register the sub-directories
    
	dir.AddDirectory("debug",func() *VDir{
		return RootDirectory.Get("tests/debug")
	})



	dir.AddDirectory("debugnodecompress",func() *VDir{
		return RootDirectory.Get("tests/debugnodecompress")
	})



	dir.AddDirectory("prod",func() *VDir{
		return RootDirectory.Get("tests/prod")
	})



	dir.AddDirectory("prodnodecompression",func() *VDir{
		return RootDirectory.Get("tests/prodnodecompression")
	})



    // register the files
    

    return dir
  }())

}


func init(){

  RootDirectory.Set("tests/prod",func() *VDir{
    var dir = NewVDir("tests/prod","tests/prod","C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets/tests/prod")

    // register the sub-directories
    

    // register the files
    
		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/prod/prod_test.go","tests/prod/prod_test.go",928,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

		dir.AddFile(NewVFile("C:/Users/Flux/Lab/go/cmd/src/github.com/influx6/assets","tests/prod/prod.go","tests/prod/prod.go",1889281,false,func(v *VFile) ([]byte, error) {
			fo, err := os.Open(v.RealPath())
			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			defer fo.Close()

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)

			_, err = io.Copy(gz, fo)
			gz.Close()

			if err != nil {
				return nil, fmt.Errorf("---> assets.readFile.gzip: Error gzipping file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
			}

			return buf.Bytes(), nil
		}))
	

    return dir
  }())

}
