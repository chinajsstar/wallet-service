package function

import (
	"blockchain_server/service"
	"blockchain_server/types"
	_ "github.com/mattn/go-sqlite3"
	"bastionpay_tools/handler"
	"os"
	"bastionpay_tools/common"
	"api_router/base/utils"
	l4g "github.com/alecthomas/log4go"
	"io"
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
	l4g.Info("===开始生成地址文件===")

	dstUniDir, err := func()(string, error){
		mul := count / common.MaxDbCountAddress
		mod := count % common.MaxDbCountAddress
		if mod != 0 {
			mul = mul + 1
		}

		l4g.Info("本次共生成DB文件个数：%d", mul)

		// 生成本次唯一id
		uniAddrInfo, err:= common.NewUniAddressInfo(coinType)
		if err != nil {
			return "", err
		}

		l4g.Info("本次产生唯一标示：%s", uniAddrInfo.GetUniName())

		// 创建唯一目录
		err = uniAddrInfo.MkUniAbsDir(f.GetAddressDataDir())
		if err != nil {
			return "", err
		}

		l4g.Info("本次生成唯一目录：%s", uniAddrInfo.GetUniAbsDir(f.dataDir))

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

		// TODO: 先备份，或者先人工复制
		l4g.Info("地址生成完成...")
		l4g.Info("TODO:准备开始备份到备份目录")

		l4g.Info("开始复制到唯一目的目录")

		// 创建目的唯一目录
		err = uniAddrInfo.MkUniAbsDir(savedDir)
		if err != nil {
			return "", err
		}
		l4g.Info("生成唯一目的目录：%s", uniAddrInfo.GetUniAbsDir(savedDir))

		srcUniDir := uniAddrInfo.GetUniAbsDir(f.GetAddressDataDir())
		dstUniDir := uniAddrInfo.GetUniAbsDir(savedDir)
		for _, ai := range addressDbInfos {
			fileName := uniAddrInfo.GetUniName() + ai.GetUniNameOnline() + common.GetOnlineExtension()
			// copy
			srcPath := srcUniDir + "/" + fileName
			dstPath := dstUniDir + "/" + fileName

			_, err = utils.CopyFile(srcPath, dstPath)
			if err != nil {
				return "", err
			}

			l4g.Info("复制到：%s", dstPath)
		}
		l4g.Info("===结束复制到唯一目的目录===")

		return dstUniDir, nil
	}()

	if err != nil {
		l4g.Error("NewAddress: %s", err.Error())
	}

	l4g.Info("===结束生成地址文件===")
	return dstUniDir, nil
}

func (f* Functions) SaveAddress(uniName string, src io.Reader) error{
	// 解析本次唯一id
	uniAddrInfo, err:= common.ParseUniAddressInfo(uniName)
	if err != nil {
		return err
	}

	l4g.Info("唯一标示：%s", uniAddrInfo.GetUniName())

	// 创建唯一目录
	err = uniAddrInfo.MkUniAbsDir(f.GetAddressDataDir())
	if err != nil {
		return err
	}

	uniAbsDir := uniAddrInfo.GetUniAbsDir(f.GetAddressDataDir())
	l4g.Info("生成唯一目录：%s", uniAbsDir)

	saveFilePath := uniAbsDir + "/" + uniName
	file, err := os.OpenFile(saveFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, src)

	return nil
}

// offline - signtx
// sign transaction from a file to a signed file
func (f *Functions)SignTx(txFilePath string, saveDir string) (string, error) {
	l4g.Info("===开始签名交易文件===")

	dstPath, err := func()(string, error){
		uniTxs, err := handler.SignTx(f.clientManager, f.GetAddressDataDir(), f.GetTxDataDir(), txFilePath)

		if err != nil {
			l4g.Error("SignTx: %s", err.Error())
			return "", err
		}

		l4g.Info("复制签名交易文件到目的目录：%s", saveDir)

		fileInfo, err := os.Stat(uniTxs)
		if err != nil {
			return "", err
		}

		dstPath := saveDir + "/" + fileInfo.Name()
		_, err = utils.CopyFile(uniTxs, dstPath)
		if err != nil {
			return "", err
		}
		l4g.Info("===结束复制到唯一目的目录===")

		return dstPath, nil
	}()

	l4g.Info("===结束签名交易文件===")
 	return dstPath, err
}

// offline/online - load address
// load addresses form a db file in data dir, named by dbname
func (f *Functions) LoadAddress(uniDbName string) ([]*types.Account, error) {
	l4g.Info("===开始加载地址===")

	aCcs, err := func() ([]*types.Account, error) {
		uniAddressInfo, err:= common.ParseUniAddressInfo(uniDbName)
		if err != nil {
			return nil, err
		}

		uniAbsDir := uniAddressInfo.GetUniAbsDir(f.GetAddressDataDir())
		return handler.LoadAddress(uniAbsDir, uniDbName)
	}()

	if err != nil {
		l4g.Error("LoadAddress: %s", err.Error())
	}

	l4g.Info("===结束加载地址===")
	return aCcs, err
}

// offline/online - verify address
func (f *Functions) VerifyAddressMd5(uniName string) (error) {
	l4g.Info("===开始验证地址文件===")

	err := func() error {
		uniAddressInfo, err:= common.ParseUniAddressInfo(uniName)
		if err != nil {
			return err
		}

		uniAbsDir := uniAddressInfo.GetUniAbsDir(f.GetAddressDataDir())
		uniAbsDbPath := uniAbsDir + "/" + uniName

		return handler.VerifyAddressMd5(uniAbsDbPath)
	}()

	if err != nil {
		l4g.Error("BuildTx: %s", err.Error())
	}

	l4g.Info("===结束验证地址文件===")
	return err
}

// offline/online - verify tx
func (f *Functions) VerifyTxMd5(uniName string) (error) {
	l4g.Info("===开始验证交易文件===")

	err := handler.VerifyTxMd5(f.GetTxDataDir() + "/" + uniName)

	if err != nil {
		l4g.Error("BuildTx: %s", err.Error())
	}

	l4g.Info("===结束验证交易文件===")
	return err
}

// online - buildtx
// build transactions to a file for sign
func (f *Functions)BuildTx(txArr []*types.CmdSendTx) (string, error) {
	l4g.Info("===开始生成交易文件===")

	uniPath, err := handler.BuildTx(f.clientManager, txArr, f.GetTxDataDir())

	if err != nil {
		l4g.Error("BuildTx: %s", err.Error())
	}

	l4g.Info("===结束生成交易文件===")
	return uniPath, err
}

// online - sendsignedtx
// send signed tx from a file
func (f *Functions)SendSignedTx(txSignedFilePath string) (error) {
	l4g.Info("===开始发送签名交易文件===")

	err := handler.SendSignedTx(f.clientManager, txSignedFilePath)

	if err != nil {
		l4g.Error("SendSignedTx: %s", err.Error())
	}

	l4g.Info("===开始发送交易交易文件===")
	return err
}