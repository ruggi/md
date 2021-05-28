package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
	"github.com/ruggi/md/config"
	"github.com/ruggi/md/engine"
	"github.com/ruggi/md/files"
	"github.com/ruggi/md/settings"
)

type ServeArgs struct {
	Directory string
	Host      string
	Port      int
	Watch     bool
}

func Serve(args ServeArgs, engine engine.Engine) error {
	mdDir, err := files.EnsureMDDir(args.Directory)
	if err != nil {
		return err
	}

	outPath := filepath.Join(mdDir, settings.OutDir)

	builder, err := NewBuild(BuildArgs{Directory: args.Directory}, engine)
	if err != nil {
		return err
	}
	err = builder()
	if err != nil {
		return err
	}

	conf, err := config.Load(mdDir)
	if err != nil {
		return errors.Wrap(err, "cannot read config file")
	}

	if args.Watch {
		log.Println("watching for changes")

		w := watcher.New()
		defer w.Close()

		for _, f := range conf.NoWatch {
			err = w.Ignore(f)
			if err != nil {
				return err
			}
		}

		err = w.Add(filepath.Join(mdDir, "layout.html"))
		if err != nil {
			return err
		}

		err = w.Ignore(mdDir)
		if err != nil {
			return err
		}

		err = w.AddRecursive(args.Directory)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case event := <-w.Event:
					log.Println(event.Op, event.Path)
					err := builder()
					if err != nil {
						log.Fatalf("cannot build: %s", err)
					}
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		go func() {
			if err := w.Start(time.Millisecond * 100); err != nil {
				log.Fatal(err)
			}
		}()
	}

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", args.Host, args.Port),
		Handler: http.FileServer(http.Dir(outPath)),
	}

	log.Printf("listening on http://%s", s.Addr)
	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("cannot start server: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	sig := <-sigCh

	log.Printf("got signal %s, terminating...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = s.Shutdown(shutdownCtx)
	if err != nil {
		return err
	}

	return nil
}
