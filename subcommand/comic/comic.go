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

const (
	BP = 0.33
	WP = 0.9
)

var CmdComic = &subcommand.Command{
	Name: "comic",
	Run:  Run,
}

var (
	comicV         bool
	comicO         string
	comicBP        float64
	comicWP        float64
	comicOverwrite bool
)

func Run(args []string) error {
	f := flag.NewFlagSet("comic", flag.ExitOnError)
	f.BoolVar(&comicV, "v", false, "verbose output")
	f.StringVar(&comicO, "o", "-", "output destination")
	f.Float64Var(&comicBP, "bp", BP, fmt.Sprintf("set leveling black point (default %f)", BP))
	f.Float64Var(&comicWP, "wp", WP, fmt.Sprintf("set leveling white point (default %f)", WP))
	f.BoolVar(&comicOverwrite, "overwrite", false, "overwrite the file or directory")
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
			monochrome(mw, path, comicBP, comicWP)

			// write file
			if comicOverwrite {
				if err = mw.WriteImage(path); err != nil {
					return err
				}
			} else {
				if comicO == "-" {
					if err = mw.WriteImageFile(os.Stdout); err != nil {
						return err
					}
				} else {
					if err = mw.WriteImage(comicO); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func monochrome(mw *imagick.MagickWand, path string, bp, wp float64) error {
	var err error
	if err = mw.ReadImage(path); err != nil {
		return err
	}

	// remove yellow tint
	if err = mw.SeparateImageChannel(imagick.CHANNEL_RED); err != nil {
		return err
	}

	// monochrome
	_, r := imagick.GetQuantumRange()
	if err = mw.LevelImage(float64(r)*bp, 1.0, float64(r)*wp); err != nil {
		return err
	}

	return nil
}
