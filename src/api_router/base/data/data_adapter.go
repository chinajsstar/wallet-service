package data

import (
	"bastionpay_api/api"
	"strings"
)

// method
func (method *SrvMethod)FromPath(path string)  {
	path = strings.TrimLeft(path, "/")
	path = strings.TrimRight(path, "/")
	paths := strings.Split(path, "/")
	for i := 0; i < len(paths); i++ {
		if i == 1 {
			method.Version = paths[i]
		}else if i == 2{
			method.Srv = paths[i]
		} else if i >= 3{
			if method.Function != "" {
				method.Function += "."
			}
			method.Function += paths[i]
		}
	}
}

func ApiMethodFromPath(method *api.UserMethod, path string)  {
	path = strings.TrimLeft(path, "/")
	path = strings.TrimRight(path, "/")
	paths := strings.Split(path, "/")
	for i := 0; i < len(paths); i++ {
		if i == 1 {
			method.Version = paths[i]
		}else if i == 2{
			method.Srv = paths[i]
		} else if i >= 3{
			if method.Function != "" {
				method.Function += "."
			}
			method.Function += paths[i]
		}
	}
}

func (method *SrvMethod)FromApiMethod(um *api.UserMethod)  {
	method.Version = um.Version
	method.Srv = um.Srv
	method.Function = um.Function
}

func (method *SrvMethod)ToApiMethod(um *api.UserMethod)  {
	um.Version = method.Version
	um.Srv = method.Srv
	um.Function = method.Function
}

// data
func (data *SrvData)FromApiData(ud *api.UserData)  {
	data.UserKey = ud.UserKey
	data.SubUserKey = ""
	data.Message = ud.Message
	data.Signature = ud.Signature
}

func (data *SrvData)ToApiData(ud *api.UserData)  {
	ud.UserKey = data.UserKey
	ud.Message = data.Message
	ud.Signature = data.Signature
}

// response
func (response *SrvResponse)ToApiResponse(ur *api.UserResponseData)  {
	ur.Err = response.Err
	ur.ErrMsg = response.ErrMsg
	response.Value.ToApiData(&ur.Value)
}