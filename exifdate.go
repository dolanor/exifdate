package main

import (
	"errors"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"os/exec"
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
	defer f.Close()
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
	sub := xDate.Sub(dirDate)
	if sub <= 2*24*time.Hour {
		return nil
	}
	log.Println("before", dirDate)
	//	y, m, d := dirDate.Date()
	h, mm, s := xDate.Clock()

	dirDate = dirDate.Add(time.Date(0, 0, 0, h, mm, s, 0, time.UTC).Sub(time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)))

	log.Println("after", dirDate)

	//exiv2 -M "set Exif.Image.DateTime 2000:01:01 00:00:17" -M  "set Exif.Photo.DateTimeOriginal 2001:01:01 00:00:17" -M "set Exif.Photo.DateTimeDigitized 2002:01:01 00:00:17" mo $file
	xformatDate := dirDate.Format("2006:01:02 15:04:05")
	commandString := []string{"exiv2", "-M", "set Exif.Image.DateTime " + xformatDate, "-M", "set Exif.Photo.DateTimeOriginal " + xformatDate, "-M", "set Exif.Photo.DateTimeDigitized " + xformatDate, "mo", path}

	_, err = exec.Command(commandString[0], commandString[1:]...).Output()
	if err != nil {
		log.Println("Couldn't modify tags with exiv2 for ", path)
		return nil
	}
	log.Println("Changed timestamp for", path, "to", dirDate)
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
