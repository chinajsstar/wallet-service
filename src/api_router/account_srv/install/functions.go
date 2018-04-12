package install

import (
	"fmt"
	"io/ioutil"
	"../user"
	"strconv"
	"../../base/utils"
)

func AddUser() (*user.ReqUserCreate, error) {
	var input string

	uc := &user.ReqUserCreate{}

	fmt.Println("输入用户名: ")
	fmt.Scanln(&input)
	uc.UserName = input

	fmt.Println("输入电话: ")
	fmt.Scanln(&input)
	uc.Phone = input

	fmt.Println("输入邮箱: ")
	fmt.Scanln(&input)
	uc.Email = input

	fmt.Println("输入权限: ")
	fmt.Scanln(&input)
	uc.Level, _ = strconv.Atoi(input)

	fmt.Println("输入公钥文件路径: ")
	fmt.Scanln(&input)
	pubPath := input
	pubKey, err := ioutil.ReadFile(pubPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	uc.PublicKey = string(pubKey)

	var pw1, pw2 string
	for ; ; {
		fmt.Println("输入密码6位或以上: ")
		fmt.Scanln(&input)
		pw1 = input

		fmt.Println("再次输入密码: ")
		fmt.Scanln(&input)
		pw2 = input
		if pw1 == pw2 && len(pw1) > 5{
			break
		}
		pw1 = ""
		pw2 = ""
	}

	uc.Password = utils.GetMd5Text(pw1)

	fmt.Println("输入国家: ")
	fmt.Scanln(&input)
	uc.Country = input

	fmt.Println("输入语言: ")
	fmt.Scanln(&input)
	uc.Language = input

	fmt.Println("输入时区: ")
	fmt.Scanln(&input)
	uc.TimeZone = input

	fmt.Println("输入google验证: ")
	fmt.Scanln(&input)
	uc.GoogleAuth = input

	return uc, nil
}

func LoginUser() (*user.ReqUserLogin, error) {
	var input string

	uc := &user.ReqUserLogin{}

	fmt.Println("用户名，电话，邮箱填一个: ")
	fmt.Println("输入用户名: ")
	fmt.Scanln(&input)
	uc.UserName = input

	fmt.Println("输入电话: ")
	fmt.Scanln(&input)
	uc.Phone = input

	fmt.Println("输入邮箱: ")
	fmt.Scanln(&input)
	uc.Email = input

	fmt.Println("输入密码: ")
	fmt.Scanln(&input)
	pw1 := input

	uc.Password = utils.GetMd5Text(pw1)

	return uc, nil
}

func ListUsers() (*user.ReqUserList, error) {
	var err error
	var input string

	uc := &user.ReqUserList{}

	fmt.Println("输入上次最小id，默认填-1: ")
	fmt.Scanln(&input)
	uc.Id, err = strconv.Atoi(input)

	return uc, err
}