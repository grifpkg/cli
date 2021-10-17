package session

import (
	"encoding/json"
	"github.com/grifpkg/cli/api"
	"time"
)

type LoginRequest struct {
	Id       string      		`json:"id"`
	Hash     string      		`json:"hash"`
	Creation int         		`json:"creation"`
	Expiry  int     `json:"expiry"`
	Session Session `json:"session"`
}

func Start() (loginRequest LoginRequest, err error){
	api.LogOne(api.Progress,"starting login request")
	request, err := api.Request("login/request/start/", make(map[string]interface{},0), nil)
	loginRequest = LoginRequest{}
	err = json.NewDecoder(request).Decode(&loginRequest)
	return loginRequest, err
}

func (loginRequest *LoginRequest) Get() (err error){
	api.LogOne(api.Progress,"updating login request")
	request, err := api.Request("login/request/get/", map[string]interface{}{
		"request": loginRequest.Hash,
	}, nil)
	err = json.NewDecoder(request).Decode(&loginRequest)
	return err
}

func (loginRequest *LoginRequest) Validate(githubToken string) (err error){
	api.LogOne(api.Progress,"validating login request")
	request, err := api.Request("login/request/validate/", map[string]interface{}{
		"request": loginRequest.Hash,
		"token": githubToken,
	}, nil)
	err = json.NewDecoder(request).Decode(&loginRequest)
	return err
}

func (loginRequest *LoginRequest) AwaitUntilValidated() (err error){
	for true {
		err := loginRequest.Get()
		if err!=nil {
			return err
		} else if (loginRequest.Session != Session{}) {
			return nil
		}
		api.LogOne(api.Progress,"awaiting for the request to be validated")
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (loginRequest LoginRequest) GetAuthURL() (url string){
	return "https://api.grifpkg.com/rest/1/login/?lr="+loginRequest.Id
}