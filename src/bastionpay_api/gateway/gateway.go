package gateway

import (
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"crypto"
	"crypto/sha512"
	"bastionpay_api/utils"
	"bastionpay_api/api"
	"time"
	"bytes"
	"net/http"
	"fmt"
)

type (
	Config struct{
		BastionPay struct{
			Url string `json:"url"`
			PubKeyPath string `json:"pubkey_path"`
		} `json:"bastionpay"`
		User struct{
			UserKey 	string `json:"user_key"`
			PubKeyPath 	string `json:"pubkey_path"`
			PrivKeyPath string `json:"privkey_path"`
		} `json:"user"`
	}

	Setting struct {
		BastionPay struct{
			Url 		string
			PubKey 	[]byte
		}
		User struct{
			UserKey 	string
			PubKey 		[]byte
			PrivKey 	[]byte
		}
	}
)

var (
	setting 	Setting
)

func Init(dir, cfgName string) error {
	data, err := ioutil.ReadFile(dir + "/" + cfgName)
	if err != nil {
		return err
	}

	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	setting.BastionPay.Url = config.BastionPay.Url
	setting.BastionPay.PubKey, err = ioutil.ReadFile(dir + "/" + config.BastionPay.PubKeyPath)
	if err != nil {
		return err
	}

	setting.User.UserKey = config.User.UserKey
	setting.User.PubKey, err = ioutil.ReadFile(dir + "/" + config.User.PubKeyPath)
	if err != nil {
		return err
	}
	setting.User.PrivKey, err = ioutil.ReadFile(dir + "/" + config.User.PrivKeyPath)
	if err != nil {
		return err
	}

	return nil
}

func SetBastionPaySetting(url string, pubKey []byte) {
	setting.BastionPay.Url = url
	setting.BastionPay.PubKey = pubKey
}

func SetUserSetting(userKey string, pubKey []byte, privKey []byte) {
	setting.User.UserKey = userKey
	setting.User.PubKey = pubKey
	setting.User.PrivKey = privKey
}

func RunApi(path string, req interface{}, ack interface{}) (*api.Error) {
	var err error
	var reqByte []byte
	if b, ok := req.([]byte); ok {
		reqByte = b
	} else {
		reqByte, err = json.Marshal(req)
		if err != nil {
			return api.NewError(1, err.Error())
		}
	}

	messageByte, apiErr := outputApi(path, reqByte)
	if apiErr != nil {
		return apiErr
	}

	if b, ok := ack.(*[]byte); ok {
		*b = messageByte
	} else {
		err = json.Unmarshal(messageByte, ack)
		if err != nil {
			return api.NewError(1, err.Error())
		}
	}

	return nil
}

func outputApi(path string, message []byte) ([]byte, *api.Error) {
	userData, err := Encryption(message)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}

	userDataByte, err := json.Marshal(userData)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}
	fmt.Println("request: ", string(userDataByte))

	httpPath := setting.BastionPay.Url + path
	resByte, err := httpPost(httpPath, userDataByte)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}

	fmt.Println("response: ", string(resByte))

	res := api.UserResponseData{}
	err = json.Unmarshal(resByte, &res)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}

	if res.Err != 0 {
		return nil, api.NewError(res.Err, res.ErrMsg)
	}

	messageByte, err := Decryption(&res.Value)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}

	return messageByte, nil
}

func Encryption(message []byte) (*api.UserData, error) {
	var (
		err error
		encMessage []byte
		signature []byte
	)

	userData := &api.UserData{}
	userData.UserKey = setting.User.UserKey

	// encrypt
	encMessage, err = utils.RsaEncrypt(message, setting.BastionPay.PubKey, utils.RsaEncodeLimit2048)
	if err != nil {
		return nil, err
	}
	userData.Message = base64.StdEncoding.EncodeToString(encMessage)

	// signature
	hs := sha512.New()
	hs.Write(encMessage)
	hashData := hs.Sum(nil)

	signature, err = utils.RsaSign(crypto.SHA512, hashData, setting.User.PrivKey)
	if err != nil {
		return nil, err
	}
	userData.Signature = base64.StdEncoding.EncodeToString(signature)

	return userData, nil
}

func Decryption(userData *api.UserData) ([]byte, error) {
	var (
		err error
		encMessage 	[]byte
		signature 	[]byte
	)

	encMessage, err = base64.StdEncoding.DecodeString(userData.Message)
	if err != nil {
		return nil, err
	}

	signature, err = base64.StdEncoding.DecodeString(userData.Signature)
	if err != nil {
		return nil, err
	}

	// verify
	hs := sha512.New()
	hs.Write([]byte(encMessage))
	hashData := hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, signature, setting.BastionPay.PubKey)
	if err != nil {
		return nil, err
	}

	// decrypt
	return utils.RsaDecrypt(encMessage, setting.User.PrivKey, utils.RsaDecodeLimit2048)
}

func httpPost(path string, data []byte) ([]byte, error) {
	client := http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		ResponseHeaderTimeout: time.Second * 30,
	}}

	resp, err := client.Post(path, "application/json;charset=utf-8", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}