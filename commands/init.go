package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ruggi/md/settings"
)

type InitArgs struct {
	Directory string
}

var defaultLayout = `
<html>
	<head>
		<meta charset="utf-8" />
		<title>{{ .Title }}</title>
	</head>
	<body>
{{ .Content }}
	</body>
</html>
`

var defaultIndex = `
# Hello world

This is a sample page.
`

func Init(args InitArgs) error {
	mdPath := filepath.Join(args.Directory, settings.MDDir)

	stat, err := os.Stat(mdPath)

	created := false
	if os.IsNotExist(err) {
		err = os.MkdirAll(mdPath, os.ModePerm)
		if err != nil {
			return err
		}
		created = true
	} else {
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return errors.Errorf("%s exists and is not a directory", mdPath)
		}
	}

	layoutPath := filepath.Join(mdPath, "layout.html")
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		err = ioutil.WriteFile(layoutPath, []byte(defaultLayout), 0644)
		if err != nil {
			return err
		}
	}

	if created {
		err = ioutil.WriteFile(filepath.Join(args.Directory, "index.md"), []byte(defaultIndex), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
