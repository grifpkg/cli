package client

type Release struct {
	Id                           string      `json:"id"`
	Service                      int         `json:"service"`
	ReleaseId                    string      `json:"releaseId"`
	Version                      string      `json:"version"`
	Creation                     int         `json:"creation"`
	FileName                     string      `json:"fileName"`
	FileExtension                string      `json:"fileExtension"`
	FileSize                     int         `json:"fileSize"`
	ManifestName                 string      `json:"manifestName"`
	ManifestAuthors              []string    `json:"manifestAuthors"`
	ManifestVersion              string      `json:"manifestVersion"`
	ManifestMain                 string      `json:"manifestMain"`
	ManifestVersionAPI           string      `json:"manifestVersionAPI"`
	ManifestDependencies         []string 	 `json:"manifestDependencies"`
	ManifestOptionalDependencies []string	 `json:"manifestOptionalDependencies"`
}
