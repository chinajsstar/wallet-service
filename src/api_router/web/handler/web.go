package handler

import (
	"api_router/base/data"
	"encoding/json"
	"net/http"
	"html/template"
	"bastionpay_api/utils"
	"io/ioutil"
	"net/url"
	"strings"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"bastionpay_api/api/v1"
	"api_router/base/nethelper"
	"encoding/base64"
	"crypto/sha512"
	"crypto"
	"errors"
	"api_router/base/config"
	"bastionpay_api/api"
	"bastionpay_api/apigroup"
	"bastionpay_api/apibackend"
)

const (
	httpaddrGateway = "http://127.0.0.1:8082"
)

var web_admin_prikey []byte
var web_admin_pubkey []byte
var web_admin_userkey string

var wallet_server_pubkey []byte
func loadAdministratorRsaKeys(dataDir string) error {
	if dataDir == ""{
		dataDir = "/Users/henly.liu/workspace"
	}

	var err error
	web_admin_prikey, err = ioutil.ReadFile(dataDir+"/private_administrator.pem")
	if err != nil {
		return err
	}

	web_admin_pubkey, err = ioutil.ReadFile(dataDir+"/public_administrator.pem")
	if err != nil {
		return err
	}

	web_admin_userkey = "1c75c668-f1ab-474b-9dae-9ed7950604b4"

	wallet_server_pubkey, err = ioutil.ReadFile(dataDir + "/" + config.BastionPayPublicKey)
	if err != nil {
		return err
	}

	return nil
}

func sendPostData(addr, subUserKey string, rawmessage, version, srv, function string) (*api.UserResponseData, []byte, error) {
	// 用户数据
	var ud api.UserData

	// 构建path
	path := "/"+apibackend.HttpRouterUser
	path += "/"+version
	path += "/"+srv
	path += "/"+function

	// user key
	ud.UserKey = web_admin_userkey

	userParams := apibackend.UserMessage{
		SubUserKey:subUserKey,
		Message:rawmessage,
	}
	bUserParams, err := json.Marshal(userParams)
	if err != nil {
		return nil, nil, err
	}

	message := string(bUserParams)

	// 加密签名数据
	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), wallet_server_pubkey, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return nil, nil, err
	}
	ud.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error){
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, web_admin_prikey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return nil, nil, err
	}
	ud.Signature = base64.StdEncoding.EncodeToString(bsignature)

	// 打包数据
	b, err := json.Marshal(ud)
	if err != nil {
		return nil, nil, err
	}
	body := string(b)
	fmt.Println("ok send msg:", body)

	// 发送数据
	var res string
	nethelper.CallToHttpServer(addr, path, body, &res)
	fmt.Println("ok get ack:", res)

	// 解包数据
	ackData := &api.UserResponseData{}
	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != data.NoErr {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	// 解密验证数据
	var d2 []byte
	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, nil, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, nil, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, wallet_server_pubkey)
	if err != nil {
		return ackData, nil, err
	}

	// 解密数据
	d2, err = utils.RsaDecrypt(bencrypted2, web_admin_prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, nil, err
	}
	ackData.Value.Message = string(d2)

	return ackData, d2, nil
}

////////////////////////////////////////////////////////////////////////////////
type Web struct{
	nodes []*v1.SrvRegisterData

	loginUsers map[string]interface{}
}

func NewWeb() *Web {
	w := &Web{}
	w.loginUsers = make(map[string]interface{})

	return w
}

func (self *Web)Init(dataDir string) error {
	if err := loadAdministratorRsaKeys(dataDir); err != nil {
		return err
	}

	if err := self.startHttpServer(); err != nil{
		return err
	}

	return nil
}

// start http server
func (self *Web) startHttpServer() error {
	// http
	l4g.Debug("Start http server on 8077")

	http.Handle("/listsrv", http.HandlerFunc(self.handleListSrv))
	http.Handle("/getapi", http.HandlerFunc(self.handleGetApi))
	http.Handle("/runapi", http.HandlerFunc(self.handleRunApi))

	http.Handle("/doc/", http.FileServer(http.Dir("template")))
	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	http.Handle("/index",http.HandlerFunc(self.handleIndex))
	http.Handle("/login",http.HandlerFunc(self.handleLogin))
	http.Handle("/register",http.HandlerFunc(self.handleRegister))
	http.Handle("/dologin",http.HandlerFunc(self.LoginAction))
	http.Handle("/doregister",http.HandlerFunc(self.RegisterAction))
	http.Handle("/testapi", http.HandlerFunc(self.handleTestApi))
	http.Handle("/devsetting", http.HandlerFunc(self.handleDevSetting))
	http.Handle("/dodevsetting",http.HandlerFunc(self.DevSettingAction))
	http.Handle("/wallet/", http.HandlerFunc(self.handleWallet))
	http.Handle("/",http.HandlerFunc(self.handle404))

	go func() {
		l4g.Info("Http server routine running... ")
		err := http.ListenAndServe(":8077", nil)
		if err != nil {
			l4g.Crashf("", err)
			return
		}
	}()

	return nil
}

