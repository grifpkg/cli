package main

import (
	"./client"
	"./config"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var rootCMD = &cobra.Command{
	Use:   "grif",
	Short: "grif is a plugin management system for bukkit-based projects",
	Long: "grif is a plugin management system for bukkit-based projects that allows you to install, remove and update packages existing on spigotmc, the grif library, and various other configurable sources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(color.Output, "%s grif-cli version %s\n", color.HiGreenString("i"), color.CyanString("1.0.0"))
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
		if len(args)>0 {
			var err error = nil
			project, err :=config.Load()
			if err!=nil {
				fmt.Fprintf(color.Output, "%s Error while loading project file: %s\n", color.HiYellowString("!"), color.RedString(err.Error()))
				return
			}
			dependency := config.ParseResourceString(args[0])
			fmt.Fprintf(color.Output, "%s Querying resource %s\n", color.HiGreenString("i"), color.CyanString(dependency.Resource))
			resources, err := client.GetResources(dependency.Resource, dependency.Author, 0)
			if len(resources)<=0 || err!=nil {
				fmt.Fprintf(color.Output, "%s No resources were found\n", color.HiYellowString("!"))
				return
			}
			fmt.Fprintf(color.Output, "%s Downloading resource %s\n", color.HiGreenString("i"), color.CyanString(resources[0].Name))
			release, err := client.DownloadResource(resources[0],"", project)
			if err!=nil {
				fmt.Fprintf(color.Output, "%s Error while downloading %s: %s\n", color.HiYellowString("!"), color.CyanString(resources[0].Name), color.RedString(err.Error()))
				return
			}
			fmt.Fprintf(color.Output, "%s Downloaded resource %s version %s\n", color.HiGreenString("i"), color.CyanString(resources[0].Name), color.CyanString(release.Release.Version))
			project.Dependencies[args[0]]="^"+release.Release.Version
			config.SaveConfig(project)
			return
		} else {
			fmt.Fprintf(color.Output, "%s Error: %s\n", color.HiYellowString("!"), color.RedString("please, provide at least a resource name"))
			return
		}
	},
}

var initCMD = &cobra.Command{
	Use:   "init",
	Short: "Initializes the grif config file",
	Long: "Creates a basic grif config file in order to start managing plugins",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := config.Load()
		if err != nil {
			fmt.Fprintf(color.Output, "%s  while initializing: %s\n", color.HiYellowString("!"), color.RedString("please, provide at least a resource name"))
			return
		}
	},
}

func main(){
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintf(color.Output, "%s  while initializing grif: %s\n", color.HiYellowString("!"), color.RedString("please, provide at least a resource name"))
		os.Exit(1)
	}
}

func init() {
	rootCMD.AddCommand(initCMD)
	rootCMD.AddCommand(installCMD)
}