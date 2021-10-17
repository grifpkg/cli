package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/grifpkg/cli/api"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var loadedProject bool = false
var openedProject Project = Project{}
var loadedProjectLock bool = false
var openedProjectLock ProjectLock = ProjectLock{}

func GetProject() (project *Project, lock *ProjectLock, err error){
	if !loadedProject {
		openedProject, err = Load()
		if err != nil {
			return &openedProject, &openedProjectLock, err
		}
		loadedProject=true
	}
	if !loadedProjectLock {
		openedProjectLock, err = LoadLock()
		if err != nil {
			return &openedProject, &openedProjectLock, err
		}
		loadedProjectLock=true
	}
	return &openedProject, &openedProjectLock, nil
}

func (project Project) InstallAll() (release []Release, err error) {
	_, lock, err := GetProject()
	if err != nil {
		return nil, err
	}

	api.LogOne(api.Progress, "matching the resources on the project file with the project-lock file")
	var dependencyMatch = make(map[string]Dependency, 0)

	for _, dependency := range lock.Dependencies {
		if !dependency.InstalledAsDependency {
			api.LogOne(api.Progress, "matching "+dependency.Resource.Readable +"with "+dependency.Resource.Resolved)
			if _, ok := project.Dependencies["@"+dependency.Author.Readable+"/"+dependency.Resource.Readable]; ok {
				dependencyMatch["@"+dependency.Author.Readable+"/"+dependency.Resource.Readable]=dependency
			} else {
				api.LogOne(api.Progress, "removing orphan resource on the project-lock: "+dependency.Resource.Readable + " ("+dependency.Resource.Resolved+")")
				err := lock.remove(dependency.Resource.Resolved)
				if err != nil {
					return nil, err
				}
				api.LogOne(api.Info, "removed orphan resource on the project-lock: "+dependency.Resource.Readable + " ("+dependency.Resource.Resolved+")")
			}
		}
	}

	api.LogOne(api.Info, "found "+strconv.Itoa(len(dependencyMatch))+" cached resource(s) out of "+strconv.Itoa(len(project.Dependencies))+" to be installed")
	for installString, versionTag := range project.Dependencies {
		if _, ok := dependencyMatch[installString]; !ok {
			// a dependency was found on the project file but not on the lock file, resolving the resource and adding it to the lock
			resourceName, resourceAuthor, version := project.ParseInstallString(installString, versionTag)
			api.LogOne(api.Progress, "installing "+resourceName+" ("+versionTag+") (not available on the project lock)")
			resource, release, err := project.Install(resourceName, version, resourceAuthor, nil)
			if err != nil {
				api.LogOne(api.Warn, "error while installing "+resourceName+" ("+versionTag+"): "+err.Error())
			} else {
				api.LogOne(api.Success, "installed "+resource.Name+" (as \""+resourceName+"\") ("+release.Version+") (wasn't available on the project lock)")
			}
		}
	}

	for _, dependency := range dependencyMatch {
		api.LogOne(api.Progress, "installing "+dependency.Resource.Readable+" ("+dependency.Release.Readable+")")
		err := dependency.EnsureInstall()
		if err != nil {
			api.LogOne(api.Warn, "error while installing "+dependency.Resource.Readable+" ("+dependency.Release.Readable+"): "+err.Error())
		}
		api.LogOne(api.Success, "installed "+dependency.Resource.Readable+" ("+dependency.Release.Readable+")")
	}

	err = project.Save()
	if err != nil {
		return nil, err
	}
	err = lock.Save()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (project Project) ParseInstallString(string string, version interface{}) (resourceName string, resourceAuthor interface{}, versionTag interface{}){
	if version == nil ||  strings.HasPrefix(fmt.Sprintf("%v", version),"^") {
		versionTag=nil
	} else {
		versionTag=fmt.Sprintf("%v", version)
	}
	if strings.HasPrefix(string,"@") && strings.Contains(string,"/"){
		resourceAuthor=strings.Split(strings.TrimPrefix(string,"@"),"/")[0]
		resourceName=strings.ReplaceAll(string,"@"+fmt.Sprintf("%v", resourceAuthor)+"/","")
	} else {
		resourceAuthor=nil
		resourceName=string
	}
	return resourceName, resourceAuthor, versionTag
}

func (project Project) GetExpectedFilePath(fileName string, fileExtension string, dependencyType int) (folderPath string, filePath string){
	var defaultPath string = ""
	for _, path := range project.InstallPaths {
		path.Path = strings.ReplaceAll(strings.ReplaceAll(path.Path,"./",""),"/",string(os.PathSeparator))
		if path.Extension == fileExtension {
			if path.Type == -1 {
				defaultPath = path.Path
			} else if path.Extension == fileExtension && path.Type == dependencyType {
				return path.Path, path.Path+fileName
			}
		}
	}
	return defaultPath, defaultPath+fileName
}

type Project struct {
	Name         	string		 					`json:"name"`
	Game     	 	string      					`json:"game"`
	Version  		string      					`json:"version"`
	Software     	string            				`json:"software"`
	Dependencies 	map[string]string 				`json:"dependencies"`
	InstallPaths 	[]InstallPath     				`json:"installPaths"`
	ExcludeFiles	[]string						`json:"excludeFiles"`
}

type InstallPath struct {
	Type int 			`json:"type"`
	Extension string 	`json:"extension"`
	Path string 		`json:"path"`
}

func getDefaultInstallPaths() []InstallPath {
	return []InstallPath{
		{
			Type: 0,
			Extension: "jar",
			Path: "./plugins/",
		},
		{
			Type: 0,
			Extension: "sk",
			Path: "./plugins/Skript/scripts/",
		},
		{
			Type: -1,
			Extension: "*",
			Path: "./packages/",
		},
	}
}

func (project Project) Install(resourceName string, version interface{}, resourceAuthor interface{}, service interface{}) (resource Resource, release Release, err error){
	resources, err := QueryResources(resourceName, resourceAuthor, service)
	if err != nil {
		return Resource{}, Release{}, err
	}
	if len(resources)<=0 {
		return Resource{}, Release{}, errors.New("couldn't find a suitable resource")
	}
	resource = resources[0]
	release, err = resource.GetRelease(version, nil)
	if err != nil {
		return Resource{}, Release{}, err
	}
	downloadable, err := release.GetDownloadable(nil)
	if err != nil {
		return Resource{}, Release{}, err
	}
	err = downloadable.Install(resourceName,version!=nil,false,nil,nil)
	if err != nil {
		return Resource{}, Release{}, err
	}
	return resource, release, nil
}

func (project Project) AddExclude(exclude... string) (excluded []string, err error){
	for _, arg := range exclude {
		found:=false
		for _, b := range project.ExcludeFiles {
			if b == arg {
				found=true
				break
			}
		}
		if !found {
			project.ExcludeFiles = append(project.ExcludeFiles, arg)
		}
	}
	err = project.Save()
	if err != nil {
		return nil, err
	}
	return project.ExcludeFiles, err
}

func getProjectFile(lockFile bool) (string, error){
	path, err := os.Getwd()
	var fileName = "project.json"
	if lockFile {
		fileName="project-lock.json"
	}
	return path+string(os.PathSeparator)+fileName, err
}

func createProjectFile(path string) (project Project, err error){

	project = Default()
	api.LogOne(api.Progress,"creating project file")
	answers := struct {
		Name string
		Game string
		Version string
		Software string
	}{}

	err = api.Ask([]*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What's the name of this project/server?",
				Default: "unnamed",
			},
		},
		{
			Name: "game",
			Prompt: &survey.Select{
				Message: "Please, select a game",
				Options: []string{"@minecraft/java"},
				Default: "@minecraft/java",
			},
		},
		{
			Name: "version",
			Prompt: &survey.Input{
				Message: "Which game version is this server running on?",
				Default: "latest",
			},
		},
		{
			Name: "software",
			Prompt: &survey.Select{
				Message: "Which server software are you using (if you are using a fork, select the closest parent)",
				Options: []string{"@spigotmc/spigot", "@spigotmc/bungeecord", "@papermc/paper", "@papermc/waterfall"},
				Default: "@minecraft/java",
			},
		},
	}, &answers)
	if err != nil {
		return Project{}, err
	}

	project.Name=answers.Name
	project.Software=answers.Software
	project.Game=answers.Game
	project.Version=answers.Version

	api.LogOne(api.Progress,"encoding project file")
	file, err := json.MarshalIndent(project, "", "	")
	if err!=nil {
		return Project{},err
	}

	api.LogOne(api.Progress,"saving project file")
	err = ioutil.WriteFile(path, file, 0644)
	if err!=nil {
		return Project{},err
	}
	return project,err
}

func Load() (project Project, err error) {
	api.LogOne(api.Progress,"loading project file")
	project = Default()
	err = nil
	finalPath, err := getProjectFile(false)
	if err == nil {
		file, err := os.Open(finalPath)
		if os.IsNotExist(err){
			err = nil // not an error, file doesn't exist, but we can create one
			project, err = createProjectFile(finalPath)
			if err!=nil {
				return Project{},err
			}
		} else if err==nil {
			api.LogOne(api.Progress,"reading project file")
			byteValue, err := ioutil.ReadAll(file)
			if err!=nil {
				return Project{},err
			}

			api.LogOne(api.Progress,"decoding project file")
			err = json.Unmarshal(byteValue, &project)
			if err!=nil {
				return Project{},err
			}
		}
	}
	return project, err
}

func Default() Project {
	return Project{
		Name: "unnamed",
		Game: "@minecraft/java",
		Version: "latest",
		Software: "@spigotmc/spigot",
		InstallPaths: getDefaultInstallPaths(),
		Dependencies: make(map[string]string, 0),
		ExcludeFiles: make([]string, 0),
	}
}

func (project Project) Save() (err error){
	file, err := json.MarshalIndent(project, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("project.json", file, 0644)
	return err
}