package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
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
		fmt.Println(resp.StatusCode)
	}
}
