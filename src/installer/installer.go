package installer

import (
	"encoding/json"
	"errors"
	"github.com/grifpkg/cli/api"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type grifRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []downloadableGrifRelease `json:"assets"`
}

type downloadableGrifRelease struct {
	BrowserDownloadUrl string `json:"browser_download_url"`
	Name               string `json:"name"`
}

func Install() (release grifRelease, err error) {
	if runtime.GOOS == "windows" {
		release, err := InstallWindows()
		if err != nil {
			return grifRelease{}, err
		}
		return release, err
	} else {
		release, err := InstallUnix()
		if err != nil {
			return grifRelease{}, err
		}
		return release, err
	}
}

func getLatest(path string) (release grifRelease, err error){
	res, err := http.Get("https://api.github.com/repos/grifpkg/cli/releases/latest")
	if err!=nil {
		return grifRelease{}, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err!=nil {
		return grifRelease{}, err
	}
	err = json.Unmarshal(body, &release)
	if err!=nil {
		return grifRelease{}, err
	}
	api.LogOne(api.Progress,"finding compatible binary")
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
		return grifRelease{}, errors.New("couldn't get any binary supporting this OS and/or CPU architecture")
	} else {
		api.LogOne(api.Progress,"downloading grif")
		err := os.MkdirAll(path, 0777)
		if err!= nil && err != os.ErrExist {
			return grifRelease{}, err
		}
		resp, err := http.Get(foundURL)
		if err != nil {
			return grifRelease{}, err
		}
		defer resp.Body.Close()
		var fileName = "grif"
		if runtime.GOOS == "windows" {
			fileName+=".exe"
		}
		api.LogOne(api.Progress,"saving binary")
		out, err := os.Create(path+fileName)
		if err != nil {
			return grifRelease{}, err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		return release, err
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