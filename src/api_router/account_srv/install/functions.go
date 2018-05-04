package install

import (
	"fmt"
	"api_router/account_srv/user"
	"strconv"
	"api_router/base/utils"
	"io/ioutil"
)

func AddUser(isinstall bool) (*user.ReqUserCreate, error) {
	var input string

	uc := &user.ReqUserCreate{}

	fmt.Println("输入用户名: ")
	input = ""
	input = utils.ScanLine()
	uc.UserName = input

	fmt.Println("输入电话: ")
	input = ""
	input = utils.ScanLine()
	uc.Phone = input

	fmt.Println("输入邮箱: ")
	input = ""
	input = utils.ScanLine()
	uc.Email = input

	if isinstall {
		uc.UserClass = 2
		uc.Level = 200
	}else{
		//0:普通用户 1:热钱包; 100:管理员
		fmt.Println("输入用户类型: 0:普通用户 1:热钱包; 2:管理员")
		input = ""
		input = utils.ScanLine()
		uc.UserClass, _ = strconv.Atoi(input)
		if uc.UserClass == 100 {
			fmt.Println("输入权限: 100：普通管理员 200：创世管理员")
			fmt.Scanln(&input)
			uc.Level, _ = strconv.Atoi(input)
		}else{
			uc.Level = 0
		}
	}

	var pw1, pw2 string
	for ; ; {
		fmt.Println("输入密码6位或以上: ")
		input = ""
		input = utils.ScanLine()
		pw1 = input

		fmt.Println("再次输入密码: ")
		input = ""
		input = utils.ScanLine()
		pw2 = input
		if pw1 == pw2 && len(pw1) > 5{
			break
		}
		pw1 = ""
		pw2 = ""
	}

	uc.Password = utils.GetMd5Text(pw1)

	fmt.Println("输入国家: ")
	input = ""
	input = utils.ScanLine()
	uc.Country = input

	fmt.Println("输入语言: ")
	input = ""
	input = utils.ScanLine()
	uc.Language = input

	fmt.Println("输入时区: ")
	input = ""
	input = utils.ScanLine()
	uc.TimeZone, _ = strconv.Atoi(input)

	fmt.Println("输入google验证: ")
	input = ""
	input = utils.ScanLine()
	uc.GoogleAuth = input

	return uc, nil
}

func UpdateKey()(*user.ReqUserUpdateKey, error){
	var input string

	userUpdateKey := &user.ReqUserUpdateKey{}

	fmt.Println("请输入注册的user_key: ")
	input = ""
	input = utils.ScanLine()
	userUpdateKey.UserKey = input

	fmt.Println("请输入公钥路径: ")
	input = ""
	input = utils.ScanLine()
	pubKey, err := ioutil.ReadFile(input)
	if err != nil {
		return nil, err
	}
	userUpdateKey.PublicKey = string(pubKey)

	fmt.Println("请输入回调地址: ")
	input = ""
	input = utils.ScanLine()
	userUpdateKey.CallbackUrl = input

	return userUpdateKey, nil
}

func LoginUser() (*user.ReqUserLogin, error) {
	var input string

	uc := &user.ReqUserLogin{}

	fmt.Println("用户名，电话，邮箱填一个: ")
	fmt.Println("输入用户名: ")
	input = ""
	input = utils.ScanLine()
	uc.UserName = input

	fmt.Println("输入电话: ")
	input = ""
	input = utils.ScanLine()
	uc.Phone = input

	fmt.Println("输入邮箱: ")
	input = ""
	input = utils.ScanLine()
	uc.Email = input

	fmt.Println("输入密码: ")
	input = ""
	input = utils.ScanLine()
	pw := input

	uc.Password = utils.GetMd5Text(pw)

	return uc, nil
}