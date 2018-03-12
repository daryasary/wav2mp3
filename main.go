package main

import (
	"path/filepath"
	"os"
	"fmt"
	"regexp"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"flag"
)

var wg sync.WaitGroup

func ConvertToMP3(filenames ...string) error {
	var toFilename string
	var fromFilename string = filenames[0]
	switch len(filenames) {
	case 1:
		toFilename = filenames[0]
		break
	case 0:
		return errors.New("error: no arguements are passed")
	default:
		toFilename = filenames[1]
	}
	// Convert to MP3
	comm := exec.Command("ffmpeg", "-i", fromFilename, "-vn", "-y", "-ab", "192k", "-f", "mp3", toFilename+".mp3")
	if err := comm.Run(); err != nil {
		return err
	}
	return nil
}

func TouchEmptyWav(filename string)error{
	comm := exec.Command("truncate", "-s", "0", filename)
	if err := comm.Run(); err != nil {
		return err
	}
	return nil
}

func main(){
	soundsDir := flag.String("dir", "", "Parent directory of wav files")
	workersCount := flag.Int("worker", 3, "Number of go concurrent workers")
	flag.Parse()

	if *soundsDir == ""{
		panic("No directory entered")
	}
	fileList := make(chan string)

	var found int

	for i:=1; i<=*workersCount; i++{
		wg.Add(1)
		go convertWorker(i, fileList)
	}

	filepath.Walk(*soundsDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir(){
			r, err := regexp.MatchString(".wav", f.Name())
			if err == nil && r {
				info, err := os.Stat(path)
				if err == nil && info.Size() > 0{
					found ++
					fileList <- path
				}
			}
		}
		return nil
	})
	close(fileList)

	fmt.Printf("%d wav files found in working directory\n", found)
	wg.Wait()
	fmt.Println("Convertion Done")
}

func convertWorker(id int, fileList <-chan string){
	defer wg.Done()
	for file := range fileList {
		fmt.Println("Worker #", id, "grabbed", file)
		//Extract file name
		splits := strings.Split(file, ".wav")
		err := ConvertToMP3(file, splits[0])
		if err != nil {
			fmt.Println("Convert error", file, err)
		} else {
			if err := TouchEmptyWav(file); err != nil {
				fmt.Printf("Convert %s completed, Truncate encountered error :%s\n", file, err)
			}
			fmt.Printf("Convert %s completed, Truncate completed\n", file)
		}
	}
}