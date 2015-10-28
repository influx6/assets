package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// createCompressWriter creates a non-gzip compressed writer
func createUnCompressWriter(w io.Writer) io.WriteCloser {
	return &nopWriter{w}
}

// createCompressWriter creates a gzip compressed writer
func createCompressWriter(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(&StringWriter{W: w})
}

// sanitize prepares a valid UTF-8 string as a raw string constant.
func sanitize(b []byte) []byte {
	// Replace ` with `+"`"+`
	b = bytes.Replace(b, []byte("`"), []byte("`+\"`\"+`"), -1)

	// Replace BOM with `+"\xEF\xBB\xBF"+`
	// (A BOM is valid UTF-8 but not permitted in Go source files.
	// I wouldn't bother handling this, but for some insane reason
	// jquery.js has a BOM somewhere in the middle.)
	return bytes.Replace(b, []byte("\xEF\xBB\xBF"), []byte("`+\"\\xEF\\xBB\\xBF\"+`"), -1)
}

// readData takes a compressed gzip bytes and decompress it unless the virtual file wants no decompression
func readData(v *VFile, data []byte) ([]byte, error) {
	if !v.Decompress {
		return data, nil
	}
	// reader, err := gzip.NewReader(strings.NewReader(data))
	reader, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("---> VFile.readData.error: read file %q at %q, due to: %q\n", v.Name(), v.Path(), err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	clerr := reader.Close()

	if err != nil {
		return nil, fmt.Errorf("---> VFile.readData.error: read file %q at %q, due to gzip reader error: %q\n", v.Name(), v.Path(), err)
	}

	if clerr != nil {
		return nil, clerr
	}

	return buf.Bytes(), nil
}

func readGZipFile(v *VFile) ([]byte, error) {
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
}

func readFile(v *VFile) ([]byte, error) {
	fo, err := ioutil.ReadFile(v.RealPath())
	if err != nil {
		return nil, fmt.Errorf("---> assets.readFile: Error reading file: %s at %s: %s\n", v.Name(), v.RealPath(), err)
	}
	return fo, nil
}

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

type nopWriter struct {
	w io.Writer
}

func (n *nopWriter) Close() error {
	return nil
}

func (n *nopWriter) Write(b []byte) (int, error) {
	return n.w.Write(b)
}
