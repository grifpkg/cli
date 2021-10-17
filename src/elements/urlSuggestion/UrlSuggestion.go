package urlSuggestion

import (
	"github.com/grifpkg/cli/api"
	"github.com/grifpkg/cli/elements/account"
)

type UrlSuggestion struct {
	Id          string 			`json:"id"`
	SuggestedBy account.Account	`json:"suggestedBy"`
	ApprovedBy 	interface{} 	`json:"approvedBy"`
	Url        	string      	`json:"url"`
	UrlSchema  	string      	`json:"urlSchema"`
	JsonURL    	string      	`json:"jsonURL"`
	JsonPath   	string      	`json:"jsonPath"`
	FileName   	string      	`json:"fileName"`
	Creation   	int         	`json:"creation"`
	Uses       	int         	`json:"uses"`
}

func (suggestion UrlSuggestion) Use() error{
	api.LogOne(api.Progress, "registering url suggestion use")
	_, err := api.Request("resource/release/suggestion/use/", map[string]interface{}{
		"suggestion": suggestion.Id,
	}, nil)
	return err
}
