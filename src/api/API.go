package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func Request(path string, data map[string]string, hash interface{}) (response io.Reader,err error) {
	var result io.Reader = nil
	err = nil
	client := &http.Client{}
	if data != nil {
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
		res, err := client.Do(req)
		if err!=nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
		}(res.Body)
		body, err := ioutil.ReadAll(res.Body)
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
		res, err := client.Do(req)
		defer func(Body io.ReadCloser) {
			err = Body.Close()
		}(res.Body)
		body, err := ioutil.ReadAll(res.Body)
		if err!=nil {
			return nil, err
		}
		result = bytes.NewReader(body)
	}
	return result, err
}