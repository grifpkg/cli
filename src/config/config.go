package config

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
)

func GetProjectPath(lockFile bool) (string, error){
	path, err := os.Getwd()
	var fileName = "project.json"
	if lockFile {
		fileName="project.lock.json"
	}
	return path+string(os.PathSeparator)+fileName, err
}

func Load() (Project, error) {
	var project = GetDefaultProject()
	var err error = nil
	finalPath, err := GetProjectPath(false)
	if err == nil {
		file, err := os.Open(finalPath)
		if os.IsNotExist(err){
			err = nil // not an error, file doesn't exist, but we can create one

			// --- project setup ---
			// type
			projectType := &survey.Select{
				Message: "What kind of project/server are you developing?",
				Options: []string{"minecraft/bukkit","minecraft/nukkit"},
				Default: "minecraft/bukkit",
			}
			err = survey.AskOne(projectType, &project.Type, survey.WithValidator(survey.Required))

			// name
			projectName := &survey.Input{
				Message: "What's the name of this project/server?",
				Default: "unnamed",
			}
			err = survey.AskOne(projectName, &project.Name, survey.WithValidator(survey.Required))

			file, _ := json.MarshalIndent(project, "", "	")
			err = ioutil.WriteFile(finalPath, file, 0644)
			fmt.Fprintf(color.Output, "%s Initialized %s\n", color.HiGreenString("i"), color.CyanString(project.Name))
		} else if err==nil {
			byteValue, _ := ioutil.ReadAll(file)
			err = json.Unmarshal(byteValue, &project)
		}
	}
	return project, err
}

func SaveConfig(project Project) error{
	finalPath, err := GetProjectPath(false)
	file, err := json.MarshalIndent(project, "", "	")
	err = ioutil.WriteFile(finalPath, file, 0644)
	if err!=nil {
		fmt.Fprintf(color.Output, "%s Error while saving project file: %s\n", color.HiYellowString("!"), color.RedString(err.Error()))
	}
	return err
}