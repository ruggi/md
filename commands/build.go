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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/ruggi/md/config"
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
	conf, err := config.Load(mdDir)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config file")
	}
	engine.SetSyntaxHighlight(conf.SyntaxHighlight)

	runHooks := func(hooks []string) error {
		for i, h := range hooks {
			log.Printf("%d/%d running hook %q", i+1, len(conf.Hooks.Build.Before), h)
			parts := strings.Fields(h)
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
		return nil
	}

	return func() error {
		log.Printf("Building %s", args.Directory)

		start := time.Now()

		// cleanup
		err := os.RemoveAll(outDir)
		if err != nil {
			return err
		}

		err = runHooks(conf.Hooks.Build.Before)
		if err != nil {
			return err
		}

		// get source files
		paths := []string{}
		err = filepath.WalkDir(args.Directory, func(path string, d fs.DirEntry, err error) error {
			for _, s := range conf.Ignore {
				if s == path {
					return nil
				}
			}
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasPrefix(path, mdDir) || strings.HasPrefix(path, filepath.Join(args.Directory, ".git")) {
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

					dir := filepath.Join(outDir, strings.TrimPrefix(strings.TrimSuffix(path, filepath.Base(path)), args.Directory))
					err = os.MkdirAll(dir, os.ModePerm)
					if err != nil {
						return errors.Wrapf(err, "cannot make folders for %s", path)
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
			base := filepath.Base(p)
			defaultTitle := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))

			crumbs := strings.Split(p, "/")
			breadcrumbs := make([]types.Breadcrumb, 0, len(crumbs))
			if crumbs[len(crumbs)-1] != "index.md" {
				for i, c := range crumbs[:len(crumbs)-1] {
					path := "/" + strings.Join(crumbs[:i+1], "/")
					if strings.HasSuffix(path, ".md") {
						path = strings.TrimSuffix(path, ".md") + ".html"
					}
					breadcrumbs = append(breadcrumbs, types.Breadcrumb{
						Path: path,
						Base: c,
					})
				}
			}

			page := types.Page{
				Title:       defaultTitle,
				Path:        "/" + strings.TrimSuffix(p, filepath.Ext(p)) + ".html",
				Breadcrumbs: breadcrumbs,
				Parent:      filepath.Dir(filepath.Dir(p)),
				Base:        base,
				Dir:         filepath.Dir(p),
				IsDir:       base == "index.md",
			}

			date, ok := files.TryParseDate(base)
			if ok {
				page.Date = date
			}

			f, err := os.Open(p)
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(f)
			if scanner.Scan() {
				firstLine := scanner.Text()
				if strings.HasPrefix(firstLine, "# ") {
					page.Title = strings.TrimPrefix(firstLine, "# ")
				} else if strings.HasPrefix(firstLine, "<!-- ") {
					var meta types.Page
					err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(firstLine, "<!-- "), " -->")), &meta)
					if err != nil {
						log.Printf("warning: cannot parse page meta for %s", p)
					}
					page.Title = meta.Title
					page.Date = meta.Date
				}
			}
			_ = f.Close()

			pages[p] = page
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
		layoutTpl, err := template.New("layout").Funcs(template.FuncMap{
			"reverse": reverse,
			"byDate":  byDate,
			"byName":  byName,
		}).Parse(string(layout))
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

			fileTpl, err := template.New("file").Funcs(template.FuncMap{
				"reverse": reverse,
				"byDate":  byDate,
				"byName":  byName,
			}).Parse(string(src))
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

		log.Printf("??? Done (%s)", time.Since(start))

		err = runHooks(conf.Hooks.Build.After)
		if err != nil {
			return err
		}

		return nil
	}, nil
}

func reverse(pages []types.Page, things ...interface{}) []types.Page {
	res := make([]types.Page, len(pages))
	for i := range pages {
		res[len(pages)-i-1] = pages[i]
	}
	return res
}

func byDate(pages []types.Page, things ...interface{}) []types.Page {
	res := make(types.PagesByDate, 0, len(pages))
	for _, p := range pages {
		res = append(res, p)
	}
	sort.Sort(res)
	return res
}

func byName(pages []types.Page, things ...interface{}) []types.Page {
	res := make(types.PagesByName, 0, len(pages))
	for _, p := range pages {
		res = append(res, p)
	}
	sort.Sort(res)
	return res
}
