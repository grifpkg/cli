package installer

import (
	"errors"
	"github.com/grifpkg/cli/api"
	"github.com/kardianos/osext"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
)

func isRoot() (bool, error) {
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	return currentUser.Username == "root", nil
}

func InstallUnix() (release grifRelease, err error){
	api.LogOne(api.Progress,"checking root permissions")
	root, _:= isRoot()
	if !root {
		return grifRelease{}, errors.New("you must execute this command as root in order to be able to copy the binary to /usr/local/bin/grifpkg/")
	}
	api.LogOne(api.Progress,"getting working directory")
	file, err := osext.Executable()
	if err != nil {
		return grifRelease{}, err
	}
	randomId := ksuid.New().String()
	installPath := "/opt/grifpkg/bin/"
	api.LogOne(api.Progress,"downloading latest")
	release, err = getLatest(installPath + randomId + "/")
	if err!=nil {
		return grifRelease{}, err
	}

	// remove all folders except new installation and current installation
	api.LogOne(api.Progress,"removing old binaries")
	files, err := ioutil.ReadDir(installPath)
	if err != nil {
		return grifRelease{}, err
	}
	for _, f := range files {
		if f.Name()!="." && path.Base(f.Name()) != path.Base(file) && path.Base(f.Name()) != randomId {
			_ = os.RemoveAll(installPath+f.Name())
		}
	}

	// symlink
	rmErr := os.Remove("/usr/local/bin/grif")
	if !os.IsNotExist(rmErr){
		err = rmErr
	}
	if err!=nil {
		return grifRelease{}, err
	}
	api.LogOne(api.Progress,"creating symlinks and adjusting permissions")
	err = os.Chmod(installPath+randomId+"/grif", 0755)
	if err!=nil {
		return grifRelease{}, err
	}
	err = exec.Command("ln", "-s", installPath+randomId+"/grif", "/usr/local/bin/grif").Run()
	if err!=nil {
		return grifRelease{}, err
	}
	err = os.Chmod("/usr/local/bin/grif", 0755)
	if err!=nil {
		return grifRelease{}, err
	}
	return release ,nil
}