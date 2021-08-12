//+build linux
//+build freebsd
//+build darwin

package installer

import (
	"fmt"
	"github.com/kardianos/osext"
	"os"
	"os/exec"
	"strings"
)

func Install(){
	path, _ := osext.ExecutableFolder()
	file, _ := osext.Executable()
	fmt.Println(path+file)
	if(!strings.HasPrefix(file, "/usr/local/grifpkg/bin/")){
		_ := os.MkdirAll("/usr/local/grifpkg/bin/", 0700)
		_ := os.Rename(path+file,"/usr/local/grifpkg/bin/grif")
	}
	exec.Command("export","PATH=/usr/local/grifpkg/bin/:$PATH")
}