package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ruggi/md/types"
)

func Load(mdDir string) (types.Config, error) {
	var conf types.Config
	f, err := os.Open(filepath.Join(mdDir, "config.json"))
	if err != nil {
		return types.Config{}, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		return types.Config{}, errors.Wrap(err, "cannot parse config file")
	}
	return conf, nil
}
