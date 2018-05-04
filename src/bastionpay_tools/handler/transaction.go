package handler

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/db"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/json"
	"bastionpay_tools/common"
	"os"
)

type TxData struct{
	Address string 	`json:"address"` //地址
	PriKey  string 	`json:"prikey"`  //标示
	Signed  int 	`json:"signed"`	 //是否签名0：无，1：签名
	Tx 		string 	`json:"tx"` //数据
}

func BuildTx(clientManager *service.ClientManager, txArr []*types.CmdSendTx, txDataDir string) (string, error) {
	var txDataArr []*TxData
	for _, tx := range txArr {
		s, err := clientManager.BuildTx(tx)
		if err != nil {
			return "", err
		}

		txData := &TxData{}
		txData.Tx = s
		txData.Signed = 0
		txData.Address = tx.Tx.From
		txData.PriKey = tx.Chiperkey

		txDataArr = append(txDataArr, txData)
	}

	b, err := json.Marshal(txDataArr)
	if err != nil {
		return "", err
	}

	// datetime@uuid@md5
	uniTx , err := common.NewUniTransaction()
	if err != nil {
		return "", err
	}

	txFilePath := txDataDir + "/" + uniTx.GetUniName() + common.GetTxExtension()
	err = ioutil.WriteFile(txFilePath, b, 0644)
	if err != nil {
		fmt.Printf("写入数据失败: %s\n", err.Error())
		return "", err
	}

	return MakeTxMd5(txFilePath, txDataDir)
}

func SignTx(clientManager *service.ClientManager, addressDir string, txDir string, txFilePath string) (string, error) {
	// 验证md5
	err := VerifyTxMd5(txFilePath)
	if err != nil {
		return "", err
	}

	txByte, err := ioutil.ReadFile(txFilePath)
	if err != nil {
		fmt.Printf("读取交易文件失败: %s\n", err.Error())
		return "", err
	}

	// 加载数组
	txDataArr := []TxData{}
	err = json.Unmarshal(txByte, &txDataArr)
	if err != nil {
		fmt.Printf("解析交易文件失败: %s\n", err.Error())
		return "", err
	}

	// 转成map
	txDataMap := make(map[string][]TxData)
	for _, v := range txDataArr {
		txDataMap[v.PriKey] = append(txDataMap[v.PriKey], v)
	}

	txDataArrSinged := []TxData{}

	for uniName, txArr := range txDataMap {
		// 加载离线DB
		uniAddressInfo, err:= common.ParseUniAddressInfo(uniName)
		if err != nil {
			return "", err
		}

		uniAbsDir := uniAddressInfo.GetUniAbsDir(addressDir)

		uniOfflineAbsDBName := uniAbsDir + "/" + uniName + common.GetOfflineExtension()
		aCcsOfflineMap, err := db.LoadAddressMap(uniOfflineAbsDBName)
		if err != nil {
			fmt.Printf("加载离线文件失败: %s\n", err.Error())
			return "", err
		}

		for _, tx := range txArr {
			aCcsOffline, ok := aCcsOfflineMap[tx.Address]
			if !ok {
				fmt.Printf("没有在离线文件中找到私钥: %s\n", tx.Address)
				return "", errors.New("not find chiper prikey")
			}

			// 签名
			txSigned := tx

			chiperPrikey := aCcsOffline.PrivateKey
			txString := tx.Tx

			s, err := clientManager.SignTx(chiperPrikey, txString)
			if err != nil {
				fmt.Printf("签名失败: %s\n", err.Error())
				return "", err
			}

			txSigned.Tx = s
			txSigned.Signed = 1

			txDataArrSinged = append(txDataArrSinged, txSigned)
		}
	}

	bSigned, err := json.Marshal(txDataArrSinged)
	if err != nil {
		fmt.Printf("序列化数据失败: %s\n", err.Error())
		return "", err
	}

	// datetime@uuid@md5
	uniTxs , err := common.NewUniTransaction()
	if err != nil {
		return "", err
	}

	txSignedFilePath := txDir + "/" + uniTxs.GetUniName() + common.GetTxSignedExtension()
	err = ioutil.WriteFile(txSignedFilePath, bSigned, 0644)
	if err != nil {
		fmt.Printf("写入签名数据失败: %s\n", err.Error())
		return "", err
	}

	return MakeTxMd5(txSignedFilePath, txDir)
}

func SendSignedTx(clientManager *service.ClientManager, txSignedFilePath string) (error) {
	// 验证md5
	err := VerifyTxMd5(txSignedFilePath)
	if err != nil {
		return err
	}

	txSignedByte, err := ioutil.ReadFile(txSignedFilePath)
	if err != nil {
		fmt.Printf("读取签名交易文件失败: %s\n", err.Error())
		return err
	}

	// 加载数组
	txSignedDataArr := []TxData{}
	err = json.Unmarshal(txSignedByte, &txSignedDataArr)
	if err != nil {
		fmt.Printf("解析交易文件失败: %s\n", err.Error())
		return err
	}

	for _, txSignedData := range txSignedDataArr {
		err = clientManager.SendSignedTx(txSignedData.Tx)
		if err != nil {
			fmt.Printf("发送交易失败:%s\n", err.Error())
			return err
		}
	}

	fmt.Println("发送交易完成")
	return nil
}

func VerifyTxMd5(txFilePath string) error{
	fileInfo, err := os.Stat(txFilePath)
	if err != nil {
		return err
	}

	fileName := fileInfo.Name()
	uniTx, err := common.ParseUniTransaction(fileName)
	if err != nil {
		return err
	}

	return common.CompareSaltMd5HexByFile(txFilePath, uniTx.Md5)
}

func MakeTxMd5(txFilePath string, txDataDir string) (string, error){
	fileInfo, err := os.Stat(txFilePath)
	if err != nil {
		return "", err
	}

	fileName := fileInfo.Name()
	uniTx, err := common.ParseUniTransaction(fileName)
	if err != nil {
		return "", err
	}

	uniTxMd5 := *uniTx
	uniTxMd5.Md5, err = common.GetSaltMd5HexByFile(txFilePath)
	if err != nil {
		return "", err
	}

	txFilePathMd5 := txDataDir + "/" + uniTxMd5.GetUniName() + uniTxMd5.Ext
	err = os.Rename(txFilePath, txFilePathMd5)
	if err != nil {
		return "", err
	}

	return txFilePathMd5, nil
}