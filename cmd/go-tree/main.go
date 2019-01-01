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
		cli.IntFlag{
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
