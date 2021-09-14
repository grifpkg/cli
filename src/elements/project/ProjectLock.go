package project

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/djherbis/times"
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/urlSuggestion"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ProjectLock struct {
	Dependencies []Dependency	`json:"dependencies"` // dependency list
}

func (projectLock *ProjectLock) AddDependency(providedName string, downloadableRelease DownloadableRelease, resolvedDependencies map[string]Resource, filePaths []string, parent interface{}, suggestion interface{}, mainModule interface{}) (dependency Dependency, err error){

	// 1. check if the dependency is already there

	api.LogOne(api.Progress,"looking for existing dependencies")
	var newDependency bool = false
	var existingDependency interface{} = nil
	for _, dependency := range projectLock.Dependencies {
		if dependency.Resource.Resolved==downloadableRelease.Resource.Id {
			existingDependency = dependency
			break
		}
	}
	if existingDependency==nil {
		api.LogOne(api.Progress,"creating dependency")

		// 1.1 if not, create it

		newDependency = true
		existingDependency = Dependency{
			Author: CreateCheckableHash(downloadableRelease.Resource.Author.Username,downloadableRelease.Resource.Author.Id.(string)),
			Resource: CreateCheckableHash(providedName,downloadableRelease.Resource.Id),
			Release: CreateCheckableHash(downloadableRelease.Release.Version,downloadableRelease.Release.Id),
			DependencyOf: make([]string, 0),
			MainModule: mainModule,
			InstalledAsDependency: parent!=nil,
		}
	}

	// 2. recalculate the dependency tree
	api.LogOne(api.Progress,"calculating dependency tree")

	var resolvedDependenciesHashed []CheckableHash = make([]CheckableHash, 0)
	var dependencyTreeIds string =  downloadableRelease.Resource.Id
	for name, resource := range resolvedDependencies {
		dependencyChild := CreateCheckableHash(name,resource.Id)
		resolvedDependenciesHashed = append(resolvedDependenciesHashed, dependencyChild)
		dependencyTreeIds+=resource.Id
		// get release from dependency and append it to the dependency list
		release, err := resource.GetRelease(nil, nil)
		if err != nil {
			return Dependency{}, err
		}
		downloadable, err := release.GetDownloadable(nil)
		if err != nil {
			return Dependency{}, err
		}
		err = downloadable.Install(name,false,false,downloadableRelease.Resource, nil)
		if err != nil {
			return Dependency{}, err
		}
	}
	if !newDependency {
		api.LogOne(api.Progress,"recalculating usages")

		// 2.1 if the dependency already existed, check for dropped dependencies

		var dependenciesToRemoveUsageFrom []CheckableHash
		for _, existingHashedDependency := range existingDependency.(Dependency).ResolvedDependencies {
			for _, hashedDependency := range resolvedDependenciesHashed {
				foundDependency := false
					if existingHashedDependency.Hash==hashedDependency.Hash {
						foundDependency = true
						break
					}
				if !foundDependency {
					dependenciesToRemoveUsageFrom = append(dependenciesToRemoveUsageFrom, existingHashedDependency)
				}
			}
		}

		if len(dependenciesToRemoveUsageFrom)>0 {
			api.LogOne(api.Progress,"removing dropped usages")
			for _, droppedDependency := range dependenciesToRemoveUsageFrom {
				err := projectLock.removeUsage(droppedDependency.Resolved, downloadableRelease.Resource.Id)
				if err != nil {
					return Dependency{}, err
				}
			}
		}

	}
	dependencyTreeHash := md5.Sum([]byte(dependencyTreeIds))
	dependency = existingDependency.(Dependency)

	if !newDependency {
		if dependency.InstalledAsDependency && parent != nil {
			dependency.InstalledAsDependency=false
		}
	}

	// 3. if the dependency is installed as a child dependency, add the list of parent dependencies using this dependency

	api.LogOne(api.Progress,"recalculating usages on all dependencies")
	var dependencyOf []string = dependency.DependencyOf
	if parent!=nil {
		existsOnUsageList := false
		for _, resourceId := range dependencyOf {
			if parent.(Resource).Id==resourceId{
				existsOnUsageList = true
				break
			}
		}
		if !existsOnUsageList {
			dependencyOf = append(dependencyOf, parent.(Resource).Id)
		}
	}

	// 4. add the suggestion url used if any

	api.LogOne(api.Progress,"specifying fallback suggestion url")
	var fallbackSuggestionId interface{} = nil
	if suggestion!=nil {
		fallbackSuggestionId = suggestion.(urlSuggestion.UrlSuggestion).Id
	}
	dependency.FallbackSuggestion=fallbackSuggestionId

	// 5. create file-link data

	api.LogOne(api.Progress,"creating file fingerprints")
	var fileFingerprints []DependencyFile = make([]DependencyFile, 0)
	for _, filePath := range filePaths {
		extension := filepath.Ext(filePath)
		if strings.HasPrefix(extension, ".") {
			extension = strings.TrimPrefix(extension,".")
		}
		fingerprint, err := fileFingerprint(filePath)
		if err != nil {
			return Dependency{}, err
		}
		fileFingerprints = append(fileFingerprints, DependencyFile{
			Fingerprint: fingerprint,
			Extension: extension,
		})
	}

	// 6. fill the object with all the data generated

	api.LogOne(api.Progress,"filling dependency data")
	dependency.ResolvedDependencies=resolvedDependenciesHashed
	dependency.DependencyTreeHash=hex.EncodeToString(dependencyTreeHash[:])
	dependency.DependencyOf=dependencyOf
	dependency.Fingerprints=fileFingerprints

	// 7. submit the data to the lock file

	api.LogOne(api.Progress,"submitting dependency to the lock file")
	if newDependency {
		projectLock.Dependencies = append(projectLock.Dependencies, dependency)
	} else {
		err := projectLock.mod(dependency.Resource.Resolved, dependency)
		if err != nil {
			return Dependency{}, err
		}
	}

	return dependency,err
}

