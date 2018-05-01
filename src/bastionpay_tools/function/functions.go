package function

import (
	"blockchain_server/service"
	"blockchain_server/types"
	_ "github.com/mattn/go-sqlite3"
	"bastionpay_tools/handler"
	"os"
	"bastionpay_tools/common"
)

type Functions struct{
	clientManager *service.ClientManager
	dataDir string
}

func (f *Functions)GetClientManager() *service.ClientManager {
	return f.clientManager
}

func (f *Functions)GetDataDir() string {
	return f.dataDir
}

func (f *Functions)GetAddressDataDir() string {
	return f.dataDir + "/" + common.AddressDirName
}

func (f *Functions)GetTxDataDir() string {
	return f.dataDir + "/" + common.TxDirName
}

func (f *Functions) Init(clientManager *service.ClientManager, dataDir string) error {
	f.clientManager = clientManager
	f.dataDir = dataDir

	var err error
	err = os.Mkdir(dataDir + "/" + common.AddressDirName, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(dataDir + "/" + common.TxDirName, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

// offline - newaddress
// create addresses to a db file in data dir, named by uniname
func (f *Functions) NewAddress(coinType string, count uint32) (string, error) {
	return handler.NewAddress(f.clientManager, f.GetAddressDataDir(), coinType, count)
}

// offline - signtx
// sign transaction from a file to a signed file
func (f *Functions)SignTx(txFilePath, txSignedFilePath string) (error) {
 	return handler.SignTx(f.clientManager, f.GetAddressDataDir(), txFilePath, txSignedFilePath)
}

// offline - load offline address
// load addresses form a db file in data dir, named by uniname
func (f *Functions) LoadOfflineAddress(uniName string) ([]*types.Account, error) {
	return handler.LoadOfflineAddress(f.GetAddressDataDir(), uniName)
}

// offline/online - load online address
// load addresses form a db file in data dir, named by uniname
func (f *Functions) LoadOnlineAddress(uniName string) ([]*types.Account, error) {
	return handler.LoadOnlineAddress(f.GetAddressDataDir(), uniName)
}

// online - buildtx
// build transactions to a file for sign
func (f *Functions)BuildTx(txArr []*types.CmdSendTx, txFilePath string) (error) {
	return handler.BuildTx(f.clientManager, txArr, txFilePath)
}

// online - sendsignedtx
// send signed tx from a file
func (f *Functions)SendSignedTx(txSignedFilePath string) (error) {
	return handler.SendSignedTx(f.clientManager, txSignedFilePath)
}