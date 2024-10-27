package magik2

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func flip(m *discordgo.MessageCreate) (flipPath string) {
	//boilerplate needed for decode the message and download the image file

	imgPath := m.ReferencedMessage.Attachments[0].Filename
	url := m.ReferencedMessage.Attachments[0].ProxyURL
	defer os.Remove(imgPath)
	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))
	flipPath = inputfile + "flip" + filepath.Ext(imgPath)

	fmt.Println("applying FLIP to " + imgPath)
	err := downloadFile(imgPath, url)
	//imageMagick initiation block
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err = mw.ReadImage(imgPath)
	if err != nil {
		fmt.Println("failed to open image at specified path")
		return
	}
	//here the image operations can take place

	mw.FlipImage()

	//the file is written, the path is returned and the function is done
	mw.WriteImage(flipPath)
	return flipPath
}
func magikResize(m *discordgo.MessageCreate) (magikPath string) {
	//boilerplate needed for decode the message and download the image file

	imgPath := m.ReferencedMessage.Attachments[0].Filename
	url := m.ReferencedMessage.Attachments[0].ProxyURL
	defer os.Remove(imgPath)
	inputfile := strings.TrimSuffix(imgPath, filepath.Ext(imgPath))
	magikPath = inputfile + "magik" + filepath.Ext(imgPath)

	fmt.Println("applying some magik to " + imgPath)
	err := downloadFile(imgPath, url)
	//imageMagick initiation block
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err = mw.ReadImage(imgPath)
	if err != nil {
		fmt.Println("failed to open image at specified path")
		return
	}
	//here the image operations can take place

	x := mw.GetImageWidth()

	y := mw.GetImageHeight()

	mw.ResizeImage(x*2, y*2, imagick.FILTER_BOX)
	mw.LiquidRescaleImage(uint(x), uint(y), 1, 0)

	//the file is written, the path is returned and the function is done
	mw.WriteImage(magikPath)
	return magikPath
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

func generateDiscordMsg() discordgo.MessageCreate {

	var refmsg discordgo.Message
	var refAtt discordgo.MessageAttachment
	testmsg := discordgo.MessageCreate{Message: &refmsg}
	testmsg.ReferencedMessage = testmsg.Message
	refAtt.Filename = "spongebase.jpg"
	refAtt.ProxyURL = "https://archive.org/download/spongebob_img/03spongebob_xp-articleLarge.jpg"
	refmsg.Attachments = append(refmsg.Attachments, &refAtt)
	return testmsg

}
