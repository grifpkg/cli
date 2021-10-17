package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

func Request(path string, data map[string]interface{}, hash interface{}) (response io.Reader,err error) {
	var result io.Reader = nil
	var body []byte
	err = nil
	client := &http.Client{}
	if data != nil && len(data) > 0 {
		postString, err := json.Marshal(&data)
		if err!=nil {
			return nil, err
		}
		req, err := http.NewRequest("POST","https://api.grifpkg.com/rest/1/"+path, bytes.NewBuffer(postString))
		if err!=nil {
			return nil, err
		}
		req.Header.Set("Accept","application/json")
		req.Header.Set("User-Agent","grifpkg/cli")
		if hash!=nil {
			req.Header.Set("Authorization","Bearer "+hash.(string))
		}
		res, err := client.Do(req)
		if err!=nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
		}(res.Body)
		body, err = ioutil.ReadAll(res.Body)
		if err!=nil {
			return nil, err
		}
		result = bytes.NewReader(body)
	} else {
		req, err := http.NewRequest("GET","https://api.grifpkg.com/rest/1/"+path, nil)
		if err!=nil {
			return nil, err
		}
		req.Header.Set("Accept","application/json")
		req.Header.Set("User-Agent","grifpkg/cli")
		if hash!=nil {
			req.Header.Set("Authorization","Bearer "+hash.(string))
		}
		res, err := client.Do(req)
		defer func(Body io.ReadCloser) {
			err = Body.Close()
		}(res.Body)
		body, err = ioutil.ReadAll(res.Body)
		if err!=nil {
			return nil, err
		}
		result = bytes.NewReader(body)
	}
	type Error struct {
		Error interface{}	`json:"error"`
	}
	var errorInfo Error = Error{}
	errorReader := bytes.NewReader(body)
	_ = json.NewDecoder(errorReader).Decode(&errorInfo)
	if errorInfo.Error!=nil {
		return result,errors.New(errorInfo.Error.(string))
	}
	return result, err
}