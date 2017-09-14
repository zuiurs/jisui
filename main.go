package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zuiurs/jisui/subcommand"
	"github.com/zuiurs/jisui/subcommand/comic"
	"github.com/zuiurs/jisui/subcommand/prepare"
)

var usage = `jisui is a tool for processing scanned books.

Usage:

	jisui command [arguments]

The commands are:

	prepare	rename as consecutive number and remove unnecessary files
	comic	convert image files for a comic
`

func init() {
	subcommand.Commands = []*subcommand.Command{
		prepare.CmdPrepare,
		comic.CmdComic,
	}
}

func main() {
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, cmd := range subcommand.Commands {
		if args[0] == cmd.Name {
			err := cmd.Run(args[1:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}

	flag.Usage()
	os.Exit(1)
}
