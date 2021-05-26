package main

import (
	"log"
	"os"

	"github.com/ruggi/md/commands"
	"github.com/ruggi/md/engine/usegoldmark"
	"github.com/urfave/cli"
)

var conf struct {
	directory      string
	highlightStyle string
	serve          struct {
		host  string
		port  int
		watch bool
	}
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "d,directory",
			Value:       ".",
			Destination: &conf.directory,
			Required:    true,
		},
		cli.StringFlag{
			Name:        "highlight",
			Value:       "solarized-light",
			Destination: &conf.highlightStyle,
		},
	}
	app.Commands = []cli.Command{
		{
			Name: "init",
			Action: func(*cli.Context) error {
				return commands.Init(commands.InitArgs{
					Directory: conf.directory,
				})
			},
		},
		{
			Name: "build",
			Action: func(*cli.Context) error {
				return commands.Build(
					commands.BuildArgs{
						Directory: conf.directory,
					},
					usegoldmark.NewEngine(usegoldmark.EngineConf{
						HighlightStyle: conf.highlightStyle,
					}),
				)
			},
		},
		{
			Name: "serve",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "H,host",
					Value:       "127.0.0.1",
					Destination: &conf.serve.host,
				},
				cli.IntFlag{
					Name:        "p,port",
					Value:       4000,
					Destination: &conf.serve.port,
				},
				cli.BoolFlag{
					Name:        "w,watch",
					Destination: &conf.serve.watch,
				},
			},
			Action: func(c *cli.Context) error {
				return commands.Serve(
					commands.ServeArgs{
						Directory: conf.directory,
						Host:      conf.serve.host,
						Port:      conf.serve.port,
						Watch:     conf.serve.watch,
					},
					usegoldmark.NewEngine(usegoldmark.EngineConf{
						HighlightStyle: conf.highlightStyle,
					}),
				)
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
