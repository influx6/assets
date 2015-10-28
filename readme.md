#Assets
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/assets)
[![Travis](https://travis-ci.org/influx6/assets.svg?branch=master)](https://travis-ci.org/influx6/assets)

Provides a convenient set of tools for handling template files and turning assets into embeddable go files

##Example

  - Emdedding
  
       *Note to run the tests in test/* subfolders first run `go test` to generate the need files*

    - To embed a given directory but in development mode(loading from disk) but also gzipping output
    ```go

    	bf, err := NewBindFS(BindFSConfig{
    		InDir:   "./",
    		OutDir:     "./tests/debug",
    		Package: "debug",
    		File:    "debug",
    		Gzipped: true,
                NoDecompression: true,
                Production: false,
    	})

    	if err != nil {
             panic("directory path is not valid")
    	}

      //to get this to create and embed the files,simple call .Record()
    	err = bf.Record() // you can call this as many times as you want to update go file

    ```

    - To embed a given directory but in development mode,where files are loaded directory from disk
    ```go

    	bf, err := NewBindFS(BindFSConfig{
    		InDir:   "./",
    		OutDir:     "./tests/debug",
    		Package: "debug",
    		File:    "debug",
    		Gzipped: false,
                Production: false,
    	})

    	if err != nil {
          panic("directory path is not valid")
    	}

      //to get this to create and embed the files,simple call .Record()
    	err = bf.Record() // you can call this as many times as you want to update go file

    ```

    - To embed files in production mode,i.e all assets are embedded into the generate go file have output decompressed

    ```go
    	bf, err := NewBindFS(BindFSConfig{
    		InDir:      "./",
    		OutDir:     "./tests/prod",
    		Package:    "prod",
    		File:       "prod",
    		Gzipped:    true,
    		Production: true,
    	})

    	if err != nil {
          panic("directory path is not valid")
    	}

    	err = bf.Record() // you can call this as many times as you want to update go file

    ```

    - To embed a given directory in production mode but also enforcing no decompression of output
    ```go

    	bf, err := NewBindFS(BindFSConfig{
    		InDir:   "./",
    		OutDir:     "./tests/debug",
    		Package: "debug",
    		File:    "debug",
    		Gzipped: true,
                NoDecompression: true,
                Production: true,
    	})

    	if err != nil {
          panic("directory path is not valid")
    	}

      //to get this to create and embed the files,simple call .Record()
    	err = bf.Record() // you can call this as many times as you want to update go file

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