func (projectLock *ProjectLock) deleteFile(fingerprint string, extension string) (err error){
	api.LogOne(api.Progress,"deleting file")
	return nil
}

func (projectLock *ProjectLock) remove(droppedId string) (err error){
	api.LogOne(api.Progress,"removing dependency and looking for usages")
	for i, dependency := range projectLock.Dependencies {
		if dependency.Resource.Resolved==droppedId {
			if dependency.InstalledAsDependency && len(dependency.DependencyOf) > 0 {
				return errors.New("this resource is a dependency of another resource")
			}
			projectLock.Dependencies = append(projectLock.Dependencies[:i], projectLock.Dependencies[i+1:]...)
			for _, fingerprint := range dependency.Fingerprints {
				err := projectLock.deleteFile(fingerprint.Fingerprint,fingerprint.Extension)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	return errors.New("unknown dependency to remove")
}

func (projectLock *ProjectLock) removeUsage(droppedId string, parentId string) (err error){
	api.LogOne(api.Progress,"removing usage")
	for _, dependency := range projectLock.Dependencies {
		if dependency.Resource.Resolved==droppedId {
			foundIndex := -1
			for o, parentUsage := range dependency.DependencyOf {
				if parentUsage == parentId {
					foundIndex = o
					break
				}
			}
			if foundIndex > -1 {
				dependency.DependencyOf = append(dependency.DependencyOf[:foundIndex], dependency.DependencyOf[foundIndex+1:]...)
				if len(dependency.DependencyOf) <= 0 && dependency.InstalledAsDependency {
					err := projectLock.mod(droppedId,dependency)
					if err != nil {
						return err
					}
					err = projectLock.remove(droppedId)
					if err != nil {
						return err
					}
				} else {
					err := projectLock.mod(droppedId,dependency)
					if err != nil {
						return err
					}
				}
				return nil
			} else {
				return errors.New("the dropped dependency isn't being used by the defined parent dependency")
			}
		}
	}
	return errors.New("unknown child dependency")
}

func (projectLock *ProjectLock) mod(resourceId string, newDependency Dependency) (err error){
	for i, dependency := range projectLock.Dependencies {
		if dependency.Resource.Resolved==resourceId {
			projectLock.Dependencies[i]=newDependency
			return nil
		}
	}
	return errors.New("unknown dependency")
}

func fileFingerprint(path string) (fingerprint string, err error){
	api.LogOne(api.Progress,"creating fingerprint")
	t, err := times.Stat(path)
	if err != nil {
		return "", err
	}
	var timeVariant time.Time
	if t.HasBirthTime() {
		timeVariant = t.BirthTime()
	} else if t.HasChangeTime() {
		timeVariant = t.ChangeTime()
	} else {
		return "", errors.New("invalid time variant while generating file fingerprint")
	}
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	// get the size
	size := fi.Size()
	fingerprint = strconv.FormatInt(size, 10) + strconv.FormatInt(timeVariant.UnixMicro(), 10)
	fingerprintHash := md5.Sum([]byte(fingerprint))
	return hex.EncodeToString(fingerprintHash[:]), nil
}

type Dependency struct {
	Type int									`json:"type"` // dependency type, 0 for plugin, 1 for mod, -1 for unknown
	Author CheckableHash						`json:"author"` // author is technically the most restrictive selector for resource querying, so project.json tries to find the resource by name on matching author names
	Resource CheckableHash 						`json:"resource"` // make sure the readable resource actually IS the resource on the id using a hash
	Release CheckableHash 						`json:"release"` // make sure the readable version actually IS the version on the id using a hash
	ResolvedDependencies []CheckableHash		`json:"resolvedDependencies"` // the resolved non-optional dependencies as check-able hashes to make sure the installed dependencies are those shown on the readable value
	DependencyTreeHash string					`json:"dependencyTreeHash"` // hash based on all the resolved dependencies to avoid altering the dependency list solely based on the array
	DependencyOf []string						`json:"dependencyOf"` // the list of resource ids that list this as a dependency
	FallbackSuggestion interface{}				`json:"fallbackSuggestion"` // fallback url suggestion
	Fingerprints []DependencyFile				`json:"files"` // file info
	InstalledAsDependency bool					`json:"installedAsDependency"` // install mode, false if not a child dependency. if true, it will get deleted if no other dependencies are using it
	MainModule interface{}						`json:"mainModule"` // check-able hash, this will only be available if this dependency is a sub-module, when uninstalling the core resource, all the sub-modules get removed too
}

type DependencyFile struct {
	Fingerprint string							`json:"fingerprint"`
	Extension string							`json:"extension"`
}

func (dependency *Dependency) AddParentUsage(parent Dependency){
	for _, existingUsage := range dependency.DependencyOf {
		if existingUsage == parent.Resource.Resolved {
			return
		}
	}
	dependency.DependencyOf = append(dependency.DependencyOf,parent.Resource.Readable)
}

func (dependency *Dependency) EnsureInstall() (err error){
	project, lock, err := GetProject()
	if err!= nil {
		return err
	}

	// 1. resolve the file fingerprints for all files on a folder that might contain the file we are looking for

	fingerprints := make(map[string]map[string]string, 0) // [extension][fingerprint]=filename
	wantedExtensions := make([]string, 0)
	wd, err := os.Getwd()
	wd += string(os.PathSeparator)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	for _, fingerprint := range dependency.Fingerprints {
		found := false
		for _, b := range wantedExtensions {
			if b == fingerprint.Extension {
				found = true
				break
			}
		}
		if !found {
			wantedExtensions = append(wantedExtensions, fingerprint.Extension)
		}
	}
	for _, extension := range wantedExtensions {
		folderPath, _ := project.GetExpectedFilePath("exampleFileName."+extension,extension,0)
		fingerprints[extension] = make(map[string]string, 0)
		files, err := ioutil.ReadDir(wd+folderPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
		for _, file := range files {
			if !file.IsDir() && strings.TrimPrefix(filepath.Ext(file.Name()),".") == extension {
				fingerprint, err := fileFingerprint(wd+folderPath+file.Name())
				if err != nil {
					return err
				}
				fingerprints[extension][fingerprint]=folderPath+file.Name()
			}
		}
	}

	// 2. check if the fingerprint is present

	missingFiles := false
	for _, fingerprint := range dependency.Fingerprints {
		foundFingerprint := false
		for existingFileFingerprint, _ := range fingerprints[fingerprint.Extension] {
			if existingFileFingerprint==fingerprint.Fingerprint {
				foundFingerprint=true
				break
			}
		}
		if !foundFingerprint {
			missingFiles=true
			break
		}
	}


	// 3. if the files are missing, download them
	if missingFiles {
		release := dependency.AsRelease()
		downloadable, err := release.GetDownloadable(dependency.FallbackSuggestion)
		if err != nil {
			return err
		}
		_, err = downloadable.DownloadFile()
		if err != nil {
			return err
		}
	}

	// 4. ensure dependencies are installed
	for _, resolvedDependency := range dependency.ResolvedDependencies {
		subDependency := lock.GetDependency(resolvedDependency.Resolved)
		err := subDependency.EnsureInstall()
		if err != nil {
			return err
		}
	}
	return nil
}

func (projectLock ProjectLock) GetDependency(id string) (dependency Dependency){
	for _, existingDependency := range projectLock.Dependencies {
		if existingDependency.Resource.Resolved == id {
			return existingDependency
		}
	}
	return Dependency{}
}

func (dependency Dependency) AsRelease() (release Release) {
	return Release {
		Id: dependency.Release.Resolved,
		Version: dependency.Release.Readable,
		Parent: Resource {
			Id: dependency.Resource.Resolved,
			Name: dependency.Resource.Readable,
		},
	}
}

type CheckableHash struct {
	Readable string								`json:"name"` // readable, user-input value
	Resolved string								`json:"id"` // resolved, actual value
	Hash string									`json:"hash"` // readable+resolved
}

func CreateCheckableHash(readableValue string, resolvedValue string) (checkableHash CheckableHash){
	hash := md5.Sum([]byte(readableValue+resolvedValue))
	return CheckableHash{
		Readable: readableValue,
		Resolved: resolvedValue,
		Hash: hex.EncodeToString(hash[:]),
	}
}


func LoadLock() (projectLock ProjectLock, err error) {
	api.LogOne(api.Progress,"loading project-lock file")
	projectLock = ProjectLock{
		Dependencies: make([]Dependency, 0),
	}
	err = nil
	finalPath, err := getProjectFile(true)
	if err == nil {
		file, err := os.Open(finalPath)
		if os.IsNotExist(err){
			err = nil // not an error, file doesn't exist, but we can create one

			api.LogOne(api.Progress,"encoding project-lock file")
			file, err := json.MarshalIndent(projectLock, "", "	")
			if err!=nil {
				return ProjectLock{},err
			}

			api.LogOne(api.Progress,"saving project-lock file")
			err = ioutil.WriteFile(finalPath, file, 0644)
			if err!=nil {
				return ProjectLock{},err
			}
			if err!=nil {
				return ProjectLock{},err
			}
		} else if err==nil {
			api.LogOne(api.Progress,"reading project-lock file")
			byteValue, err := ioutil.ReadAll(file)
			if err!=nil {
				return ProjectLock{},err
			}

			api.LogOne(api.Progress,"decoding project-lock file")
			err = json.Unmarshal(byteValue, &projectLock)
			if err!=nil {
				return ProjectLock{},err
			}
		}
	} else {
		return ProjectLock{},err
	}
	return projectLock, err
}

func (projectLock ProjectLock) Save() (err error){
	file, err := json.MarshalIndent(projectLock, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("project-lock.json", file, 0644)
	return err
}
