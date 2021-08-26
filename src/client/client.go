package client

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/grifpkg/cli/config"
	"github.com/segmentio/ksuid"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

func GetResources(name string, author string, service int) ([]Resource, error) {
	var err error = nil
	var data map[string]string = make(map[string]string)
	data["name"]=name
	if author!="" {
		data["author"]=author
	}
	if service>-1 {
		data["service"]= string(rune(service))
	}
	request, err := request("resource/query/", data)
	resourceList := make([]Resource,0)
	err = json.NewDecoder(request).Decode(&resourceList)
	return resourceList, err
}

func getRelease(resource Resource, version string) (Release, error) {
	var err error = nil
	var data map[string]string = make(map[string]string)
	data["resource"] = resource.Id
	if version!="" {
		data["version"] = version
	}
	request, err := request("resource/release/get/", data)
	release := Release{}
	err = json.NewDecoder(request).Decode(&release)
	return release, err
}

func GetReleases(resource Resource) ([]Release, error) {
	var err error = nil
	request, err := request("resource/release/list/", map[string]string{
		"resource":resource.Id,
	})

	releaseList := make([]Release,0)
	err = json.NewDecoder(request).Decode(&releaseList)
	return releaseList, err
}

func DownloadResource(resource Resource, version string, projectConfig config.Project, releaseId string) (DownloadableRelease, error) {
	var err error = nil
	var release = Release{}

	// gets release
	if releaseId == "" {
		release, err = getRelease(resource,version)
		if err != nil {
			return DownloadableRelease{}, err
		}
		releaseId=release.Id
	}

	var downloadableRelease = DownloadableRelease{}
	if release.HasSuggestions!=nil {
		if release.HasSuggestions == true {
			// gets suggestion list release
			request, err := request("resource/release/suggestion/list/", map[string]string{
				"release": releaseId,
			})
			suggestionList := make([]UrlSuggestion,0)
			err = json.NewDecoder(request).Decode(&suggestionList)
			if err != nil {
				return DownloadableRelease{}, err
			}
			var optionList []string
			for _, suggestion := range suggestionList {
				optionList = append(optionList, suggestion.Id+"\t\t| " + suggestion.Url)
			}
			optionList = append(optionList, "suggest another URL\t| grifpkg.com")
			optionList = append(optionList, "skip\t\t\t| cancel")
			resultURL := ""
			prompt := &survey.Select{
				Message: "This resource is hosted externally, here is a list of suggested download URLs, note this URLs are NOT verified and may not point to the desired target:",
				Options: optionList,
			}
			survey.AskOne(prompt, &resultURL)
			if strings.HasPrefix(resultURL,"suggest another URL"){
				return DownloadableRelease{}, errors.New("suggest a download URL here: https://grifpkg.com/suggest/"+resource.Id+"/")
			} else if strings.HasPrefix(resultURL,"skip") {
				return DownloadableRelease{}, errors.New("skipped release")
			} else {
				downloadableRelease.Url=strings.Split(resultURL,"| ")[1]
				downloadableRelease.Release=release
				downloadableRelease.Resource=resource
				downloadableRelease.Release.FileExtension=".jar"
				downloadableRelease.Release.FileName="primitiveExternalSupport.jar"
			}
		} else {
			if err != nil {
				return DownloadableRelease{}, errors.New("this release is hosted externally and no valid download URLs have been suggested, please, suggest a download URL for this release here: https://grifpkg.com/suggest/"+resource.Id+"/")
			}
		}
	} else {

		// gets downloadable release
		request, err := request("resource/release/download/", map[string]string{
			"release": releaseId,
		})
		if err != nil {
			return DownloadableRelease{}, err
		}
		downloadableRelease = DownloadableRelease{}
		err = json.NewDecoder(request).Decode(&downloadableRelease)
		if err != nil {
			return DownloadableRelease{}, err
		}

	}

	// download
	finalPath, separator, basepath, err := downloadFile(downloadableRelease,projectConfig)
	if err != nil {
		return DownloadableRelease{}, err
	}

	// unzip
	if downloadableRelease.Release.FileExtension=="zip" {
		err := handleZip(finalPath, separator, basepath, downloadableRelease, projectConfig)
		if err != nil {
			return DownloadableRelease{}, err
		}
	}

	return downloadableRelease, err
}

