package obj

type DownloadableRelease struct {
	Url     string `json:"url"`
	Release Release `json:"release"`
	Resource Resource `json:"resource"`
}
