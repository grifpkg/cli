package obj

type Resource struct {
	Id         string      	`json:"id"`
	Service    int         	`json:"service"`
	ResourceId string	 	`json:"resourceId"`
	Paid       bool        	`json:"paid"`
	Name       string      	`json:"name"`
	Author     Author 		`json:"author"`
	Downloads   int         `json:"downloads"`
	Ratings     int 		`json:"ratings"`
	Rating      float64	    `json:"rating"`
	Description string      `json:"description"`
}