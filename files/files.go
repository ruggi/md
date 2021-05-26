package files

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/ruggi/md/settings"
)

func ListToMap(files []string) map[string][]string {
	result := map[string][]string{}
	for _, f := range files {
		parts := strings.Split(f, string(os.PathSeparator))
		nonEmptyParts := make([]string, 0, len(parts))
		for _, p := range parts {
			if p == "" {
				continue
			}
			nonEmptyParts = append(nonEmptyParts, p)
		}
		for nest := 0; nest < len(nonEmptyParts); nest++ {
			prefix := strings.Join(nonEmptyParts[:nest], string(os.PathSeparator))
			result[prefix] = append(result[prefix], strings.Join(nonEmptyParts, string(os.PathSeparator)))
		}
	}
	return result
}

func EnsureMDDir(dir string) (string, error) {
	mdDir := filepath.Join(dir, settings.MDDir)
	stat, err := os.Stat(mdDir)
	if os.IsNotExist(err) {
		return "", errors.Errorf("%s does not exist, please use the 'init' command to set it up", mdDir)
	}
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", errors.Errorf("%s exists but is not a directory", mdDir)
	}
	return mdDir, nil
}
