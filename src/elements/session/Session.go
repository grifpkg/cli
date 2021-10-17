package session

import (
	"encoding/json"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/grifpkg/cli/api"
	"github.com/zalando/go-keyring"
	"log"
	"strings"
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

func (session *Session) Link() (err error){

	answers := struct {
		Service string
	}{}

	var services []string
	services = append(services, "spigotmc")
	err = api.Ask([]*survey.Question{
		{
			Name: "service",
			Prompt: &survey.Select{
				Message: "please, select the service you'd like to link to your grifpkg account",
				Options: services,
			},
		},
	}, &answers)
	if err != nil {
		return err
	}

	// not tried, 0: no error, 1: invalid password, 2: invalid tfa, 3: not tested
	var username string
	var password string
	var tfa interface{} = nil
	for errorCode := 3; errorCode !=0; {

		if errorCode!=2 { // password is not right yet

			usernameAndPassword := struct {
				Username string
				Password string
			}{}

			err := api.Ask([]*survey.Question{
				{
					Name: "username",
					Prompt: &survey.Input{
						Message: "Please, enter your " + answers.Service + " username",
					},
				},
				{
					Name: "password",
					Prompt: &survey.Password{
						Message: "Please, enter your " + answers.Service + " password",
					},
				},
			}, &usernameAndPassword)
			if err != nil {
				return err
			}

			username=usernameAndPassword.Username
			password=usernameAndPassword.Password
			tfa=nil
		} else { // password is right, tfa isn't

			tfaCode := struct {
				Tfa string
			}{}

			err := api.Ask([]*survey.Question{
				{
					Name: "tfa",
					Prompt: &survey.Password{
						Message: "Please, enter your tfa code for " + answers.Service,
					},
				},
			}, &tfaCode)
			if err != nil {
				return err
			}
			tfa=tfaCode.Tfa
		}

		var params map[string]interface{} = map[string]interface{}{}
		params["service"]=answers.Service
		params["username"]=username
		params["password"]=password
		params["tfa"]=tfa

		api.LogOne(api.Progress,"exchanging credentials with "+answers.Service)
		_, err := api.Request("accounts/link/", params, session.Hash)
		if err!=nil {
			if strings.Contains(err.Error(), "two-factor") {
				api.LogOne(api.Info,"your username and password were correct, please, enter your two-factor authentication code")
				errorCode=2
			} else {
				api.LogOne(api.Warn,err.Error())
				errorCode=1
			}
		} else {
			api.LogOne(api.Progress,"valid credentials exchanged with "+answers.Service)
			errorCode=0
		}

	}

	return nil
}

func (session *Session) Close() (err error){
	if len(session.Hash)>0 {
		api.LogOne(api.Progress,"logging out")
		request, err := api.Request("session/close/", map[string]interface{}{}, GetHash())
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
		_ = api.Log(api.Info, []api.Message{
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