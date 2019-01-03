package comic

import (
	"C"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zuiurs/jisui/subcommand"
	"gopkg.in/gographics/imagick.v2/imagick"
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
	comicSkip      string

	skipList []int
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
	f.StringVar(&comicSkip, "skip", "", "set number of image skipping monochrome process (for color page)")
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

	// skip パース処理
	var err error
	skipList, err = parseSkipList(comicSkip)
	if err != nil {
		return err
	}
	if comicV {
		fmt.Printf("Skip Image Index: %v\n", skipList)
	}

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

func parseSkipList(str string) ([]int, error) {
	result := make([]int, 0, 10)
	s := strings.Split(str, ",")
	for _, v := range s {
		if strings.ContainsAny(v, "-") {
			r := strings.Split(v, "-")
			fromIdx, err := strconv.Atoi(r[0])
			if err != nil {
				return nil, err
			}
			toIdx, err := strconv.Atoi(r[1])
			if err != nil {
				return nil, err
			}
			for i := fromIdx; i <= toIdx; i++ {
				result = append(result, i)
			}
		} else {
			idx, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			result = append(result, idx)
		}
	}
	return result, nil
}

func processSingleFile(srcPath string) error {
	var err error

	mw := imagick.NewMagickWand()
	if err = mw.ReadImage(srcPath); err != nil {
		return err
	}

	if err = processImage(mw, false); err != nil {
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
		for i, fi := range fis {
			if comicV {
				fmt.Printf("%s processing...\n", fi.Name())
			}

			if err = mw.ReadImage(srcPath + "/" + fi.Name()); err != nil {
				return err
			}

			// skipList に入っていたらモノクロ化しない
			// skipList の値は 1 起算を考慮
			isSkip := false
			for _, v := range skipList {
				if v == i+1 {
					isSkip = true
				}
			}
			if err = processImage(mw, isSkip); err != nil {
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
		for i, fi := range fis {
			if comicV {
				fmt.Printf("%s processing...\n", fi.Name())
			}

			mw := imagick.NewMagickWand()
			if err = mw.ReadImage(srcPath + "/" + fi.Name()); err != nil {
				return err
			}

			isSkip := false
			for _, v := range skipList {
				if v == i+1 {
					isSkip = true
				}
			}
			if err = processImage(mw, isSkip); err != nil {
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

func processImage(mw *imagick.MagickWand, monoSkip bool) error {
	var err error
	if !monoSkip {
		if err = monochromeImage(mw, comicBP, comicWP); err != nil {
			return err
		}
	}
	if err = resizeImage(mw); err != nil {
		return err
	}

	return nil

}

// - イメージのリサイズ
// - 拡張子の変更
func resizeImage(mw *imagick.MagickWand) error {
	var err error
	if err = mw.SetImageFormat(comicE); err != nil {
		return err
	}

	var h, w uint
	if comicH < 0 {
		return nil
	} else {
		h = uint(comicH)
		w = uint((float64(h) / float64(mw.GetImageHeight())) * float64(mw.GetImageWidth()))
	}

	if comicV {
		fmt.Printf("Image size: %d x %d(SRC) -> %d x %d(DST)\n", mw.GetImageWidth(), mw.GetImageHeight(), w, h)
	}

	if err = mw.ResizeImage(w, h, imagick.FILTER_MITCHELL, 1); err != nil { // 0 is sharp
		return err
	}

	return nil

}

// - モノクロ化
// - Red Channel でグレースケール化
// - 色レベル補正 (裏写り軽減、シャープネス化)
func monochromeImage(mw *imagick.MagickWand, bp, wp float64) error {
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
