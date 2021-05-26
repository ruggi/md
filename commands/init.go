package commands

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ruggi/md/settings"
	"github.com/ruggi/md/types"
)

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

var defaultConfig = types.Config{
	SyntaxHighlight: types.SyntaxHighlightConfig{
		Enabled:     true,
		Style:       "solarized-light",
		LineNumbers: true,
	},
}

type InitArgs struct {
	Directory string
}

func Init(args InitArgs) error {
	mdDir := filepath.Join(args.Directory, settings.MDDir)
	created := false

	stat, err := os.Stat(mdDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(mdDir, os.ModePerm)
		if err != nil {
			return err
		}
		created = true
	} else {
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return errors.Errorf("%s exists but is not a directory", mdDir)
		}
	}

	layoutPath := filepath.Join(mdDir, "layout.html")
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

	configFile := filepath.Join(mdDir, "config.json")
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(configFile, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
