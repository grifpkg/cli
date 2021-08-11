package client

import (
	"./obj"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func GetResources(name string, author string, service int) ([]obj.Resource, error) {
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
	resourceList := make([]obj.Resource,0)
	err = json.NewDecoder(request).Decode(&resourceList)
	return resourceList, err
}

func getRelease(resource obj.Resource, version string) (obj.Release, error) {
	var err error = nil
	var data map[string]string = make(map[string]string)
	data["resource"] = resource.Id
	if version!="" {
		data["version"] = version
	}
	request, err := request("resource/release/get/", data)
	release := obj.Release{}
	err = json.NewDecoder(request).Decode(&release)
	return release, err
}

func GetReleases(resource obj.Resource) ([]obj.Release, error) {
	var err error = nil
	request, err := request("resource/release/list/", map[string]string{
		"resource":resource.Id,
	})

	releaseList := make([]obj.Release,0)
	err = json.NewDecoder(request).Decode(&releaseList)
	return releaseList, err
}

func DownloadResource(resource obj.Resource, version string) (obj.DownloadableRelease, error) {
	var err error = nil

	// gets release
	release, err := getRelease(resource,version)

	// gets downloadable release
	request, err := request("resource/release/download/", map[string]string{
		"release": release.Id,
	})
	downloadableRelease := obj.DownloadableRelease{}
	err = json.NewDecoder(request).Decode(&downloadableRelease)

	// download
	resp, err := http.Get(downloadableRelease.Url)
	defer resp.Body.Close()

	path, err := os.Getwd()
	separator := string(os.PathSeparator)
	out, err := os.Create(path+separator+downloadableRelease.Release.FileName)
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
