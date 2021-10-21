package deployment

import (
	"encoding/json"
	"errors"
	"github.com/grifpkg/cli/api"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DirectDeploy deploys directly with a command, f.e. creating a java vm/**
func DirectDeploy(fork string, version string, path string, args map[string]interface{}) (err error){
	version=strings.ToLower(version)
	if version=="latest" {
		version, err = GetLatestVersion("minecraft")
		if err != nil {
			return err
		}
	}
	downloadURL, err := GetLatestSoftwareVersion(fork,version)
	api.LogOne(api.Progress,"downloading the server software")
	_, err = api.DownloadFile(downloadURL,nil,"bin.jar")
	if err != nil {
		return err
	}
	api.LogOne(api.Progress, "downloaded, serving locally")
	var cmd = "java -jar bin.jar"
	if runtime.GOOS == "windows" {
		_, err = exec.Command("cmd", "/c", cmd).Output()
		if err != nil {
			return err
		}
	} else {
		_, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return err
		}
	}
	return nil
}

// DockerDeploy deploys locally with a docker image
func DockerDeploy(fork string, version string, path string, envs map[string]interface{}, image string) (err error){
	return errors.New("not implemented")
}

// CloudDeploy deploys to a cloud service, AWS, Google Cloud or purecore hosting
func CloudDeploy(fork string, version string, path string, opts map[string]interface{}, service string, token string, identifier interface{}) (err error){
	return errors.New("not implemented")
}

func GetLatestSoftwareVersion(software string, gameVersion string) (downloadURL string, err error){
	api.LogOne(api.Progress,"fetching latest software version")
	software=strings.ToLower(software)
	if software=="@papermc/paper" {

		type PaperBuilds struct {
			ProjectId string `json:"project_id"`
			ProjectName string `json:"project_name"`
			Version string `json:"version"`
			Builds []int `json:"builds"`
		}

		type PaperBuild struct {
			ProjectId   string    `json:"project_id"`
			ProjectName string    `json:"project_name"`
			Version     string    `json:"version"`
			Build       int       `json:"build"`
			Time        time.Time `json:"time"`
			Changes     []struct {
				Commit  string `json:"commit"`
				Summary string `json:"summary"`
				Message string `json:"message"`
			} `json:"changes"`
			Downloads struct {
				Application struct {
					Name   string `json:"name"`
					Sha256 string `json:"sha256"`
				} `json:"application"`
			} `json:"downloads"`
		}

		result, err := api.WorldWideWebRequest("https://papermc.io/api/v2/projects/paper/versions/" + gameVersion, map[string]interface{}{})
		if err != nil {
			return "", err
		}

		var buildData = PaperBuilds{}

		err = json.NewDecoder(result).Decode(&buildData)
		if err != nil {
			return "", err
		}

		var latestBuildData = PaperBuild{}

		result, err = api.WorldWideWebRequest("https://papermc.io/api/v2/projects/paper/versions/"+gameVersion+"/builds/"+strconv.Itoa(buildData.Builds[len(buildData.Builds)-1]), map[string]interface{}{})
		if err != nil {
			return "", err
		}
		err = json.NewDecoder(result).Decode(&latestBuildData)

		return "https://papermc.io/api/v2/projects/paper/versions/"+gameVersion+"/builds/"+strconv.Itoa(buildData.Builds[len(buildData.Builds)-1])+"/downloads/"+latestBuildData.Downloads.Application.Name, nil

	} else if software=="@spigotmc/bungeecord" {
		return "https://ci.md-5.net/job/BungeeCord/lastStableBuild/artifact/bootstrap/target/BungeeCord.jar", nil
	} else {
		return "", errors.New("unsupported software")
	}
}

func GetLatestVersion(game string) (version string,err error){
	api.LogOne(api.Progress,"fetching latest game version")
	if game=="minecraft" {

		type MinecraftLatestReleaseStrings struct {
			Release string	`json:"release"`
			Snapshot string	`json:"snapshot"`
		}

		type MinecraftReleaseInfo struct {
			Id string	`json:"id"`
			Type string	`json:"type"`
			URL string	`json:"url"`
			Time string	`json:"time"`
			ReleaseTime string	`json:"releaseTime"`
		}

		type MinecraftVersionList struct {
			Latest MinecraftLatestReleaseStrings	`json:"latest"`
			Versions []MinecraftReleaseInfo			`json:"versions"`
		}

		var versionData = MinecraftVersionList{}

		result, err := api.WorldWideWebRequest("https://launchermeta.mojang.com/mc/game/version_manifest.json",map[string]interface{}{})
		if err != nil {
			return "", err
		}

		err = json.NewDecoder(result).Decode(&versionData)
		if err != nil {
			return "", err
		}

		return versionData.Latest.Release, nil
	} else {
		return "", errors.New("unsupported game")
	}
}
