package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/ruggi/md/engine"
	"github.com/ruggi/md/files"
	"github.com/ruggi/md/settings"
	"github.com/ruggi/md/types"
)

type BuildArgs struct {
	Directory string
}

type BuildCmd func() error

func NewBuild(args BuildArgs, engine engine.Engine) (BuildCmd, error) {
	mdDir, err := files.EnsureMDDir(args.Directory)
	if err != nil {
		return nil, err
	}
	outDir := filepath.Join(mdDir, settings.OutDir)

	// config setup
	readConfig := func() (types.Config, error) {
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
	conf, err := readConfig()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config file")
	}
	engine.SetSyntaxHighlight(conf.SyntaxHighlight)

	return func() error {
		log.Printf("Building %s", args.Directory)

		start := time.Now()

		// cleanup
		err = os.RemoveAll(outDir)
		if err != nil {
			return err
		}

		// get source files
		paths := []string{}
		err = filepath.WalkDir(args.Directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasPrefix(path, mdDir) {
				return nil
			}
			if !settings.SourceFileExtensions[filepath.Ext(path)] {
				copyFile := func() error {
					log.Println("copying", path)

					src, err := os.Open(path)
					if err != nil {
						return errors.Wrapf(err, "cannot open source %s", path)
					}
					defer src.Close()

					fld := filepath.Join(outDir, strings.TrimSuffix(path, filepath.Base(path)))
					err = os.MkdirAll(fld, os.ModePerm)
					if err != nil {
						return errors.Wrapf(err, "cannot make folders for %s", fld)
					}

					dstPath := filepath.Join(outDir, strings.TrimPrefix(path, args.Directory))
					dst, err := os.Create(dstPath)
					if err != nil {
						return errors.Wrapf(err, "cannot open destination %s", dstPath)
					}
					defer dst.Close()

					_, err = io.Copy(dst, src)
					if err != nil {
						return errors.Wrapf(err, "canont copy %s", path)
					}

					return nil
				}
				return copyFile()
			}
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "cannot list files in %s", args.Directory)
		}

		pages := map[string]types.Page{}
		for _, p := range paths {
			defaultTitle := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
			title := defaultTitle
			f, err := os.Open(p)
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(f)
			if scanner.Scan() {
				firstLine := scanner.Text()
				if strings.HasPrefix(firstLine, "# ") {
					title = strings.TrimPrefix(firstLine, "# ")
				} else if strings.HasPrefix(firstLine, "<!--page ") {
					var meta types.Page
					err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(firstLine, "<!--page "), " -->")), &meta)
					if err != nil {
						return err
					}
					title = meta.Title
				}
			}
			_ = f.Close()

			pages[p] = types.Page{
				Title: title,
				Path:  p,
			}
		}
		pathsWithoutDir := make([]string, 0, len(paths))
		for _, p := range paths {
			pathsWithoutDir = append(pathsWithoutDir, strings.TrimPrefix(p, args.Directory))
		}
		rawPagesMap := files.ListToMap(pathsWithoutDir)
		pagesMap := make(map[string][]types.Page)
		for k, v := range rawPagesMap {
			mapPage := make([]types.Page, 0, len(v))
			for _, path := range v {
				page := pages[filepath.Join(args.Directory, path)]
				page.Path = strings.ReplaceAll(
					strings.TrimSuffix(strings.TrimPrefix(page.Path, args.Directory), filepath.Ext(page.Path))+".html",
					string(os.PathSeparator),
					"/",
				)
				mapPage = append(mapPage, page)
			}
			if k == "" {
				k = "_"
			}
			pagesMap[k] = mapPage
		}

		layout, err := ioutil.ReadFile(filepath.Join(mdDir, "layout.html"))
		if err != nil {
			return err
		}
		layoutTpl, err := template.New("layout").Parse(string(layout))
		if err != nil {
			return err
		}

		// build output files
		for _, path := range paths {
			pathWithoutDir := strings.TrimPrefix(path, args.Directory)
			log.Println(pathWithoutDir)

			src, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			dstFilename := strings.TrimSuffix(pathWithoutDir, filepath.Ext(pathWithoutDir)) + ".html"
			dstFile := filepath.Join(outDir, dstFilename)
			dstDir := strings.TrimSuffix(dstFile, filepath.Base(dstFilename))
			err = os.MkdirAll(dstDir, os.ModePerm)
			if err != nil {
				return err
			}

			page := pages[path]

			fileTpl, err := template.New("file").Parse(string(src))
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			err = fileTpl.Execute(&buf, types.FileData{
				Page:  page,
				Pages: pagesMap,
			})
			if err != nil {
				return err
			}

			var contentBuf bytes.Buffer
			err = engine.Convert(&buf, &contentBuf)
			if err != nil {
				return err
			}

			dst, err := os.Create(dstFile)
			if err != nil {
				return err
			}
			defer dst.Close()

			err = layoutTpl.Execute(dst, types.Layout{
				Page:    page,
				Content: contentBuf.String(),
			})
			if err != nil {
				return err
			}
		}

		log.Printf("✔ Done (%s)", time.Since(start))

		return nil
	}, nil
}
