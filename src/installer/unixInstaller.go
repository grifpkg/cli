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

func InstallUnix(){
	root, _:= isRoot()
	if !root {
		fmt.Fprintf(color.Output, "%s %s\n", color.HiYellowString("!"), color.RedString("You must execute this command as root in order to be able to copy the binary to /usr/local/grifpkg/bin"))
		return
	}
	file, _ := osext.Executable()
	randomId := ksuid.New().String()
	installPath := "/usr/local/etc/grifpkg/bin/"
	err := getLatest(installPath+randomId+"/")
	if err != nil {
		log.Fatal(err)
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

	// symlink
	_ = os.Remove("/usr/local/bin/grif")
	exec.Command("ln", "-s", installPath+randomId+"/grif", "/usr/local/bin/grif").Run()
	exec.Command("chmod", "+x", "/usr/local/bin/grif").Run()

	// install notice
	fmt.Fprintf(color.Output, "%s grif %s has been installed\n", color.HiGreenString("i"), color.CyanString(globals.Version))
}