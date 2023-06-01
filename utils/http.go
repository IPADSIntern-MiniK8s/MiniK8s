package utils

import (
	"bytes"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func SendJsonObject(method string, jsonObject []byte, url string) bool {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonObject))

	if err != nil {
		log.Error(err)
		return false
	}

	request.Header.Set("content-type", "application/json")
	//request.Header.Set("accept", "application/json, text/plain, */*")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		//log.Fatal("client.Do err:")
		log.Error(err)
		return false
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		log.Error(err)
		return false
	}
	resp.Body.Close()
	//fmt.Println(resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		fmt.Println(body)
	}
	return true
}

func SendRequest(method string, str []byte, url string) (string, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(str))
	if err != nil {
		return "", err
	}
	request.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)

	body := &bytes.Buffer{}
	if err != nil {
		log.Error(err)
	} else {
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Error(err)
		}
		resp.Body.Close()
		//fmt.Println(resp.StatusCode)
	}
	return body.String(), err

}

func SendRequestWithHb(method string, str []byte, url string, source string) (string, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(str))
	if err != nil {
		return "", err
	}
	request.Header.Set("content-type", "application/json")
	request.Header.Set("source", source)

	client := &http.Client{}
	resp, err := client.Do(request)

	body := &bytes.Buffer{}
	if err != nil {
		log.Error(err)
	} else {
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Error(err)
		}
		resp.Body.Close()
		//fmt.Println(resp.StatusCode)
	}
	return body.String(), err

}
