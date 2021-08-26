package client

import "strings"

type DownloadableRelease struct {
	Url     string `json:"url"`
	Release Release `json:"release"`
	Resource Resource `json:"resource"`
}

func DownloadableReleaseFromSuggestion(suggestion UrlSuggestion, release Release, resource Resource) DownloadableRelease {
	return DownloadableRelease{suggestion.Url,Release{release.Id,release.Service,release.ReleaseId,release.Version,release.Creation,suggestion.FileName,extensionFromFileName(suggestion.FileName),0,nil,nil,nil,nil,nil,nil,nil,true},resource}
}

func extensionFromFileName(fileName string) interface{} {
	parts := strings.Split(fileName,".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return nil
}
