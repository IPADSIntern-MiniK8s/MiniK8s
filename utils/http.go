package utils

import (
	"bytes"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func SendJsonObject(method string, jsonObject []byte, url string) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonObject))

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("content-type", "application/json")
	//request.Header.Set("accept", "application/json, text/plain, */*")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		//log.Fatal("client.Do err:")
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		//fmt.Println(resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			fmt.Println(body)
		}
	}
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