func (self *Web) handle404(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/index", http.StatusFound)
	}

	t, err := template.ParseFiles("template/html/404.html")
	if (err != nil) {
		l4g.Error("%s", err.Error())
	}
	t.Execute(w, nil)
}

// http handler
func (self *Web) handleListSrv(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	// 获取cookie
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	self.nodes = self.nodes[:0]
	d1, _, err := sendPostData(httpaddrGateway, cookie.Value, "", "v1", "gateway", "listsrv")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if d1.Err != data.NoErr {
		w.Write([]byte(d1.ErrMsg))
		return
	}

	json.Unmarshal([]byte(d1.Value.Message), &self.nodes)
	l4g.Debug(d1)

	// listsrv
	t, err := template.ParseFiles("template/html/listsrv.html")
	if err != nil {
		return
	}

	t.Execute(w, self.nodes)
	return
}

// http handler
func (self *Web) handleGetApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
	}

	if len(self.nodes) == 0 {
		d1, _, err := sendPostData(httpaddrGateway, cookie.Value, "", "v1", "center", "listsrv")
		if d1.Err != data.NoErr {
			w.Write([]byte(d1.ErrMsg))
			return
		}
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		json.Unmarshal([]byte(d1.Value.Message), &self.nodes)
	}

	// getapi?srv
	vv := req.URL.Query()
	srvname := vv.Get("srv")
	vername := vv.Get("ver")

	apiGroup, err := apigroup.ListApiGroupBySrv(vername, srvname)
	if err != nil {
		w.Write([]byte("获取API失败"))
		return
	}
	fmt.Println(apiGroup)

	//srvNode := data.SrvRegisterData{}
	//for _, v := range self.nodes {
	//	if v.Srv == srvname && v.Version == vername{
	//		srvNode = *v
	//		break
	//	}
	//}

	t, err := template.ParseFiles("template/html/getapi.html")
	if err != nil {
		return
	}

	t.Execute(w, apiGroup)
	return
}

// http handler
func (self *Web) handleRunApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	example := ""
	bb, err := ioutil.ReadAll(req.Body)
	u, err := url.Parse(string(bb))
	kvs := strings.Split(u.Path, "&")
	for _, v := range kvs{
		kvs2 := strings.Split(v, "=")
		if len(kvs2) == 2 && kvs2[0] == "argv" {
			example = kvs2[1]
			break
		}
	}
	//fmt.Println("argv", example)

	if len(self.nodes) == 0 {
		d1, _, err := sendPostData(httpaddrGateway, cookie.Value, "", "v1", "gateway", "listsrv")
		if d1.Err != data.NoErr {
			w.Write([]byte(d1.ErrMsg))
			return
		}
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		json.Unmarshal([]byte(d1.Value.Message), &self.nodes)
	}

	// getapi?srv
	vv := req.URL.Query()
	srv := vv.Get("srv")
	ver := vv.Get("ver")
	function := vv.Get("func")
	fmt.Println(srv, ver, function)

	var ures api.UserResponseData
	d1, _, err := sendPostData(httpaddrGateway, cookie.Value, example, ver, srv, function)
	if d1.Err != data.NoErr {
		w.Write([]byte(d1.ErrMsg))
		return
	}
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	ures = *d1

	t, err := template.ParseFiles("template/html/runapi.html")
	if err != nil {
		return
	}

	t.Execute(w, ures)
	return
}

type LoginUser struct {
	UserKey string
	Demo1 string
	Demo2 string
}
// http handler
func (self *Web) handleIndex(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	// listsrv
	t, err := template.ParseFiles("template/html/index.html")
	if err != nil {
		return
	}

	t.Execute(w, &LoginUser{UserKey:cookie.Value, Demo1:req.Host+"/listsrv", Demo2:"/testapi"})
	return
}

// http handler
func (self *Web) handleLogin(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	// listsrv
	t, err := template.ParseFiles("template/html/login.html")
	if err != nil {
		l4g.Error("%s", err.Error())
		return
	}

	t.Execute(w, nil)
	return
}

