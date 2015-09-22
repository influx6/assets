package assets

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/influx6/flux"
)

func TestAssetMap(t *testing.T) {
	tree, err := AssetTree("./", ".go")

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
