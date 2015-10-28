#Assets
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/assets)
[![Travis](https://travis-ci.org/influx6/assets.svg?branch=master)](https://travis-ci.org/influx6/assets)

Provides a convenient set of tools for handling template files and turning assets into embeddable go files

##Example

  - Emdedding

    ```go
    	bf, err := NewBindFS(BindFSConfig{
    		InDir:   "./",
    		Dir:     "./tests/debug",
    		Package: "debug",
    		File:    "debug",
    		Gzipped: false,
    	})

    	if err != nil {
        panic("directory path is not valid")
    	}

      //to get this to create and embed the files,simple call .Record()
    	err = bf.Record()

    ```


  - Templates
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
