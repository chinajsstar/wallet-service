package handler

import (
	"api_router/base/data"
	"encoding/json"
	"net/http"
	"html/template"
	"api_router/base/utils"
	"io/ioutil"
	"net/url"
	"strings"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"api_router/account_srv/user"
	"api_router/base/nethelper"
	"encoding/base64"
	"crypto/sha512"
	"crypto"
	"errors"
)

const httpaddrGateway = "http://127.0.0.1:8080"

var web_admin_prikey []byte
var web_admin_pubkey []byte
var web_admin_userkey string

var wallet_server_pubkey []byte
func loadAdministratorRsaKeys() error {
	var err error
	web_admin_prikey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_administrator.pem")
	if err != nil {
		return err
	}

	web_admin_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_administrator.pem")
	if err != nil {
		return err
	}

	web_admin_userkey = "1c75c668-f1ab-474b-9dae-9ed7950604b4"

	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	accountDir := appDir + "/account"
	wallet_server_pubkey, err = ioutil.ReadFile(accountDir + "/public.pem")
	if err != nil {
		return err
	}

	return nil
}

func sendPostData(addr, message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	var ud data.UserData

	// 构建path
	path := "/wallet"
	path += "/"+version
	path += "/"+srv
	path += "/"+function

	// user key
	ud.UserKey = web_admin_userkey

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
	ackData := &data.UserResponseData{}
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
	nodes []*data.SrvRegisterData

	users map[string]*user.AckUserLogin
}

func NewWeb() *Web {
	w := &Web{}
	w.users = make(map[string]*user.AckUserLogin)

	return w
}

func (self *Web)Init() error {
	if err := loadAdministratorRsaKeys(); err != nil {
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
	http.Handle("/dologin",http.HandlerFunc(self.LoginAction))
	http.Handle("/testapi", http.HandlerFunc(self.handleTestApi))
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
	}

	self.nodes = self.nodes[:0]
	d1, _, err := sendPostData(httpaddrGateway, "", "v1", "center", "listsrv")
	if d1.Err != data.NoErr {
		w.Write([]byte(d1.ErrMsg))
		return
	}
	if err != nil {
		w.Write([]byte(err.Error()))
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
		d1, _, err := sendPostData(httpaddrGateway, "", "v1", "center", "listsrv")
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

	srvNode := data.SrvRegisterData{}
	for _, v := range self.nodes {
		if v.Srv == srvname {
			srvNode = *v
			break
		}
	}

	t, err := template.ParseFiles("template/html/getapi.html")
	if err != nil {
		return
	}

	t.Execute(w, srvNode)
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
		d1, _, err := sendPostData(httpaddrGateway, "", "v1", "center", "listsrv")
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

	var ures data.UserResponseData
	if example != ""{
		d1, _, err := sendPostData(httpaddrGateway, example, ver, srv, function)
		if d1.Err != data.NoErr {
			w.Write([]byte(d1.ErrMsg))
			return
		}
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		ures = *d1
	}else{
		ures.Err = 1
		ures.ErrMsg = "没有提供测试实例"
	}

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

func (this *Web)LoginAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	ures := data.UserResponseData{}

	func(){
		message := ""
		bb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			ures.ErrMsg = data.ErrDataCorruptedText
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		ul := user.ReqUserLogin{}
		json.Unmarshal(bb, &ul)

		if ul.UserName == "" || ul.Password == ""{
			ures.Err = 1
			ures.ErrMsg = "参数错误"
			return
		}

		d1, _, err := sendPostData(httpaddrGateway, message, "v1", "account", "login")
		if d1.Err != data.NoErr {
			w.Write([]byte(d1.ErrMsg))
			return
		}
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Println(d1)

		aul := user.AckUserLogin{}
		json.Unmarshal([]byte(d1.Value.Message), &aul)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		// 存入cookie,使用cookie存储
		//expiration := time.Unix(5, 0)
		cookie := http.Cookie{Name: "name", Value: aul.UserKey, Path: "/"}
		http.SetCookie(w, &cookie)

		ures = *d1
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

	var ures data.UserResponseData
	func(){
		fmt.Println("path=", req.URL.Path)

		path := req.URL.Path
		path = strings.Replace(path, "wallet", "", -1)
		path = strings.TrimLeft(path, "/")
		path = strings.TrimRight(path, "/")

		ver := ""
		srv := ""
		function := ""
		// get method
		paths := strings.Split(path, "/")
		for i := 0; i < len(paths); i++ {
			if i == 0 {
				ver = paths[i]
			}else if i == 1{
				srv = paths[i]
			} else{
				if function != "" {
					function += "."
				}
				function += paths[i]
			}
		}

		message := ""
		bb, err := ioutil.ReadAll(req.Body)
		if err != nil {
			ures.Err = data.ErrDataCorrupted
			ures.ErrMsg = data.ErrDataCorruptedText
			return
		}
		message = string(bb)
		fmt.Println("argv=", message)

		d1, _, err := sendPostData(httpaddrGateway, message, ver, srv, function)
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