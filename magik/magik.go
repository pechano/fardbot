package magik

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/hahnicity/go-wget"
)

func Magick(m *discordgo.MessageCreate) (magikPath string) {
	imgPath := m.ReferencedMessage.Attachments[0].Filename
	url := m.ReferencedMessage.Attachments[0].ProxyURL

	fmt.Println("applying MAGIk to " + imgPath)
	wget.Wget(url, imgPath)

	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))
	tempFile := inputfile + "temp" + filepath.Ext(imgPath)

	magikPath = inputfile + "magikold" + filepath.Ext(imgPath)

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

func FlipImg(m *discordgo.MessageCreate) (flipPath string) {

	imgPath := m.ReferencedMessage.Attachments[0].Filename
	url := m.ReferencedMessage.Attachments[0].ProxyURL
	defer os.Remove(imgPath)
	fmt.Println("applying FLIP to " + imgPath)
	err := downloadFile(imgPath, url)
	if err != nil {
		fmt.Println(err)
		fmt.Println("downlaod command failed")
	}

	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))

	flipPath = inputfile + "flip" + filepath.Ext(imgPath)

	flip := exec.Command("magick", imgPath, "-flip", flipPath)
	err = flip.Run()
	if err != nil {
		fmt.Println(err)
		fmt.Println("flip command failed")
	}
	return flipPath

}

func Deepfry(m *discordgo.MessageCreate) (fryPath string) {

	imgPath := m.ReferencedMessage.Attachments[0].Filename
	url := m.ReferencedMessage.Attachments[0].ProxyURL
	defer os.Remove(imgPath)
	fmt.Println("deepfryin' " + imgPath)
	err := downloadFile(imgPath, url)
	if err != nil {
		fmt.Println(err)
		fmt.Println("downlaod command failed")
	}

	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))

	fryPath = inputfile + "fry" + ".jpg"
	deepfryer := []string{
		imgPath,
		"-modulate",
		"80,200,90",
		"-sharpen",
		"0x10",
		"-sigmoidal-contrast",
		"5,0%",
		"-fill",
		"orange4",
		"-colorize",
		"50%",
		"-quality",
		"5",
		fryPath}

	fry := exec.Command("magick", deepfryer...)

	err = fry.Run()
	if err != nil {
		fmt.Println(err)
		fmt.Println("flip command failed")
	}
	return fryPath
}
func Sponge(m *discordgo.MessageCreate) (spongePath string) {

	imgPath := "spongebase.jpg"
	mockString := m.ReferencedMessage.Content
	mockString = spongify(mockString)
	fmt.Println(mockString)

	fmt.Println("creating best argument")

	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))

	mockPath := inputfile + "mock" + fmt.Sprint(rand.Int()) + ".jpg"
	spongemock := []string{
		imgPath,
		"-size",
		"200x50",
		"-font",
		"Impact",
		"-pointsize",
		"160",
		"-fill",
		"white",
		"-gravity",
		"south",
		"-stroke",
		"black",
		"-strokewidth",
		"5",
		"-annotate",
		"+25+70",
		mockString,
		"-trim",
		"+repage",
		mockPath}

	fry, err := exec.Command("magick", spongemock...).CombinedOutput()
	fmt.Println(string(fry))

	if err != nil {
		fmt.Println(err)
		fmt.Println("sponge command failed")
	}
	return mockPath
}

func spongify(input string) (output string) {
	unicodeString := []rune(input)

	for i := 1; i <= len(unicodeString)-1; i++ {

		if i%2 == 0 {
			unicodeString[i] = unicode.ToUpper(unicodeString[i])
		}
	}
	var mixedCaseWithNewlines []rune
	newline := []rune("\n")

	finalIndex := 0
	if len(unicodeString) < 40 {
		output = string(unicodeString)
		return output
	}

	if len(unicodeString) > 40 {
		for i := 1; i <= len(unicodeString)-1; i++ {
			if i%40 == 0 {

				mixedCaseWithNewlines = append(mixedCaseWithNewlines, unicodeString[i-40:i]...)

				mixedCaseWithNewlines = append(mixedCaseWithNewlines, newline...)
				finalIndex = i
			}
		}
		mixedCaseWithNewlines = append(mixedCaseWithNewlines, unicodeString[finalIndex:]...)
	}
	fmt.Println(finalIndex)
	output = string(mixedCaseWithNewlines)
	return output

}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
