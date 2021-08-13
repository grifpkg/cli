package installer

import (
	"io"
	"os"
	"runtime"
)

func Install() {
	if runtime.GOOS == "windows" {
		InstallWindows()
	} else {
		InstallUnix()
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