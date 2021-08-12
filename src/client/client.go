package client

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grifpkg/cli/config"
	"github.com/segmentio/ksuid"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func DownloadResource(resource Resource, version string, projectConfig config.Project) (DownloadableRelease, error) {
	var err error = nil

	// gets release
	release, err := getRelease(resource,version)

	// gets downloadable release
	request, err := request("resource/release/download/", map[string]string{
		"release": release.Id,
	})
	downloadableRelease := DownloadableRelease{}
	err = json.NewDecoder(request).Decode(&downloadableRelease)

	// download
	finalPath, separator, basepath, err := downloadFile(downloadableRelease,projectConfig)
	if err != nil {
		return DownloadableRelease{}, err
	}

	// unzip
	if downloadableRelease.Release.FileExtension=="zip" {
		handleZip(finalPath, separator, basepath,downloadableRelease,projectConfig)
	}

	return downloadableRelease, err
}

func downloadFile(downloadableRelease DownloadableRelease, projectConfig config.Project) (string, string, string, error){
	resp, err := http.Get(downloadableRelease.Url)
	defer resp.Body.Close()

	path, err := os.Getwd()
	separator := string(os.PathSeparator)
	var addedPath = ""
	if val, ok := projectConfig.InstallPaths[downloadableRelease.Release.FileExtension]; ok {
		addedPath = strings.ReplaceAll(strings.ReplaceAll(val,"./",""),"/",separator)
	} else {
		addedPath = strings.ReplaceAll(strings.ReplaceAll(projectConfig.InstallPaths["default"],"./",""),"/",separator)
	}
	finalPath := path+separator+addedPath
	_ = os.Mkdir(finalPath, 0700)
	out, err := os.Create(finalPath+downloadableRelease.Release.FileName)
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
	_, err := unzip(finalPath+downloadableRelease.Release.FileName,tempFolder)
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
	err = os.Remove(finalPath+downloadableRelease.Release.FileName)
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