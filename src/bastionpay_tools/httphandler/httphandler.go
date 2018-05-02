package httphandler

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bytes"
	"mime/multipart"
	"net/http"
	"errors"
	"strings"
)

func UploadFile(filePath string, targetUrl string) (string, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	pps := strings.Split(filePath, "/")
	if len(pps) == 0 {
		return "", errors.New("url error")
	}

	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", pps[len(pps)-1])
	if err != nil {
		return "", err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	//io copy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return "", err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println(resp.Status)
	if resp.StatusCode != http.StatusOK {
		return string(respBody), errors.New(resp.Status)
	}

	return string(respBody), nil
}

func DownloadFile(filepath string, url string) (error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		return err
	}

	return nil
}