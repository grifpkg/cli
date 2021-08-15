package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func Install() {
	if runtime.GOOS == "windows" {
		InstallWindows()
	} else {
		InstallUnix()
	}
}

type grifRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []downloadableGrifRelease `json:"assets"`
}

type downloadableGrifRelease struct {
	BrowserDownloadUrl string `json:"browser_download_url"`
	Name               string `json:"name"`
}

func getLatest(path string) error{
	res, err := http.Get("https://api.github.com/repos/grifpkg/cli/releases/latest")
	if err!=nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err!=nil {
		return err
	}
	var release grifRelease
	err = json.Unmarshal(body, &release)
	if err!=nil {
		return err
	}
	const is64Bit = uint64(^uintptr(0)) == ^uint64(0)
	var expectedFileName = "grif_"+runtime.GOOS
	expectedFileName = strings.ReplaceAll(expectedFileName,"darwin", "macos")
	if runtime.GOARCH == "arm64" ||runtime.GOARCH == "amd64" {
		expectedFileName+="_x64"
	} else {
		expectedFileName+="_x32"
	}
	if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
		expectedFileName+="_arm"
	}
	if runtime.GOOS == "windows" {
		expectedFileName+=".exe"
	}
	var foundURL = ""
	for _, asset := range release.Assets {
		if asset.Name==expectedFileName{
			foundURL = asset.BrowserDownloadUrl
			break
		}
	}
	if foundURL=="" {
		return errors.New("couldn't get any binary supporting this OS and/or CPU architecture")
	} else {
		fmt.Fprintf(color.Output, "%s grif %s is being downloaded and installed\n", color.HiGreenString("i"), color.CyanString(release.TagName))
		err := os.MkdirAll(path, 0777)
		if err!= nil && err != os.ErrExist {
			return err
		}
		resp, err := http.Get(foundURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var fileName = "grif"
		if runtime.GOOS == "windows" {
			fileName+=".exe"
		}
		out, err := os.Create(path+fileName)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		return err
	}
}

func copyBinary(src string, dst string) error{
	err := os.MkdirAll(dst, 0777)
	if err!= nil && err != os.ErrExist {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	var added = ""
	if runtime.GOOS == "windows" {
		added+=".exe"
	}
	out, err := os.Create(dst+"grif"+added)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return err
}