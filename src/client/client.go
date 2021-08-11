package client

import (
	"../config"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	resp, err := http.Get(downloadableRelease.Url)
	defer resp.Body.Close()

	path, err := os.Getwd()
	separator := string(os.PathSeparator)
	addedPath := strings.ReplaceAll(strings.ReplaceAll(projectConfig.InstallPaths["default"],"./",""),"/",separator)
	_ = os.Mkdir(path+separator+addedPath, 0700)
	out, err := os.Create(path+separator+addedPath+downloadableRelease.Release.FileName)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return downloadableRelease, err
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
