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

	skipMap map[int]bool
)

func Run(args []string) error {
	f := flag.NewFlagSet("comic", flag.ExitOnError)
	f.BoolVar(&comicV, "v", false, "verbose output")
	f.StringVar(&comicO, "o", "", "set output destination")
	f.StringVar(&comicE, "e", "png", "set output file extension")
	f.IntVar(&comicH, "h", -1, "set image height (iPad Air2: 2048x1536, iPad Pro: 2732x2048)")
	f.Float64Var(&comicBP, "bp", BP, "set leveling black point")
	f.Float64Var(&comicWP, "wp", WP, "set leveling white point")
	f.BoolVar(&comicPack, "pack", false, "packing images to PDF")
	f.BoolVar(&comicOverwrite, "overwrite", false, "overwrite the file or directory")
	f.StringVar(&comicSkip, "skip", "", "set one-based number of image skipping monochrome process (for color page)")
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

	// parse skip list
	var err error
	skipMap, err = parseSkipList(comicSkip)
	if err != nil {
		return err
	}
	if comicV {
		fmt.Printf("Skip Image Index: %v\n", skipMap)
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

// "1,3-5,6,8-9"
// -> 1, 3, 4 , 5, 6, 8, 9
func parseSkipList(str string) (map[int]bool, error) {
	result := make(map[int]bool)
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
				result[i] = true
			}
		} else {
			idx, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			result[idx] = true
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

	dstPath := getDestImagePath(srcPath, comicO, comicE, false)
	// write file
	if dstPath == "-" {
		if err = mw.WriteImageFile(os.Stdout); err != nil {
			return err
		}
	} else {
		if err = mw.WriteImage(dstPath); err != nil {
			return err
		}
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
		mw.SetImageFormat("pdf")

		for i, fi := range fis {
			if comicV {
				fmt.Printf("%s processing...\n", fi.Name())
			}

			if err = mw.ReadImage(srcPath + "/" + fi.Name()); err != nil {
				return err
			}
			// if skipMap has the key of the index, this image does not be monochrome.
			_, isSkip := skipMap[i+1]
			if err = processImage(mw, isSkip); err != nil {
				return err
			}

			mw.SetLastIterator()
		}

		dstPath := getDestImagePath(srcPath, comicO, "pdf", false)
		if _, err := os.Stat(dstPath); err != nil {
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

			_, isSkip := skipMap[i+1]
			if err = processImage(mw, isSkip); err != nil {
				return err
			}

			if _, err := os.Stat(comicO); err != nil {
				if err = os.Mkdir(comicO, 0755); err != nil {
					return err
				}
			}
			dstPath := getDestImagePath(srcPath+"/"+fi.Name(), comicO, comicE, true)
			// write file
			if dstPath == "-" {
				if err = mw.WriteImageFile(os.Stdout); err != nil {
					return err
				}
			} else {
				if err = mw.WriteImage(dstPath); err != nil {
					return err
				}
			}

			mw.Destroy()
		}
	}

	return nil
}

// srcPath is image source path.
// output is symbol to be based to determinate destination path.
// ext is file name extension.
// if output is directory, then isDir is true, else false.
func getDestImagePath(srcPath, output, ext string, isDir bool) string {
	var dstPath string

	if output == "" {
		pos := strings.LastIndex(srcPath, ".")
		if pos == -1 {
			dstPath = srcPath + "." + ext
		} else {
			dstPath = srcPath[:pos+1] + ext
		}
	} else {
		if isDir {
			s := strings.Split(srcPath, "/")
			filename := s[len(s)-1]

			pos := strings.LastIndex(filename, ".")
			if pos == -1 {
				filename = filename + "." + ext
			} else {
				filename = filename[:pos+1] + ext
			}

			dstPath = output + "/" + filename
		} else {
			dstPath = output
		}
	}

	return dstPath
}

func processImage(mw *imagick.MagickWand, monoSkip bool) error {
	var err error
	if !monoSkip {
		if err = monochromeImage(mw, comicBP, comicWP); err != nil {
			return err
		}
	}
	if err = resizeImage(mw, comicH, comicE); err != nil {
		return err
	}

	return nil

}

// resizeImage does below:
// - resize the image
// - change file format
func resizeImage(mw *imagick.MagickWand, height int, format string) error {
	var err error
	if err = mw.SetImageFormat(format); err != nil {
		return err
	}

	var h, w uint
	if height < 0 {
		return nil
	}
	h = uint(height)
	w = uint((float64(h) / float64(mw.GetImageHeight())) * float64(mw.GetImageWidth()))

	if comicV {
		fmt.Printf("Image size: %d x %d(SRC) -> %d x %d(DST)\n", mw.GetImageWidth(), mw.GetImageHeight(), w, h)
	}

	if err = mw.ResizeImage(w, h, imagick.FILTER_MITCHELL, 1); err != nil { // 0 is sharp
		return err
	}

	return nil

}

// monochromeImage does below:
// - monochrome image
//   - grayscale by Red Channel because yellow tint has few red element
// - compensate color level for mitigating show-through and sharpness
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
