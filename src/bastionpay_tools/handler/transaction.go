package handler

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/db"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/json"
)

type TxData struct{
	Address string 	`json:"address"` //地址
	PriKey  string 	`json:"prikey"`  //标示
	Signed  int 	`json:"signed"`	 //是否签名0：无，1：签名
	Tx 		string 	`json:"tx"` //数据
}

func BuildTx(clientManager *service.ClientManager, txArr []*types.CmdSendTx, txFilePath string) (error) {
	var txDataArr []*TxData
	for _, tx := range txArr {
		s, err := clientManager.BuildTx(tx)
		if err != nil {
			fmt.Printf("生成交易失败: %s\n", err.Error())
			return err
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
		fmt.Printf("序列化数据失败: %s\n", err.Error())
		return err
	}

	err = ioutil.WriteFile(txFilePath, b, 0644)
	if err != nil {
		fmt.Printf("写入数据失败: %s\n", err.Error())
		return err
	}

	fmt.Println("生成交易完成")
	return nil
}

func SignTx(clientManager *service.ClientManager, addressDataDir string, txFilePath, txSignedFilePath string) (error) {
	txByte, err := ioutil.ReadFile(txFilePath)
	if err != nil {
		fmt.Printf("读取交易文件失败: %s\n", err.Error())
		return err
	}

	// 加载数组
	txDataArr := []TxData{}
	err = json.Unmarshal(txByte, &txDataArr)
	if err != nil {
		fmt.Printf("解析交易文件失败: %s\n", err.Error())
		return err
	}

	// 转成map
	txDataMap := make(map[string][]TxData)
	for _, v := range txDataArr {
		txDataMap[v.PriKey] = append(txDataMap[v.PriKey], v)
	}

	txDataArrSinged := []TxData{}

	for uniName, txArr := range txDataMap {
		// 加载离线DB
		uniOfflineDBName := db.GetOfflineUniDBName(uniName)
		aCcsOfflineMap, err := db.ImportAddressMap(addressDataDir, uniOfflineDBName)
		if err != nil {
			fmt.Printf("加载离线文件失败: %s\n", err.Error())
			return err
		}

		for _, tx := range txArr {
			aCcsOffline, ok := aCcsOfflineMap[tx.Address]
			if !ok {
				fmt.Printf("没有在离线文件中找到私钥: %s\n", tx.Address)
				return errors.New("not find chiper prikey")
			}

			// 签名
			txSigned := tx

			chiperPrikey := aCcsOffline.PrivateKey
			txString := tx.Tx

			s, err := clientManager.SignTx(chiperPrikey, txString)
			if err != nil {
				fmt.Printf("签名失败: %s\n", err.Error())
				return err
			}

			txSigned.Tx = s
			txSigned.Signed = 1

			txDataArrSinged = append(txDataArrSinged, txSigned)
		}
	}

	bSigned, err := json.Marshal(txDataArrSinged)
	if err != nil {
		fmt.Printf("序列化数据失败: %s\n", err.Error())
		return err
	}

	err = ioutil.WriteFile(txSignedFilePath, bSigned, 0644)
	if err != nil {
		fmt.Printf("写入签名数据失败: %s\n", err.Error())
		return err
	}

	fmt.Println("签名交易完成")
	return nil
}

func SendSignedTx(clientManager *service.ClientManager, txSignedFilePath string) (error) {
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