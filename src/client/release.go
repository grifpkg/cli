package client

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
}
