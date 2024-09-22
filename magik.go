package main

import (
	"fmt"
	"github.com/hahnicity/go-wget"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func magik(imgPath string, url string) (magikPath string) {

	wget.Wget(url, imgPath)

	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))
	tempFile := inputfile + "temp" + filepath.Ext(imgPath)

	magikPath = inputfile + "magik" + filepath.Ext(imgPath)

	upscale := exec.Command("magick", imgPath, "-liquid-rescale", "200%", tempFile)
	err := upscale.Run()
	if err != nil {
		fmt.Println(err)
	}
	os.Remove(imgPath)

	downscale := exec.Command("magick", tempFile, "-liquid-rescale", "50%", magikPath)
	err = downscale.Run()
	if err != nil {
		fmt.Println(err)
	}
	os.Remove(tempFile)
	return magikPath

}
