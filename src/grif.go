package main

import (
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/project"
	"github.com/grifpkg/cli/elements/session"
	"github.com/grifpkg/cli/installer"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

var rootCMD = &cobra.Command{
	Use:   "grif",
	Short: "check the installed version",
	Run: func(cmd *cobra.Command, args []string) {
		api.LogOne(api.Info,"grifpkg version "+api.Version+", upgrade your version with grif upgrade")
	},
}

var linkCMD = &cobra.Command{
	Use:   "link",
	Aliases: []string{"li"},
	Short: "links an external service to your grifpkg account",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := session.Get()
		if err != nil {
			api.LogOne(api.Warn, err.Error())
			return
		}
		err = session.Link()
		if err != nil {
			api.LogOne(api.Warn, err.Error())
			return
		}
		api.LogOne(api.Success,"successfully linked account")
	},
}

var importCMD = &cobra.Command{
	Use:   "import",
	Aliases: []string{"im"},
	Short: "recognizes the installed dependencies and adds them to the project",
	Run: func(cmd *cobra.Command, args []string) {
		api.LogOne(api.Warn, "the import command is not ready yet")
	},
}

var updateCMD = &cobra.Command{
	Use:   "update",
	Aliases: []string{"u"},
	Short: "updates resource releases to the latest versions as long as they are set to be updated",
	Run: func(cmd *cobra.Command, args []string) {
		api.LogOne(api.Warn, "the update command is not ready yet")
	},
}

var installCMD = &cobra.Command{
	Use:   "install",
	Aliases: []string{"i"},
	Short: "installs a package and its dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		project, _, err := project.GetProject()
		if len(args)>0 {
			if err != nil {
				api.LogOne(api.Warn, err.Error())
			}
			name, author, version := project.ParseInstallString(args[0],nil)
			resource, release, err := project.Install(name,author,version,nil)
			if err != nil {
				api.LogOne(api.Warn, err.Error())
				return
			}
			api.LogOne(api.Success,"installed "+resource.Name+" by "+resource.Author.Username+", version "+release.Version)
			return
		} else {
			_, err := project.InstallAll()
			if err != nil {
				api.LogOne(api.Warn, err.Error())
			}
			return
		}
	},
}

var initCMD = &cobra.Command{
	Use:   "init",
	Short: "initializes a grif config file",
	Run: func(cmd *cobra.Command, args []string) {
		_, _, err := project.GetProject()
		if err != nil {
			api.LogOne(api.Warn, err.Error())
			return
		}
		api.LogOne(api.Success, "the project file should be present on the current directory")
	},
}

var excludeCMD = &cobra.Command{
	Use:   "exclude",
	Aliases: []string{"e"},
	Short: "adds to the file-exclude list file names or regex, every param is an exclude regex",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project, _, err := project.GetProject()
		if err != nil {
			api.LogOne(api.Warn, err.Error())
			return
		}
		excluded, err := project.AddExclude(args...)
		if err != nil {
			api.LogOne(api.Warn, err.Error())
			return
		}
		api.LogOne(api.Success, "excluding "+strconv.Itoa(len(excluded))+" name(s)/regex")
	},
}
var upgradeCMD = &cobra.Command{
	Use:   "upgrade",
	Aliases: []string{"up"},
	Short: "upgrades grif binaries to the latest version and makes them globally executable",
	Run: func(cmd *cobra.Command, args []string) {
		release, err := installer.Install()
		if err != nil {
			api.LogOne(api.Warn,err.Error())
		}
		api.LogOne(api.Success,"upgraded to "+release.TagName)
	},
}

func main(){
	_ = sentry.Init(sentry.ClientOptions{
		Dsn: "https://e796fb7d9ac5498c860a40b798e1fb51@o958174.ingest.sentry.io/5906985",
		Release: "go@"+api.Version,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			beautifulJsonByte, _ := json.MarshalIndent(&event.Exception, "", "   ")
			api.LogOne(api.Warn,string(beautifulJsonByte))
			api.LogOne(api.Info,"we got an un handled error. we are sending this error to sentry in order to quickly solve it. in the future, you'll be able to choose whether you want automatic error reporting, but for the moment I'm too lazy to implement it and I feel like reporting on early stages is actually really mandatory. If you wanna fix this error yourself, the error should be appearing just above this line; feel free to commit changes to https://github.com/grifpkg/cli/ and report it to our Discord server so you nag someone and tell them about event id")
			return event
		},

	})

	defer sentry.Flush(time.Second * 2)
	defer sentry.Recover()

	if err := rootCMD.Execute(); err != nil {
		api.LogOne(api.Warn,err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCMD.AddCommand(initCMD)
	rootCMD.AddCommand(upgradeCMD)
	rootCMD.AddCommand(installCMD)
	rootCMD.AddCommand(updateCMD)
	rootCMD.AddCommand(importCMD)
	rootCMD.AddCommand(excludeCMD)
	rootCMD.AddCommand(linkCMD)
}