package project

import (
	"archive/zip"
	"errors"
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/urlSuggestion"
	"github.com/segmentio/ksuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type DownloadableRelease struct {
	Url     string `json:"url"`
	Release Release `json:"release"`
	Resource Resource `json:"resource"`
	Suggestion interface{}
}

func (downloadableRelease DownloadableRelease) Install(providedName string, forcedVersion bool, skipDependencies bool, parentResource interface{}, mainModule interface{}) (err error){

	api.LogOne(api.Progress,"installing release")
	project, projectLock, err := GetProject()
	if err!= nil {
		return err
	}

	// 0. check if the resource is a dependency, if it is, unmark it
	var resolvedDependencies map[string]Resource = map[string]Resource{}
	if !skipDependencies {
		resolvedDependencies, err = downloadableRelease.Release.GetResolvedDependencies()
		if err!= nil {
			return err
		}
	}

	// download

	filePaths, err := downloadableRelease.DownloadFile()
	if err != nil {
		return err
	}

	// install

	_, err = projectLock.AddDependency(providedName, downloadableRelease, resolvedDependencies, filePaths, parentResource, downloadableRelease.Suggestion, nil)
	if err != nil {
		return err
	}

	var versionTag = downloadableRelease.Release.Version
	if !forcedVersion {
		versionTag = "^" + downloadableRelease.Release.Version
	}
	if parentResource == nil {
		project.Dependencies["@"+downloadableRelease.Resource.Author.Username+"/"+providedName]=versionTag
	}

	err = project.Save()
	if err != nil {
		return err
	}
	err = projectLock.Save()
	if err != nil {
		return err
	}

	return nil
}

func (downloadableRelease DownloadableRelease) DownloadFile() (paths []string, err error) {

	// download
	api.LogOne(api.Progress,"downloading file")

	resp, err := http.Get(downloadableRelease.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	project, _, err := GetProject()
	if err!= nil {
		return nil, err
	}

	addedPath, folderPath := project.GetExpectedFilePath(downloadableRelease.Release.FileName.(string), downloadableRelease.Release.FileExtension.(string), 0)

	finalPath := path+string(os.PathSeparator)+addedPath
	mainFilePath := finalPath+downloadableRelease.Release.FileName.(string)

	err = os.Mkdir(finalPath, 0700)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	out, err := os.Create(mainFilePath)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	// unzip if necessary
	var movedFiles = make([]string, 0)
	if downloadableRelease.Release.FileExtension=="zip" {
		movedFiles, err = handleZip(folderPath, mainFilePath)
		if err != nil {
			return nil, err
		}
	}

	// return paths
	if downloadableRelease.Release.FileExtension=="zip" {
		if len(movedFiles)>0 {
			return movedFiles, nil
		} else {
			return nil, errors.New("no files were moved, please, check if the returned zip file was empty or if you are ignoring all contents")
		}
	} else {
		return append(movedFiles, mainFilePath), nil
	}
}
func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, errors.New("illegal file path")
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return nil, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		_, err = io.Copy(outFile, rc)
		err = outFile.Close()
		if err != nil {
			return nil, err
		}
		err = rc.Close()
		if err != nil {
			return nil, err
		}
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func handleZip(folderPath string, zipPath string) (filePaths []string, err error) {
	api.LogOne(api.Progress,"handling zip file")
	filePaths = make([]string, 0)
	separator := string(os.PathSeparator)
	api.LogOne(api.Progress,"creating temporal folder")
	tempFolder := folderPath+ksuid.New().String()+separator
	_ = os.Mkdir(tempFolder, 0700)
	api.LogOne(api.Progress,"unzipping")
	_, err = unzip(zipPath, tempFolder)
	if err != nil {
		return nil, err
	}
	project, _, err := GetProject()
	if err != nil {
		return nil, err
	}
	api.LogOne(api.Progress,"walking through the extracted files")
	err = filepath.Walk(tempFolder,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				excluded := false
				for _, excludeSchema := range project.ExcludeFiles {
					res, _ := regexp.MatchString(excludeSchema, info.Name())
					if res {
						excluded = true
						break
					}
				}
				if !excluded {
					_, filePath := project.GetExpectedFilePath(info.Name(), filepath.Ext(info.Name()), 0)
					err = os.Rename(path, filePath)
					filePaths = append(filePaths, filePath)
				}
			}
			return nil
		})
	api.LogOne(api.Progress,"removing zip file")
	err = os.Remove(zipPath)
	if err != nil {
		return nil, err
	}
	api.LogOne(api.Progress,"removing temporal folder")
	err = os.RemoveAll(tempFolder)
	if err != nil {
		return nil, err
	}
	return filePaths, err
}

func DownloadableReleaseFromSuggestion(suggestion urlSuggestion.UrlSuggestion, release Release, resource interface{}) DownloadableRelease {
	var resourceFinal = Resource{}
	if resource!=nil{
		resourceFinal = resource.(Resource)
	}
	return DownloadableRelease{suggestion.Url,Release{release.Id,release.Service,release.ReleaseId,release.Version,release.Creation,suggestion.FileName,extensionFromFileName(suggestion.FileName),0,nil,nil,nil,nil,nil,nil,nil,true, resourceFinal},resourceFinal, suggestion}
}

func extensionFromFileName(fileName string) interface{} {
	parts := strings.Split(fileName,".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return nil
}