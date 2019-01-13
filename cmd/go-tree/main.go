package main

import (
	"github.com/urfave/cli"
	"go-tree"
	"log"
	"os"
)

func makeApp() *cli.App {
	app := cli.NewApp()

	app.Name = "go-tree"
	app.Usage = "Re-implemented tree command"
	app.Version = "0.1.1"
	app.Action = go_tree.TreeCommand
	app.Flags = []cli.Flag {
		cli.BoolFlag{
			Name: "a",
			Usage: "All files are listed.",
		},
		cli.BoolFlag{
			Name: "d",
			Usage: "List directories only.",
		},
		cli.BoolFlag{
			Name: "l",
			Usage: "Follow symbolic links like directories.",
		},
		cli.StringFlag{
			Name: "L",
			Usage: "Descend only level directories deep.",
		},
	}
	return app
}

func main() {
	app := makeApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
