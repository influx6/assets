#Assets
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/assets)
[![Travis](https://travis-ci.org/influx6/assets.svg?branch=master)](https://travis-ci.org/influx6/assets)

Provides a convenient set of tools for handling template loading

##Example

  ```go

	dir := NewTemplateDir(&TemplateConfig{
		Dir:       "./fixtures",
		Extension: ".tmpl",
	})

	dirs := []string{"base"}

	asst, _ := dir.Create("base.tmpl", dirs, nil)

	buf := bytes.NewBuffer([]byte{})

	do := &dataPack{
		Name:  "alex",
		Title: "flabber",
	}

	_ = asst.Tmpl.ExecuteTemplate(buf, "base", do)

  /*
   buf => `

            <html>
                   <head>


                   </head>
                   <body>

                 <div class=alex>flabber</div>

                   <i>we are equal</i>


                   </body>
                 </html>
   `

  */

  ```
