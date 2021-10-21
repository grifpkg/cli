package installer

import (
	"github.com/grifpkg/cli/api"
	"github.com/kardianos/osext"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func InstallWindows() (release grifRelease, err error){
	// dest
	api.LogOne(api.Progress,"getting current working directory and binary")
	file, _ := osext.Executable()
	api.LogOne(api.Progress,"getting config dir")
	targetPath, _ := os.UserConfigDir()
	installPath := targetPath+"\\grifpkg\\bin\\"
	randomId := ksuid.New().String()[0:3]
	api.LogOne(api.Progress,"downloading latest release")
	release, err = getLatest(installPath+randomId+"/")
	if err != nil {
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

	api.LogOne(api.Progress,"creating path script")
	err = createInstallScript(installPath,randomId)
	if err != nil {
		return grifRelease{}, err
	}

	api.LogOne(api.Progress,"running path script")
	err = exec.Command(installPath+"install.bat").Run()

	return release, nil
}

func createInstallScript(installPath string, id string)(err error){
	installScript, err := os.Create(installPath+"install.bat")
	if os.IsExist(err) {
		err = os.Remove(installPath+"install.bat")
	}

	// path value gen
	envPath := os.Getenv("PATH")
	programs := strings.Split(envPath,";")
	finalPrograms := installPath+id+";"
	for _, program := range programs {
		if !strings.HasSuffix(program,installPath) {
			finalPrograms+=strings.TrimSuffix(program,";")+";"
		}
	}

	// path value set
	_, err = installScript.WriteString("setx path \""+finalPrograms+"\"")
	defer installScript.Close()
	return nil
}