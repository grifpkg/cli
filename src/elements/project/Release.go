package project

import (
	"encoding/json"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/session"
	"github.com/grifpkg/cli/elements/urlSuggestion"
	"io"
	"strings"
)

type Release struct {
	Id                           	string      	`json:"id"`
	Service                      	int         	`json:"service"`
	ReleaseId                   	string    		`json:"releaseId"`
	Version                      	string      	`json:"version"`
	Creation                   	  	int         	`json:"creation"`
	FileName                   		interface{}     `json:"fileName"`
	FileExtension                	interface{}     `json:"fileExtension"`
	FileSize                     	interface{}     `json:"fileSize"`
	ManifestName                 	interface{}     `json:"manifestName"`
	ManifestAuthors              	interface{}    	`json:"manifestAuthors"`
	ManifestVersion              	interface{}     `json:"manifestVersion"`
	ManifestMain                 	interface{}     `json:"manifestMain"`
	ManifestVersionAPI           	interface{}     `json:"manifestVersionAPI"`
	ManifestDependencies         	interface{} 	`json:"manifestDependencies"`
	ManifestOptionalDependencies 	interface{}	 	`json:"manifestOptionalDependencies"`
	HasSuggestions 					interface{}	 	`json:"hasSuggestions"`
	Parent							interface{}
}

func (release Release) GetResolvedDependencies() (dependencies map[string]Resource, err error){
	api.LogOne(api.Progress,"resolving dependencies")
	var resolvedDependencies map[string]Resource = map[string]Resource{}
	if release.ManifestDependencies != nil {
		for _, dependency := range release.ManifestDependencies.([]interface{}) {
			dependencyName := dependency.(string)
			resources, err := QueryResources(dependencyName, nil, nil)
			if err != nil {
				return nil, err
			}
			if len(resources)<=0 {
				return nil, errors.New("no suitable resource found while resolving the dependency '"+dependencyName+"'")
			}
			resolvedDependencies[dependencyName] = resources[0]
		}
	}
	return resolvedDependencies, nil
}

func (release Release) ListSuggestions() (suggestions []urlSuggestion.UrlSuggestion, err error){
	api.LogOne(api.Progress, "querying url suggestions")
	request, err := api.Request("resource/release/suggestion/list/", map[string]interface{}{
		"release": release.Id,
	}, nil)
	if err!=nil{
		return []urlSuggestion.UrlSuggestion{}, nil
	}
	suggestions = make([]urlSuggestion.UrlSuggestion,0)
	err = json.NewDecoder(request).Decode(&suggestions)
	return suggestions, err
}

func (_ Release) Get(id string) (release Release, err error){
	request, err := api.Request("resource/release/get/", map[string]interface{}{
		"release": id,
	}, nil)
	if err != nil {
		return Release{}, err
	}
	err = json.NewDecoder(request).Decode(&release)
	if err != nil {
		return Release{}, err
	}
	return release, nil
}

func (release Release) GetDownloadable(suggestionIdFallback interface{}) (downloadableRelease DownloadableRelease, err error) {
	api.LogOne(api.Progress, "requesting downloadable release")
	downloadableRelease = DownloadableRelease{}
	if release.HasSuggestions==nil {
		updatedRelease, err := release.Get(release.Id)
		if err != nil {
			return DownloadableRelease{}, err
		}
		release.HasSuggestions=updatedRelease.HasSuggestions
	}
	if release.HasSuggestions!=nil {
		if release.HasSuggestions == true {
			suggestionList, err := release.ListSuggestions()
			if err!=nil {
				return DownloadableRelease{}, nil
			}
			var selectedSuggestion = urlSuggestion.UrlSuggestion{}

			var optionList []string
			for _, suggestion := range suggestionList {
				if suggestion.Id==suggestionIdFallback {
					api.LogOne(api.Progress, "a predefined url suggestion was found on the suggestion list, building suggestion based on that one")
					selectedSuggestion=suggestion
					break
				}
				optionList = append(optionList, suggestion.Url + " ("+suggestion.Id+")")
			}

			if (selectedSuggestion == urlSuggestion.UrlSuggestion{}) {
				api.LogOne(api.Progress, "there was no predefined url suggestion for this release, asking for one")

				// no suggestion match a fallback suggestion, therefore we ask for a new one
				optionList = append(optionList, "suggest another URL")
				optionList = append(optionList, "skip")

				answers := struct {
					Selection string
				}{}

				err := api.Ask([]*survey.Question{
					{
						Name: "selection",
						Prompt: &survey.Select{
							Message: "this resource is hosted externally, here is a list of suggested download URLs, note these URLs are NOT verified and may not point to the desired target, UAYOR",
							Options: optionList,
						},
					},
				}, &answers)
				if err != nil {
					return DownloadableRelease{}, err
				}

				if strings.HasPrefix(answers.Selection,"suggest another URL"){
					return DownloadableRelease{}, errors.New("suggest a download URL here: https://grifpkg.com/suggest/"+release.Parent.(Resource).Id)
				} else if strings.HasPrefix(answers.Selection,"skip") {
					return DownloadableRelease{}, errors.New("skipped release")
				} else {
					api.LogOne(api.Progress, "matching selected suggestion to a suggestion object")
					spacedParts := strings.Split(answers.Selection," ")
					var resultSuggestionId = ""
					resultSuggestionId = strings.ReplaceAll(spacedParts[len(spacedParts)-1],"(","")
					resultSuggestionId = strings.ReplaceAll(resultSuggestionId,")","")
					selectedSuggestion = urlSuggestion.UrlSuggestion{}
					for _, suggestion := range suggestionList {
						if suggestion.Id==resultSuggestionId {
							selectedSuggestion=suggestion
							break
						}
					}
					if (selectedSuggestion == urlSuggestion.UrlSuggestion{}) {
						return DownloadableRelease{}, errors.New("url suggestion selection mismatch")
					}
				}
			}

			// notify usage
			err = selectedSuggestion.Use()
			if err != nil {
				return DownloadableRelease{}, errors.New("couldn't register url suggestion usage")
			}

			api.LogOne(api.Progress, "building url suggestion")
			downloadableRelease = DownloadableReleaseFromSuggestion(selectedSuggestion,release,release.Parent)
		} else {
			return DownloadableRelease{}, errors.New("this release is hosted externally and no valid download URLs have been suggested, please, suggest a download URL for this release here: https://grifpkg.com/suggest/"+release.Parent.(Resource).Id)
		}
	} else {

		// gets downloadable release

		var request interface{} = nil
		if release.Parent != nil && release.Parent.(Resource).Paid {
			request, err = api.Request("resource/release/download/", map[string]interface{}{
				"release": release.Id,
			}, session.GetHash())
		} else {
			request, err = api.Request("resource/release/download/", map[string]interface{}{
				"release": release.Id,
			}, nil)
		}
		if err != nil {
			return DownloadableRelease{}, err
		}
		downloadableRelease = DownloadableRelease{}
		err = json.NewDecoder(request.(io.Reader)).Decode(&downloadableRelease)
		if err != nil {
			return DownloadableRelease{}, err
		}

	}

	return downloadableRelease, err
}