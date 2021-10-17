package project

import (
	"encoding/json"
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/author"
)

type Resource struct {
	Id         string      		`json:"id"`
	Service    int         		`json:"service"`
	ResourceId string	 		`json:"resourceId"`
	Paid       bool        		`json:"paid"`
	Name      string        `json:"name"`
	Author    author.Author `json:"author"`
	Downloads int           `json:"downloads"`
	Ratings     int 			`json:"ratings"`
	Rating      float64	    	`json:"rating"`
	Description string      	`json:"description"`
}


func QueryResources(name string, author interface{}, service interface{}) ([]Resource, error) {
	api.LogOne(api.Progress, "querying resources")
	var err error = nil
	var data = make(map[string]interface{}, 0)
	data["name"]=name
	if author!=nil {
		data["author"]=author.(string)
	}
	if service!=nil {
		data["service"]= "0" // TODO fix service
	}
	request, err := api.Request("resource/query/", data, nil)
	resourceList := make([]Resource,0)
	err = json.NewDecoder(request).Decode(&resourceList)
	return resourceList, err
}

func (resource Resource) GetReleases() (releases []Release, err error){
	api.LogOne(api.Progress, "getting releases")
	request, err := api.Request("resource/release/list/", map[string]interface{}{
		"resource":resource.Id,
	}, nil)

	releaseList := make([]Release,0)
	err = json.NewDecoder(request).Decode(&releaseList)
	for _, release := range releaseList {
		release.Parent=resource
	}
	return releaseList, err
}

func (resource Resource) GetRelease(version interface{}, id interface{}) (releases Release, err error){
	api.LogOne(api.Progress, "getting release")
	var data map[string]interface{} = make(map[string]interface{})
	data["resource"] = resource.Id
	if version!=nil {
		data["version"] = version.(string)
	}
	if id!=nil {
		data["release"] = id.(string)
	}
	request, err := api.Request("resource/release/get/", data, nil)
	release := Release{}
	release.Parent=resource
	err = json.NewDecoder(request).Decode(&release)
	return release, err
}