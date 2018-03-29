package admin

import (
	"fmt"
	"io/ioutil"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"../../user_srv/user"
	"strconv"
)

func AddUser() (string, error) {
	var input string

	uc := user.UserCreate{}

	fmt.Println("输入用户名: ")
	fmt.Scanln(&input)
	uc.UserName = input

	fmt.Println("输入电话: ")
	fmt.Scanln(&input)
	uc.Phone = input

	fmt.Println("输入邮箱: ")
	fmt.Scanln(&input)
	uc.Email = input

	fmt.Println("输入公钥文件路径: ")
	fmt.Scanln(&input)
	pubPath := input
	pubKey, err := ioutil.ReadFile(pubPath)
	if err != nil {
		fmt.Println(err)
		return "",err
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

	// 密码md5
	h := md5.New()
	h.Write([]byte(pw1))
	pw := h.Sum(nil)
	// 密码base64
	uc.Password = base64.StdEncoding.EncodeToString(pw)

	uc.Language = "ch"
	uc.Country = "China"
	uc.TimeZone = "Beijing"
	uc.GoogleAuth = ""

	// 打包
	m, err := json.Marshal(uc)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(m), nil
}

func LoginUser() (string, error) {
	var input string

	uc := user.UserLogin{}

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

	// 密码md5
	h := md5.New()
	h.Write([]byte(pw1))
	pw := h.Sum(nil)
	// 密码base64
	uc.Password = base64.StdEncoding.EncodeToString(pw)

	// 打包
	m, err := json.Marshal(uc)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(m), nil
}

func ListUsers() (string, error) {
	var err error
	var input string

	uc := user.UserList{}

	fmt.Println("输入上次最小id，默认填-1: ")
	fmt.Scanln(&input)
	uc.Id, err = strconv.Atoi(input)

	// 打包
	m, err := json.Marshal(uc)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(m), nil
}