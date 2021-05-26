package files

import (
	"os"
	"strings"
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
