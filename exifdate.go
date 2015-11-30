package main

import (
	"errors"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func herr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var (
	dateDirRegex *regexp.Regexp = regexp.MustCompile("^[0-9]{2}")
	jpgRegex     *regexp.Regexp = regexp.MustCompile("(?i)^.*\\.jp[e]?g$")
)

/*<F10>
func init() {
	dateDirRegex = regexp.MustCompile("^[0-9-]{2}")
	jpgRegex = regexp.MustCompile("(?i)^.*\\.jp[e]?g$")
}
*/

func fullDateInDirname(dirname string) (time.Time, error) {
	dateSplit := strings.Split(dirname, "_")

	if len(dateSplit[0]) != 10 {
		return time.Time{}, errors.New("Format in dir name  is not 2006-01-02_*")
	}
	return time.Parse("2006-01-02", dateSplit[0])
}

func fullDateInDirlist(pathlist []string, i int) (time.Time, error) {
	concat := ""

	for j := i; j > i-3; j-- {
		concat = concat + pathlist[j] + "-"
	}
	concat = concat[:len(concat)-1]

	return time.Parse("02-01-2006", concat)
}

func getDateFromPath(path string) (time.Time, error) {
	dirname, _ := filepath.Split(path)
	pathlist := strings.Split(dirname, string(os.PathSeparator))
	pathlist = pathlist[1 : len(pathlist)-1]
	i := len(pathlist) - 1
	if !dateDirRegex.MatchString(pathlist[i]) {
		for ; i >= 0; i-- {
			if dateDirRegex.MatchString(pathlist[i]) {
				break
			}
		}
	}

	fd, err := fullDateInDirname(pathlist[i])
	if err == nil {
		return fd, nil
	}

	fd, err = fullDateInDirlist(pathlist, i)
	if err == nil {
		return fd, nil
	}
	return time.Time{}, err
}

func visitFiles(path string, fi os.FileInfo, err error) error {
	if err != nil {
		log.Println("Error in walking ", err)
		return err
	}
	if fi.IsDir() {
		return nil
	}
	if !jpgRegex.MatchString(path) {
		log.Println("Not a jpg file")
		return nil
	}

	dirDate, err := getDateFromPath(path)
	if err != nil {
		log.Println("Couldn't get date from dir tree", path)
		return nil
	}

	log.Println("FSTREE:", path, dirDate)

	f, err := os.Open(path)

	if err != nil {
		log.Println("Error opening file during walk ", err)
		return err
	}

	x, err := exif.Decode(f)
	if err != nil {
		log.Println("EXIF: error in decoding", err)
		return err
	}

	xDate, err := x.DateTime()
	if err != nil {
		log.Println("EXIF: error in getting time", err)
		return err
	}
	log.Println("EXIF:", path, xDate)
	return nil
}

func main() {
	fname := os.Args[1]

	filepath.Walk(fname, visitFiles)
	/*
		f, err := os.Open(fname)
		herr(err)

		x, err := exif.Decode(f)
		herr(err)

		tm, err := x.DateTime()
		fmt.Println("Taken: ", tm)
	*/
}
