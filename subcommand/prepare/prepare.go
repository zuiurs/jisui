package prepare

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/zuiurs/jisui/subcommand"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var CmdPrepare = &subcommand.Command{
	Name: "prepare",
	Run:  Run,
}

var (
	prepareV bool
	prepareD bool
	prepareE string
)

func Run(args []string) error {
	f := flag.NewFlagSet("prepare", flag.ExitOnError)
	f.StringVar(&prepareE, "e", "jpg", "set target file `extension`")
	f.BoolVar(&prepareV, "v", false, "verbose output")
	f.BoolVar(&prepareD, "d", false, "dry-run (not execute rename, but execute clean depending on your operation)")
	f.Parse(args)

	if len(f.Args()) == 0 {
		f.Usage()
		return nil
	}

	// TODO: prepare test
	// MAR_1_001.jpg -> ok
	// MAR_001.jpg -> ok
	// MAR_001_hoge.jpg -> ng
	// MAR_01.jpg -> ng
	// MAR_001.foo -> ng
	re, err := regexp.Compile(`(.*)_[0-9]{3}\.` + prepareE)
	if err != nil {
		return err
	}

	// verify arguments for atomicity
	// rename 処理とまとめると引数の一部のディレクトリだけ処理されなかったりする (まあそれでも良いが...)
	// - is exits?
	// - is directory?
	// - has proper image files?
	for _, v := range f.Args() {
		path, err := filepath.Abs(v)
		if err != nil {
			return err
		}

		dirInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if !dirInfo.IsDir() {
			return fmt.Errorf("%s is not directory", path)
		}

		if err = cleanDirectory(path, re); err != nil {
			return err
		}
	}

	for _, v := range f.Args() {
		path, _ := filepath.Abs(v)
		fis, _ := ioutil.ReadDir(path)

		var prefix string
		for i, fi := range fis {
			if i == 0 {
				prefix = re.FindStringSubmatch(fi.Name())[1]
			}

			src := fmt.Sprintf("%s/%s", path, fi.Name())
			dest := fmt.Sprintf("%s/%s_%03d.%s", path, prefix, i+1, prepareE)
			if !prepareD && src != dest {
				os.Rename(src, dest)
			}
			if (prepareD || prepareV) && src != dest {
				fmt.Printf("%s -> %s\n", src, dest)
			}
		}
	}

	return nil
}

// path 以下の re にマッチしないファイルを削除する
func cleanDirectory(path string, re *regexp.Regexp) error {
	fis, err := ioutil.ReadDir(path) // fis is sorted by name already
	if err != nil {
		return err
	}

	for _, fi := range fis {
		isMatch := re.MatchString(fi.Name())
		if !isMatch {
			target := path + "/" + fi.Name()
			fmt.Printf("Can I delete %s? [y/N]: ", target)
			s := bufio.NewScanner(os.Stdin)
			if s.Scan() {
				if s.Text() == "y" {
					if err = os.Remove(target); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("%s has non proper file: %s", path, fi.Name())
				}
			}
			if err := s.Err(); err != nil {
				return err
			}
		}
	}

	return nil
}
