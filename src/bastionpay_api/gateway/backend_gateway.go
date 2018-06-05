package gateway

import (
	"encoding/json"
	"bastionpay_api/api"
	"bastionpay_api/apibackend"
	"fmt"
)

func RunApiTest(path string, req interface{}, ack interface{}) (*api.Error) {
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

	messageByte, apiErr := outputApiTest(path, reqByte)
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

func RunUser(path string, subUserKey string, req interface{}, ack interface{}) (*api.Error) {
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

	userMessage := apibackend.UserMessage{}
	userMessage.SubUserKey = subUserKey
	userMessage.Message = string(reqByte)

	var reqUserByte []byte
	reqUserByte, err = json.Marshal(userMessage)
	if err != nil {
		return api.NewError(1, err.Error())
	}

	messageByte, apiErr := outputApi(path, reqUserByte)
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

func RunAdmin(path string, subUserKey string, req interface{}, ack interface{}) (*api.Error) {
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

	adminMessage := apibackend.AdminMessage{}
	adminMessage.SubUserKey = subUserKey
	adminMessage.Message = string(reqByte)

	var reqUserByte []byte
	reqUserByte, err = json.Marshal(adminMessage)
	if err != nil {
		return api.NewError(1, err.Error())
	}

	messageByte, apiErr := outputApi(path, reqUserByte)
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

func outputApiTest(path string, message []byte) ([]byte, *api.Error) {
	userData := api.UserData{}
	userData.UserKey = setting.User.UserKey
	userData.Message = string(message)

	userDataByte, err := json.Marshal(userData)
	if err != nil {
		return nil, api.NewError(1, err.Error())
	}
	fmt.Println("request: ", string(userDataByte))

	httpPath := setting.BastionPay.Url + path
	resByte, err := HttpPost(httpPath, userDataByte)
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

	messageByte := []byte(res.Value.Message)
	return messageByte, nil
}