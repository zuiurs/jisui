package comic

import (
	"C"
	"flag"
	"fmt"
	"github.com/zuiurs/jisui/subcommand"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	BP = 0.40
	WP = 0.85
)

var CmdComic = &subcommand.Command{
	Name: "comic",
	Run:  Run,
}

var (
	comicV         bool
	comicO         string
	comicE         string
	comicH         int
	comicBP        float64
	comicWP        float64
	comicPack      bool
	comicOverwrite bool
)

func Run(args []string) error {
	f := flag.NewFlagSet("comic", flag.ExitOnError)
	f.BoolVar(&comicV, "v", false, "verbose output")
	f.StringVar(&comicO, "o", "", "set output destination")
	f.StringVar(&comicE, "e", "png", "set output file extension")
	f.IntVar(&comicH, "h", -1, "set image height (iPad Air2: 2048x1536)")
	f.Float64Var(&comicBP, "bp", BP, "set leveling black point")
	f.Float64Var(&comicWP, "wp", WP, "set leveling white point")
	f.BoolVar(&comicPack, "pack", false, "packing images to PDF")
	f.BoolVar(&comicOverwrite, "overwrite", false, "overwrite the file or directory")
	f.Parse(args)

	if len(f.Args()) == 0 {
		f.Usage()
		return nil
	}

	if comicOverwrite && (comicO != "") {
		return fmt.Errorf(`option "o" and "overwrite" is exclusive`)
	}
	if comicOverwrite && comicPack {
		return fmt.Errorf(`option "pack" and "overwrite" is exclusive`)
	}
	comicO = strings.TrimRight(comicO, "/")

	imagick.Initialize()
	defer imagick.Terminate()

	for _, v := range f.Args() {
		path, err := filepath.Abs(v)
		if err != nil {
			return err
		}

		dirInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if dirInfo.IsDir() { // directory process
			if err = processDirectory(path); err != nil {
				return err
			}
		} else { // single file process
			if err = processSingleFile(path); err != nil {
				return err
			}
		}
	}

	return nil
}

func processSingleFile(srcPath string) error {
	var err error

	mw := imagick.NewMagickWand()
	if err = mw.ReadImage(srcPath); err != nil {
		return err
	}

	if err = processImage(mw); err != nil {
		return err
	}

	// write file
	var dstPath string
	if comicO == "" {
		pos := strings.LastIndex(srcPath, ".")
		if pos == -1 {
			dstPath = srcPath + comicE
		} else {
			dstPath = srcPath[:pos+1] + comicE
		}
	} else if comicO == "-" {
		if err = mw.WriteImageFile(os.Stdout); err != nil {
			return err
		}
		return nil
	} else {
		dstPath = comicO
	}
	if err = mw.WriteImage(dstPath); err != nil {
		return err
	}

	return nil
}

func processDirectory(srcPath string) error {
	fis, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	if comicPack {
		mw := imagick.NewMagickWand()
		for _, fi := range fis {
			if err = mw.ReadImage(srcPath + "/" + fi.Name()); err != nil {
				return err
			}

			if err = processImage(mw); err != nil {
				return err
			}

			// add file
			mw.SetLastIterator()
		}

		// write pdf
		mw.SetImageFormat("pdf")
		var dstPath string
		if comicO == "" {
			dstPath = srcPath + ".pdf"
		} else if comicO == "-" {
			if err = mw.WriteImageFile(os.Stdout); err != nil {
				return err
			}
			return nil
		} else {
			dstPath = comicO
		}
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			if err = mw.WriteImages(dstPath, true); err != nil {
				return err
			}
		}
	} else {
		for _, fi := range fis {
			mw := imagick.NewMagickWand()
			if err = mw.ReadImage(srcPath + "/" + fi.Name()); err != nil {
				return err
			}

			if err = processImage(mw); err != nil {
				return err
			}

			// write file
			var dstPath string
			if comicO == "" {
				dstPath = srcPath + "/" + fi.Name()
			} else if comicO == "-" {
				if err = mw.WriteImageFile(os.Stdout); err != nil {
					return err
				}
				continue
			} else {
				if _, err := os.Stat(comicO); os.IsNotExist(err) {
					if err = os.Mkdir(comicO, 0755); err != nil {
						return err
					}
				}
				dstPath = comicO + "/" + fi.Name()
			}
			pos := strings.LastIndex(dstPath, ".")
			if pos == -1 {
				dstPath = dstPath + comicE
			} else {
				dstPath = dstPath[:pos+1] + comicE
			}
			if err = mw.WriteImage(dstPath); err != nil {
				return err
			}

		}
	}

	return nil
}

// - モノクロ化
// - 拡張子の変更
// - イメージのリサイズ
func processImage(mw *imagick.MagickWand) error {
	var err error
	if err = monochrome(mw, comicBP, comicWP); err != nil {
		return err
	}

	if err = mw.SetImageFormat(comicE); err != nil {
		return err
	}

	var h, w uint
	if comicH < 0 {
		h = mw.GetImageHeight()
		w = mw.GetImageWidth()
	} else {
		h = uint(comicH)
		w = uint((float64(h) / float64(mw.GetImageHeight())) * float64(mw.GetImageWidth()))
	}

	if comicV {
		fmt.Printf("SRC Image size: %d x %d\n", mw.GetImageWidth(), mw.GetImageHeight())
		fmt.Printf("DST Image size: %d x %d\n", w, h)
	}

	if err = mw.ResizeImage(w, h, imagick.FILTER_MITCHELL, 1); err != nil { // 0 is sharp
		return err
	}

	return nil
}

// - Red Channel でグレースケール化
// - 色レベル補正 (裏写り軽減、シャープネス化)
func monochrome(mw *imagick.MagickWand, bp, wp float64) error {
	var err error
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
