package assets

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/influx6/flux"
)

func TestListings(t *testing.T) {
	tree, err := DirListings("./", func(dir string, info os.FileInfo) bool {
		if strings.Contains(dir, ".git") {
			return false
		}
		return true
	}, func(dir string, info os.FileInfo) string {
		return filepath.Join("static/", dir)
	})

	if err != nil {
		flux.FatalFailed(t, "Unable to create asset map: %s", err.Error())
	}

	if tree.Listings.Size() <= 0 || tree.Listings.Size() > 5 {
		flux.FatalFailed(t, "expected size to be below 5")
	}

	flux.LogPassed(t, "Succesfully created directory listings")

	err = os.Mkdir("./fixtures/bucker", 0700)

	if err != nil {
		flux.FatalFailed(t, "Unable to create dir: %s", err.Error())
	}

	defer os.Remove("./fixtures/bucker")

	err = tree.Reload()

	if err != nil {
		flux.FatalFailed(t, "Unable to reload listings: %s", err.Error())
	}

	if tree.Listings.Size() <= 0 || tree.Listings.Size() < 5 {
		flux.FatalFailed(t, "expected size to be above 5")
	}

	flux.LogPassed(t, "Succesfully reloaded directory listings")
}

func TestAssetListings(t *testing.T) {
	tree, err := Assets("./", func(dir string, info os.FileInfo) bool {
		if strings.Contains(dir, ".git") {
			return false
		}
		return true
	}, func(dir string, info os.FileInfo) string {
		return filepath.Join("static/", dir)
	})

	if err != nil {
		flux.FatalFailed(t, "Unable to create asset map: %s", err.Error())
	}

	if tree.Size() <= 0 {
		flux.FatalFailed(t, "expected one key atleast: %s")
	}

	flux.LogPassed(t, "Succesfully created asset map")
}

// BenchmarkAssetListings tests the speed it takes to load up a directory listings
func BenchmarkListings(t *testing.B) {
	for i := 0; i < t.N; i++ {
		tree, err := DirListings("./", func(dir string, info os.FileInfo) bool {
			if strings.Contains(dir, ".git") {
				return false
			}
			return true
		}, func(dir string, info os.FileInfo) string {
			return filepath.Join("static/", dir)
		})
		if err == nil {
			tree.Reload()
		}
	}
}

// BenchmarkAssetListings tests the speed it takes to load up a directory listings
func BenchmarkAssetListings(t *testing.B) {
	for i := 0; i < t.N; i++ {
		Assets("./", func(dir string, info os.FileInfo) bool {
			if strings.Contains(dir, ".git") {
				return false
			}
			return true
		}, func(dir string, info os.FileInfo) string {
			return filepath.Join("static/", dir)
		})
	}
}

func TestAssetMap(t *testing.T) {
	tree, err := AssetTree("./", []string{".go"}, nil)

	if err != nil {
		flux.FatalFailed(t, "Unable to create asset map: %s", err.Error())
	}

	if len(tree) <= 0 {
		flux.FatalFailed(t, "expected one key atleast: %s")
	}

	flux.LogPassed(t, "Succesfully created asset map %+s", tree)
}

type dataPack struct {
	Name  string
	Title string
}

func TestBasicAssets(t *testing.T) {
	tmpl, err := LoadTemplates("./fixtures/base", ".tmpl", nil, []template.FuncMap{DefaultTemplateFunctions})

	if err != nil {
		flux.FatalFailed(t, "Unable to load templates: %+s", err)
	}

	buf := bytes.NewBuffer([]byte{})

	do := &dataPack{
		Name:  "alex",
		Title: "world war - z",
	}

	err = tmpl.ExecuteTemplate(buf, "base", do)

	if err != nil {
		flux.FatalFailed(t, "Unable to exec templates: %+s", err)
	}

	flux.LogPassed(t, "Loaded Template succesfully: %s", string(buf.Bytes()))
}

func TestTemplateDir(t *testing.T) {
	dir := NewTemplateDir(&TemplateConfig{
		Dir:       "./fixtures",
		Extension: ".tmpl",
	})

	dirs := []string{"base"}

	asst, err := dir.Create("base.tmpl", dirs, nil)

	if err != nil {
		flux.FatalFailed(t, "Failed to load: %s", err.Error())
	}

	if len(asst.Funcs) < 1 {
		flux.FatalFailed(t, "AssetTemplate Func map is empty")
	}

	buf := bytes.NewBuffer([]byte{})

	do := &dataPack{
		Name:  "alex",
		Title: "flabber",
	}

	err = asst.Tmpl.ExecuteTemplate(buf, "base", do)

	if err != nil {
		flux.FatalFailed(t, "Unable to exec templates: %+s", err)
	}

	flux.LogPassed(t, "Loaded Template succesfully: %s", string(buf.Bytes()))
}

func TestTemplateAssets(t *testing.T) {
	dirs := []string{"./fixtures/includes/index.tmpl", "./fixtures/layouts"}
	asst, err := NewAssetTemplate("home.html", ".tmpl", dirs)

	if err != nil {
		flux.FatalFailed(t, "Failed to load: %s", err.Error())
	}

	buf := bytes.NewBuffer([]byte{})

	do := &dataPack{
		Name:  "alex",
		Title: "flabber",
	}

	err = asst.Tmpl.ExecuteTemplate(buf, "base", do)

	if err != nil {
		flux.FatalFailed(t, "Unable to exec templates: %+s", err)
	}

	flux.LogPassed(t, "Loaded Template succesfully: %s", string(buf.Bytes()))
}
