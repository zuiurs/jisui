package comic

import (
	"flag"
	"fmt"
	"github.com/zuiurs/jisui/subcommand"
)

var CmdComic = &subcommand.Command{
	Name: "comic",
	Run:  Run,
}

var (
	comicV bool
)

func Run(args []string) error {
	f := flag.NewFlagSet("comic", flag.ExitOnError)
	f.BoolVar(&comicV, "v", false, "verbose output")
	f.Parse(args)

	fmt.Println("args:", f.Args())
	fmt.Println("comicV:", comicV)
	return nil
}
