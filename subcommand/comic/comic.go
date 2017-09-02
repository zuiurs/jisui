package comic

import (
	"C"
	"flag"
	"fmt"
	"github.com/zuiurs/jisui/subcommand"
	"gopkg.in/gographics/imagick.v2/imagick"
	"os"
	"path/filepath"
)

var CmdComic = &subcommand.Command{
	Name: "comic",
	Run:  Run,
}

var (
	comicV         bool
	comicO         string
	comicOverwrite bool
)

func Run(args []string) error {
	f := flag.NewFlagSet("comic", flag.ExitOnError)
	f.BoolVar(&comicV, "v", false, "verbose output")
	f.BoolVar(&comicOverwrite, "overwrite", false, "overwrite the file or directory")
	f.StringVar(&comicO, "o", "-", "output directory")
	f.Parse(args)

	if len(f.Args()) == 0 {
		f.Usage()
		return nil
	}

	if comicOverwrite && (comicO != "-") {
		return fmt.Errorf(`option "o" and "overwrite" is exclusive`)
	}

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()

	for _, v := range f.Args() {
		path, err := filepath.Abs(v)
		if err != nil {
			return err
		}

		dirInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if dirInfo.IsDir() {
			fmt.Println("directory processing")
		} else {
			if err = mw.ReadImage(path); err != nil {
				return err
			}
			if err = mw.SeparateImageChannel(imagick.CHANNEL_RED); err != nil {
				return err
			}
			if err = mw.WriteImageFile(os.Stdout); err != nil {
				return err
			}
		}
	}
	//setimagechannelmask

	return nil
}