type UserDev struct {
	SerPubKey string
}
// http handler
func (self *Web) handleDevSetting(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	// listsrv
	t, err := template.ParseFiles("template/html/devsetting.html")
	if err != nil {
		l4g.Error("%s", err.Error())
		return
	}

	t.Execute(w, &UserDev{SerPubKey:string(wallet_server_pubkey)})
	return
}

// http handler
func (self *Web) handleRegister(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)

	// listsrv
	t, err := template.ParseFiles("template/html/register.html")
	if err != nil {
		l4g.Error("%s", err.Error())
		return
	}

	t.Execute(w, nil)
	return
}

func (this *Web)LoginAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	ures := api.UserResponseData{}

	func(){
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			return
		}
		fmt.Println("argv=", string(bb))

		ul := struct {
			UserKey string `json:"user_key""`
		}{}
		json.Unmarshal(bb, &ul)

		if ul.UserKey == "" {
			ures.Err = 1
			ures.ErrMsg = "no user key"
			return
		}

		d1, _, err := sendPostData(httpaddrGateway, ul.UserKey, "", "v1", "account", "readprofile")
		fmt.Println(d1)

		ures = *d1
		if d1.Err != data.NoErr {
			return
		}
		if err != nil {
			return
		}

		// 存入cookie,使用cookie存储
		//expiration := time.Unix(5, 0)
		cookie := http.Cookie{Name: "name", Value: ul.UserKey, Path: "/"}
		http.SetCookie(w, &cookie)
	}()

	b, _ := json.Marshal(ures)
	w.Write(b)

	return
}

func (this *Web)DevSettingAction(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("name")
	if err != nil || cookie.Value == ""{
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("content-type", "application/json")

	ures := api.UserResponseData{}

	func(){
		message := ""
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			return
		}
		message = string(bb)
		fmt.Println("dev argv=", message)

		ul := v1.ReqUserUpdateProfile{}
		json.Unmarshal(bb, &ul)

		if ul.PublicKey == "" && ul.SourceIP == "" && ul.CallbackUrl == ""{
			ures.Err = 1
			ures.ErrMsg = "no pubkey, sourceip and url"
			return
		}

		b1, err := base64.StdEncoding.DecodeString(ul.PublicKey)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			return
		}

		ul.PublicKey = string(b1)
		m, err := json.Marshal(ul)

		d1, _, err := sendPostData(httpaddrGateway, cookie.Value, string(m), "v1", "account", "updateprofile")
		fmt.Println(d1)

		ures = *d1
		if d1.Err != data.NoErr {
			return
		}
		if err != nil {
			return
		}
	}()

	b, _ := json.Marshal(ures)
	w.Write(b)

	return
}

func (this *Web)RegisterAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	ures := api.UserResponseData{}

	func(){
		message := ""
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		uc := v1.ReqUserRegister{}
		json.Unmarshal(bb, &uc)

		d1, _, err := sendPostData(httpaddrGateway, "", message, "v1", "account", "register")
		fmt.Println(d1)

		ures = *d1
		if d1.Err != data.NoErr {
			return
		}
		if err != nil {
			return
		}
	}()

	b, _ := json.Marshal(ures)
	w.Write(b)

	return
}

// http handler
func (self *Web) handleTestApi(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	//fmt.Println("path=", req.URL.Path)
	//fmt.Println("query=", req.URL.RawQuery)
	cookie, err := req.Cookie("name")
	if err != nil || cookie.Value == ""{
		//http.Redirect(w, req, "/login", http.StatusFound)
		//return
	}

	// listsrv
	t, err := template.ParseFiles("template/html/testapi.html")
	if err != nil {
		return
	}

	t.Execute(w, nil)
	return
}

// http handler
func (self *Web) handleWallet(w http.ResponseWriter, req *http.Request) {
	//log.Println("Http server Accept a rest client: ", req.RemoteAddr)
	//defer req.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	//cookie, err := req.Cookie("name")
	//if err != nil || cookie.Value == ""{
	//	http.Redirect(w, req, "/login", http.StatusFound)
	//	return
	//}

	var ures api.UserResponseData
	func(){
		fmt.Println("path=", req.URL.Path)

		method := api.UserMethod{}
		data.ApiMethodFromPath(&method, req.URL.Path)

		message := ""
		bb, err := ioutil.ReadAll(req.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		d1, _, err := sendPostData(httpaddrGateway, "", message, method.Version, method.Srv, method.Function)
		if d1.Err != data.NoErr {
			w.Write([]byte(d1.ErrMsg))
			return
		}
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		ures = *d1
	}()

	b, _ := json.Marshal(ures)

	w.Write(b)
	return
}