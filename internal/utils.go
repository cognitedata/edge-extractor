package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error("Error : ", err)
	}

}

func GetBinaryDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\Cognite\\EdgeExtractor"
	}
	filename := os.Args[1] // get command line first parameter
	return filepath.Dir(filename)
}
