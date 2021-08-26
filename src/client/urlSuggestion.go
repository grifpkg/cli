package client

type UrlSuggestion struct {
	Id          string 			`json:"id"`
	SuggestedBy Account			`json:"suggestedBy"`
	ApprovedBy 	interface{} 	`json:"approvedBy"`
	Url        	string      	`json:"url"`
	UrlSchema  	string      	`json:"urlSchema"`
	JsonURL    	string      	`json:"jsonURL"`
	JsonPath   	string      	`json:"jsonPath"`
	FileName   	string      	`json:"fileName"`
	Creation   	int         	`json:"creation"`
	Uses       	int         	`json:"uses"`
}
