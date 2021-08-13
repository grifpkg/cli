//+build windows

package installer

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/grifpkg/cli/globals"
	"github.com/kardianos/osext"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

func Install(){
	// run as admin
	// dest
	file, _ := osext.Executable()
	targetPath, _ := os.UserConfigDir()
	installPath := targetPath+"\\grifpkg\\bin\\"
	randomId := ksuid.New().String()
	copyerr := copyBinary(file,installPath+randomId+"\\")
	if copyerr!=nil {
		fmt.Println(copyerr.Error())
	}
	// remove all folders except new installation and current installation
	files, err := ioutil.ReadDir(installPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.Name()!="." && path.Base(f.Name()) != path.Base(file) && path.Base(f.Name()) != randomId {
			_ = os.RemoveAll(installPath+f.Name())
		}
	}

	createInstallScript(installPath, randomId)

	exec.Command(installPath+"install.bat").Run()
	fmt.Fprintf(color.Output, "%s grif %s has been installed\n", color.HiGreenString("i"), color.CyanString(globals.Version))
}

func createInstallScript(installPath string, id string){
	installScript, err := os.Create(installPath+"install.bat")
	if os.IsExist(err) {
		err = os.Remove(installPath+"install.bat")
	} else {
		// error
	}
	_, err = installScript.WriteString("setx path \""+installPath+id+"\\;%PATH%\"")
	defer installScript.Close()
}