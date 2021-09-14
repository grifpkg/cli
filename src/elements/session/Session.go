package session

import (
	"encoding/json"
	"github.com/fatih/color"
	"github.com/grifpkg/cli/api"
	"github.com/zalando/go-keyring"
	"log"
)

type Session struct {
	Id        string 		`json:"id"`
	Hash      string 		`json:"hash"`
	UserAgent string 		`json:"userAgent"`
	Creation  int    		`json:"creation"`
	Expiry    int    		`json:"expiry"`
	City      interface{}	`json:"city"`
	Country   interface{}	`json:"country"`
}

func (session *Session) Close() (err error){
	if len(session.Hash)>0 {
		api.LogOne(api.Progress,"logging out")
		request, err := api.Request("session/close/", map[string]string{}, GetHash())
		if err!=nil {
			err = nil // ignore error
		} else {
			err = json.NewDecoder(request).Decode(&session)
			err = nil // ignore error
		}
	}
	api.LogOne(api.Progress,"removing session from the keychain")
	err = keyring.Delete(api.KeychainService, api.KeychainHash)
	if err != nil {
		return err
	}
	return nil
}

func GetHash() (hash interface{}){
	currentSession, err := Get()
	if err!=nil || currentSession.Hash=="" {
		return nil
	} else {
		return currentSession.Hash
	}
}

func Get() (session Session, err error){
	api.LogOne(api.Progress,"loading grifpkg's session hash from the keychain")
	secret, err := keyring.Get(api.KeychainService, api.KeychainHash)
	if err != nil {
		api.LogOne(api.Progress,"no sessions found, requesting login")
		login, err := Start()
		api.Log(api.Info, []api.Message{
			{
				Value: "please, login to your grifpkg account by using this link:",
				Color: nil,
			},
			{
				Value: login.GetAuthURL(),
				Color: color.New(color.FgHiBlue),
			},
		})
		if err!=nil {
			return Session{}, err
		}
		err = login.AwaitUntilValidated()
		if err!=nil {
			return Session{}, err
		}
		api.LogOne(api.Progress,"saving newly generated session into the keychain")
		session, err := json.Marshal(login.Session)
		if err!=nil {
			return Session{}, err
		}
		err = keyring.Set(api.KeychainService, api.KeychainHash, string(session))
		if err != nil {
			log.Fatal(err)
		}
		return login.Session, nil
	} else {
		finalSession := Session{}
		err := json.Unmarshal([]byte(secret), &finalSession)
		if err!=nil {
			return Session{}, err
		}
		return finalSession, nil
	}
}