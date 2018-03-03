package eth

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"fmt"
	"context"
	"time"
)

func NewHttpClient(url string) (client *ethclient.Client, err error) {
	rpc_client, err := rpc.Dial(url)

	if err != nil {
		fmt.Println("can not dail to : ", url)
		return nil, err
	}

	if rpc_client == nil {
		return nil, fmt.Errorf("rpc client is nil")
	}

	client = ethclient.NewClient(rpc_client)

	return
}

func NewWsClient(url string)(*ethclient.Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)

	client_ch := make(chan *rpc.Client)
	err_ch := make(chan error)

	go func(ctx context.Context, url string, client_ch chan *rpc.Client, err_ch chan error) {
		client, err := rpc.DialWebsocket(ctx, url, "")
		if nil!=err {
			err_ch<-err
			return
		}
		client_ch<-client

	}(ctx, url, client_ch, err_ch)

	select {
		case <-ctx.Done():
			err_message := fmt.Sprintf("connect to ws server error:%v\n", ctx.Err())
			fmt.Println(err_message)
			return nil, fmt.Errorf(err_message)
		case err := <-err_ch:
			return nil, err
		case c := <- client_ch:
			client :=  ethclient.NewClient(c)
			return client, nil
	}
}


//func (this *s_eth_rpc_client) Call(reply interface{}, serviceMethod string, args ...interface{}) error {
//err := this.client.
//eh.client.Call(&reply, serviceMethod, args...)
//if err != nil {
//fmt.Println("Call : ", err)
//return err
//}
//return nil
//}

//func NewEthIpc() (*ethIpcHandler, error) {
//eh := new(ethIpcHandler)
//usr, err := user.Current()
//if err != nil {
//log.Println(err)
//return nil, nil
//}
//eh.ipcFileLocation = viper.GetString("IPC_PATH")
//if len(eh.ipcFileLocation) == 0 {
//eh.ipcFileLocation = usr.HomeDir + "/Library/Ethereum/geth.ipc"
//}
//_, err = os.Stat(eh.ipcFileLocation)
//if os.IsNotExist(err) {
//return nil, nil
//}
//client, err := rpc.DialIPC(context.TODO(), eh.ipcFileLocation)
//// laddr := net.UnixAddr{Net: "unix", Name: eh.ipcFileLocation}
//// conn, err := net.DialUnix("unix", nil, &laddr)
//if err != nil {
//fmt.Println("DialUnix : ", err)
//return nil, err
//}
////defer client.Close()
//eh.client = client
//return eh, err
//}
