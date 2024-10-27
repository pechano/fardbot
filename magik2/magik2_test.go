package magik2

import (
	"fardbot/magik"
	"fmt"
	"testing"
	"time"
)

func TestFlip(t *testing.T) {
	testmsg := generateDiscordMsg()

	flip(&testmsg)
}

func TestMagik(t *testing.T) {
	testmsg := generateDiscordMsg()

	startOld := time.Now()
	magik.Magick(&testmsg)
	finishOld := time.Since(startOld)

	startNew := time.Now()
	magikResize(&testmsg)
	finishNew := time.Since(startNew)

	fmt.Println("it took " + finishOld.String() + " to finish the old function")
	fmt.Println("it took " + finishNew.String() + " to finish the new function")
}
