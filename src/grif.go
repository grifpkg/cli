package main

import (
	"github.com/grifpkg/cli/client"
	"github.com/grifpkg/cli/config"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strconv"
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
			dependencyTag := "@"+resources[0].Author.Username+"/"+dependency.Resource

			versionHash := md5.New()
			versionHash.Write([]byte(release.Release.Id+release.Release.Version))

			resourceHash := md5.New()
			resourceHash.Write([]byte(dependencyTag+dependency.Resource))

			project.Dependencies[dependencyTag]=config.DependencyIdentifier{Version: "^"+release.Release.Version, Resource: resources[0].Id, Release: release.Release.Id, Hash: []string{hex.EncodeToString(resourceHash.Sum(nil)),hex.EncodeToString(versionHash.Sum(nil))}}
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

var excludeCMD = &cobra.Command{
	Use:   "exclude",
	Aliases: []string{"e"},
	Short: "Adds to the file-exclude list file names or regex",
	Long: "Use as many arguments as you want forming regex expressions or file names to add to the file-exclude list to skip certain files during installs or updates",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project, err := config.Load()
		if err != nil {
			fmt.Fprintf(color.Output, "%s  while initializing: %s\n", color.HiYellowString("!"), color.RedString(err.Error()))
			return
		}
		var added int = 0
		for _, arg := range args {
			found:=false
			for _, b := range project.ExcludeFiles {
				if b == arg {
					found=true
					break
				}
			}
			if !found {
				added++
				project.ExcludeFiles = append(project.ExcludeFiles, arg)
			}
		}
		err = config.SaveConfig(project)
		if err != nil {
			fmt.Fprintf(color.Output, "%s  while updating the project file: %s\n", color.HiYellowString("!"), color.RedString(err.Error()))
			return
		}
		fmt.Fprintf(color.Output, "%s added %s names/regex to the file-exclude list\n", color.HiGreenString("i"), color.CyanString(strconv.Itoa(added)))
	},
}

func main(){
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintf(color.Output, "%s  while initializing grif: %s\n", color.HiYellowString("!"), color.RedString(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCMD.AddCommand(initCMD)
	rootCMD.AddCommand(installCMD)
	rootCMD.AddCommand(excludeCMD)
}