func UpdateReleases(project config.Project) (updated int, skipped int, upToDate int){
	updated = 0
	skipped = 0
	upToDate = 0
	// (^) indicates this resource is expecting new releases
	for identifier, resource := range project.Dependencies {
		actualCurrentVersion := resource.Version
		if strings.HasPrefix(resource.Version,"^") {
			_, i := utf8.DecodeRuneInString(resource.Version)
			actualCurrentVersion = resource.Version[i:]

			latestRelease, err := getRelease(Resource{Id: resource.Resource}, "")
			if err==nil {
				if latestRelease.Version!=actualCurrentVersion {
					versionHash := md5.New()
					versionHash.Write([]byte(latestRelease.Id+latestRelease.Version))

					resourceHash := md5.New()
					resourceHash.Write([]byte(identifier+resource.Resource))

					fmt.Fprintf(color.Output, "%s Updating %s: %s → %s\n", color.HiGreenString("i"), color.CyanString(identifier), color.CyanString(latestRelease.Version), color.CyanString(actualCurrentVersion))
					_, err := DownloadResource(Resource{Id: resource.Resource}, latestRelease.Version, project, latestRelease.Id)
					if err != nil {
						skipped++
						fmt.Fprintf(color.Output, "%s Error while updating %s: %s\n", color.HiYellowString("!"), color.CyanString(identifier), color.RedString(err.Error()))
					}
					project.Dependencies[identifier] = config.DependencyIdentifier{Version: "^"+latestRelease.Version, Resource: resource.Resource, Release: latestRelease.Id, Hash: []string{hex.EncodeToString(resourceHash.Sum(nil)),hex.EncodeToString(versionHash.Sum(nil))}}
					updated++
				} else {
					fmt.Fprintf(color.Output, "%s Updating %s: already up-to-date, %s\n", color.HiGreenString("i"), color.CyanString(identifier), color.CyanString(actualCurrentVersion))
					upToDate++
				}
			} else {
				skipped++
				fmt.Fprintf(color.Output, "%s Error while updating %s: %s\n", color.HiYellowString("!"), color.CyanString(identifier), color.RedString(err.Error()))
			}
		} else {
			fmt.Fprintf(color.Output, "%s Updating %s: staying on forced version %s\n", color.HiGreenString("i"), color.CyanString(identifier), color.CyanString(actualCurrentVersion))
			upToDate++
		}
	}
	config.SaveConfig(project)
	return updated, skipped, upToDate
}

func downloadFile(downloadableRelease DownloadableRelease, projectConfig config.Project) (string, string, string, error){
	if downloadableRelease.Url == "" {
		return "", "", "", errors.New("this release is hosted externally. support for externally hosted releases is coming this week")
	}

	resp, err := http.Get(downloadableRelease.Url)
	defer resp.Body.Close()

	path, err := os.Getwd()
	separator := string(os.PathSeparator)
	var addedPath = ""
	if val, ok := projectConfig.InstallPaths[downloadableRelease.Release.FileExtension.(string)]; ok {
		addedPath = strings.ReplaceAll(strings.ReplaceAll(val,"./",""),"/",separator)
	} else {
		addedPath = strings.ReplaceAll(strings.ReplaceAll(projectConfig.InstallPaths["default"],"./",""),"/",separator)
	}
	finalPath := path+separator+addedPath
	_ = os.Mkdir(finalPath, 0700)
	out, err := os.Create(finalPath+downloadableRelease.Release.FileName.(string))
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return finalPath, separator, path+separator, err
}

func request(path string, data map[string]string) (io.Reader,error) {
	var result io.Reader = nil
	var err error = nil
	client := &http.Client{}
	if data != nil {
		postString, _ := json.Marshal(&data)
		req, _ := http.NewRequest("POST","https://api.grifpkg.com/rest/1/"+path, bytes.NewBuffer(postString))
		req.Header.Set("Accept","application/json")
		res, _ := client.Do(req)
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		result = bytes.NewReader(body)
	} else {
		req, _ := http.NewRequest("GET","https://api.grifpkg.com/rest/1/"+path, nil)
		req.Header.Set("Accept","application/json")
		res, _ := client.Do(req)
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		result = bytes.NewReader(body)
	}
	return result, err
}

func handleZip(finalPath string, separator string, basePath string, downloadableRelease DownloadableRelease, projectConfig config.Project) error {
	tempFolder := finalPath+ksuid.New().String()+separator
	_ = os.Mkdir(tempFolder, 0700)
	_, err := unzip(finalPath+downloadableRelease.Release.FileName.(string),tempFolder)
	if err != nil {
		return err
	}
	err = filepath.Walk(tempFolder,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				excluded := false
				for _, excludeSchema := range projectConfig.ExcludeFiles {
					res, _ := regexp.MatchString(excludeSchema, info.Name())
					if res {
						excluded = true
						break
					}
				}
				if !excluded {
					var addedPath = ""
					if val, ok := projectConfig.InstallPaths[filepath.Ext(info.Name())]; ok {
						addedPath = strings.ReplaceAll(strings.ReplaceAll(val,"./",""),"/",separator)
					} else {
						addedPath = strings.ReplaceAll(strings.ReplaceAll(projectConfig.InstallPaths["default"],"./",""),"/",separator)
					}
					err = os.Rename(path, basePath+addedPath+info.Name())
				}
			}
			return nil
		})
	err = os.Remove(finalPath+downloadableRelease.Release.FileName.(string))
	err = os.RemoveAll(tempFolder)
	return err
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
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
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
		outFile.Close()
		rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}