//+build windows

package installer

import (
	"fmt"
	"github.com/kardianos/osext"
	"github.com/segmentio/ksuid"
	"io"
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
	out, err := os.Create(dst+"grif.exe")
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