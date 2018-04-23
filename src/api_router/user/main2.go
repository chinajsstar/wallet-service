package main

import (
	//"golang.org/x/net/websocket"
	"fmt"
	"../base/utils"
	"golang.org/x/net/websocket"
	"../base/data"
	"encoding/json"
	"encoding/base64"
	"crypto/sha512"
	"crypto"
	"errors"
	"io/ioutil"
	//l4g "github.com/alecthomas/log4go"
	"../base/config"
	"reflect"
)

var G_admin_prikey2 []byte
var G_admin_pubkey2 []byte
var G_admin_licensekey2 string

var G_server_pubkey2 []byte

func LoadRsaKeys2() error {
	var err error
	G_admin_prikey2, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_admin.pem")
	if err != nil {
		return err
	}

	G_admin_pubkey2, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_admin.pem")
	if err != nil {
		return err
	}

	G_server_pubkey2, err = ioutil.ReadFile("/Users/henly.liu/wallet-service/src/api_router/account_srv/worker/public.pem")
	if err != nil {
		return err
	}

	G_admin_licensekey2 = "25143234-b958-44a8-a87f-5f0f4ef46eb5"

	return nil
}

func encryptData(message string, userData *data.UserData) (error) {
	// 用户数据
	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), G_server_pubkey2, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return err
	}

	userData.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error) {
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, G_admin_prikey2)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return err
	}

	userData.Signature = base64.StdEncoding.EncodeToString(bsignature)

	return nil
}

func decryptData(res string) (*data.UserResponseData, error) {
	ackData := &data.UserResponseData{}
	err := json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, err
	}

	if ackData.Err != data.NoErr {
		return ackData, errors.New("# got err: " + ackData.ErrMsg)
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, G_server_pubkey2)
	if err != nil {
		return ackData, err
	}

	// 解密数据
	_, err = utils.RsaDecrypt(bencrypted2, G_admin_prikey2, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, err
	}

	return ackData, nil
}

func StartWsClient() *websocket.Conn {
	conn, err := websocket.Dial("ws://127.0.0.1:8040/ws", "", "test://wallet/")
	if err != nil {
		fmt.Println("#error", err)
		return nil
	}

	go func(conn *websocket.Conn) {
		var data []byte
		for ; ; {
			_, err := conn.Read(data)
			if err != nil {
				fmt.Println("read failed:", err)
				break
			}

			fmt.Println("read:", string(data))
		}
	}(conn)
	return conn
}
/*
func StartWsServer2() (*rpc2.Server, net.Listener) {
	startServer := func(addr string) (*rpc2.Server, net.Listener) {
		srv := rpc2.NewServer()
		srv.RegisterName("arith", new(Arith))
		l, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Println(err)
		}
		go http.Serve(l, srv.WebsocketHandler([]string{"*"}))
		return srv, l
	}

	srv, l1 := startServer("127.0.0.1:8300")
	fmt.Println(l1.Addr().String())
	return srv, l1
}

func StartWsClient2() *rpc2.Client {
	client, err := rpc2.Dial("ws://127.0.0.1:8300")
	if err != nil {
		fmt.Println("can't dial", err)
		return nil
	}
	return client
}
*/

type Args2 struct {
	A int `json:"a" comment:"加数1"`
	B int `json:"b" comment:"加数2"`
}

func FieldTag(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}
	n := t.NumField()
	for i := 0; i < n; i++ {
		fmt.Println(t.Field(i).Tag.Get("json"), "--", t.Field(i).Tag.Get("comment"))
	}
	return true
}

func main() {
	appDir, _:= utils.GetCurrentDir()
	appDir += "/test.json"

	cn := config.ConfigNode{}
	err := config.LoadJsonNode(appDir, "node", &cn)
	fmt.Println(cn)
	fmt.Println(err)

	a2 := Args2{}
	at2 := reflect.TypeOf(a2)
	FieldTag(at2)

	return

/*
	// Start a server and corresponding client.
	////
	LoadRsaKeys2()
	var conn *websocket.Conn

	for ; ; {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0]=="q"{
			break
		} else if argv[0] == "w" {
		}else if argv[0] == "c" {
			conn = StartWsClient()
		}else if argv[0] == "s" {
			if(conn != nil){
				conn.Write([]byte(argv[1]))
			}
		}else if argv[0] == "login" {
			m, err := install.LoginUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var ud data.UserData
			encryptData(string(d), &ud)

			dispatchData := data.UserRequestData{}
			dispatchData.Method.Version = "v1"
			dispatchData.Method.Srv = "account"
			dispatchData.Method.Function = "login"
			dispatchData.Argv = ud

			d, err = json.Marshal(dispatchData)

			if conn != nil && err == nil{
				conn.Write(d)
			}
		}
	}
*/
	//return
}

