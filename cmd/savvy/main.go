package main

import (
	"context"
	"log"
	"os"

	"github.com/heyajulia/savvy/internal"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "savvy",
		Usage:   "Energy price bot for Telegram and Bluesky",
		Version: internal.Version,
		Commands: []*cli.Command{
			serveCommand(),
			reportCommand(),
			upgradeCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
