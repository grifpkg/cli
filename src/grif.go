package main

import (
	"./client"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCMD = &cobra.Command{
	Use:   "grif",
	Short: "grif is a plugin management system for bukkit-based projects",
	Long: "grif is a plugin management system for bukkit-based projects that allows you to install, remove and update packages existing on spigotmc, the grif library, and various other configurable sources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("grif-cli version", Yellow("x.x.x"))
	},
}

func fileExists (filepath string) bool {

	fileInfo, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}
	// Return false if the fileInfo says the file path is a directory.
	return !fileInfo.IsDir()
}

var installCMD = &cobra.Command{
	Use:   "install",
	Aliases: []string{"i"},
	Short: "installs a package",
	Long: "installs a package and its dependencies, and creates a project file if it isn't present yet",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := os.Getwd()
		separator := string(os.PathSeparator)
		if err != nil {
			log.Println(Red("error while creating config file: "), "couldn't find the current working path ('"+err.Error()+"')")
		} else {
			finalPath := path+separator+"grif.json"
			if !fileExists(finalPath) {
				// ignore
			}

			if len(args)>0 {
				dependency := client.ParseResourceString(args[0])
				resources, err := client.GetResources(dependency.Resource, dependency.Author, 0)
				if err != nil {
					return
				}

				fmt.Println("downloading resource")
				_, err = client.DownloadResource(resources[0],"")
				if err != nil {
					return
				}
				fmt.Println("downloaded")
			} else {
				return
			}
		}

	},
}

var initCMD = &cobra.Command{
	Use:   "init",
	Short: "Initializes the grif config file",
	Long: "Creates a basic grif config file in order to start managing plugins",
	Run: func(cmd *cobra.Command, args []string) {

		path, err := os.Getwd()
		separator := string(os.PathSeparator)
		if err != nil {
			log.Println(Red("error while creating config file: "), "couldn't find the current working path ('"+err.Error()+"')")
		} else {
			fmt.Println("created grif config file:", Yellow(path+separator+"grif.json"))
		}

	},
}

func main(){
	if err := rootCMD.Execute(); err != nil {
		fmt.Println(Red("error while initializing grif:"),err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCMD.AddCommand(initCMD)
	rootCMD.AddCommand(installCMD)
}