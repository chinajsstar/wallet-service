package nethelper

import (
	"log"
	"bytes"
	"io/ioutil"
	"net/http"
)

// Do a post to Http server
func CallToHttpServer(addr string, path string, body string, res *string) error {
	url := addr + path
	contentType := "application/json;charset=utf-8"

	b := []byte(body)
	b2 := bytes.NewBuffer(b)

	resp, err := http.Post(url, contentType, b2)
	if err != nil {
		log.Println("#CallToHttpServer Post failed:", err)
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("#CallToHttpServer Read failed:", err)
		return err
	}

	*res = string(content)
	return nil
}