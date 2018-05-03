package function

import (
	"blockchain_server/service"
	"blockchain_server/types"
	_ "github.com/mattn/go-sqlite3"
	"bastionpay_tools/handler"
	"os"
	"bastionpay_tools/common"
	"api_router/base/utils"
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
func (f *Functions) NewAddress(coinType string, count uint32, savedDir string) (string, error) {
	mul := count / common.MaxDbCountAddress
	mod := count % common.MaxDbCountAddress
	if mod != 0 {
		mul = mul + 1
	}

	// 生成本次唯一id
	uniAddrInfo, err:= common.NewUniAddressInfo(coinType)
	if err != nil {
		return "", err
	}

	// 创建唯一目录
	err = uniAddrInfo.MkUniAbsDir(f.GetAddressDataDir())
	if err != nil {
		return "", err
	}

	// 批量生成
	var addressDbInfos []*common.UniAddressDbInfo
	left := count
	var realCount uint32
	for left > 0 {
		realCount = common.MaxDbCountAddress
		if left < common.MaxDbCountAddress {
			realCount = left
		}

		uniAddrDbInfo, err := handler.NewAddress(f.clientManager, f.GetAddressDataDir(), uniAddrInfo, realCount)
		if err != nil {
			return "", err
		}

		// next loop
		addressDbInfos = append(addressDbInfos, uniAddrDbInfo)
		left -= realCount
	}

	// 创建目的唯一目录
	err = uniAddrInfo.MkUniAbsDir(savedDir)
	if err != nil {
		return "", err
	}

	srcUniDir := uniAddrInfo.GetUniAbsDir(f.GetAddressDataDir())
	dstUniDir := uniAddrInfo.GetUniAbsDir(savedDir)
	for _, ai := range addressDbInfos {
		fileName := uniAddrInfo.GetUniName() + "@" + ai.GetUniNameOnline() + "@" + common.GetOnlineDbNameSuffix()
		// copy
		srcPath := srcUniDir + "/" + fileName
		dstPath := dstUniDir + "/" + fileName

		_, err = utils.CopyFile(srcPath, dstPath)
		if err != nil {
			return "", err
		}
	}

	return dstUniDir, nil
}

// offline - signtx
// sign transaction from a file to a signed file
func (f *Functions)SignTx(txFilePath, txSignedFilePath string) (error) {
 	return handler.SignTx(f.clientManager, f.GetAddressDataDir(), txFilePath, txSignedFilePath)
}

// offline/online - load address
// load addresses form a db file in data dir, named by dbname
func (f *Functions) LoadAddress(uniDbName string) ([]*types.Account, error) {
	uniAddressInfo, err:= common.ParseUniAddressInfo(uniDbName)
	if err != nil {
		return nil, err
	}

	uniAbsDir := uniAddressInfo.GetUniAbsDir(f.GetAddressDataDir())
	return handler.LoadAddress(uniAbsDir, uniDbName)
}

// offline/online - load address
// load addresses form a db file in data dir, named by dbname
func (f *Functions) VerifyDbMd5(uniDbName string) (error) {
	uniAddressInfo, err:= common.ParseUniAddressInfo(uniDbName)
	if err != nil {
		return err
	}

	uniAddressLineDbInfo, err := common.ParseUniAddressLineDbInfo(uniDbName)
	if err != nil {
		return err
	}

	uniAbsDir := uniAddressInfo.GetUniAbsDir(f.GetAddressDataDir())
	uniAbsDbPath := uniAbsDir + "/" + uniDbName

	return common.CompareSaltMd5HexByFile(uniAbsDbPath, uniAddressLineDbInfo.Md5)
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