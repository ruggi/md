package files

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

var dateLayouts = []string{
	"20060102",
}

func TryParseDate(s string) (time.Time, bool) {
	stamp := make([]rune, 0, len(s))
	for _, c := range s {
		if c == '-' || (c >= '0' && c <= '9') {
			stamp = append(stamp, c)
		}
	}
	v := strings.ReplaceAll(string(stamp), "-", "")

	if len(v) == 10 {
		n, err := strconv.Atoi(v)
		if err == nil {
			t := time.Unix(int64(n), 0)
			if !t.IsZero() {
				return t, true
			}
		}
	}
	tryParse := func(layout, v string) (time.Time, bool) {
		t, err := time.Parse(layout, v)
		if err != nil {
			return time.Time{}, false
		}
		if t.IsZero() {
			return time.Time{}, false
		}
		return t, true
	}
	for _, l := range dateLayouts {
		if t, ok := tryParse(l, v); ok {
			return t, true
		}
	}
	return time.Time{}, false
}
