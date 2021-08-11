package client

type Author struct {
	Service  int         `json:"service"`
	Id       interface{} `json:"id"`
	Username string      `json:"username"`
	AuthorId interface{} `json:"authorId"`
}
