package handler

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/db"
	l4g "github.com/alecthomas/log4go"
	"fmt"
	"errors"
	"github.com/satori/go.uuid"
	"time"
)

// 返回文件名的唯一标示
func NewAddress(clientManager *service.ClientManager, dataDir string, coinType string, count uint32) (string, error) {
	fmt.Printf("==============NewAddressByCmd================\n")
	fmt.Printf("您正在创建新地址，类型：%s, 数量为：%d\n", coinType, count)

	var err error

	// 创建新地址
	accCmd := service.NewAccountCmd("message id", coinType, uint32(count))
	var aCcs []*types.Account
	aCcs, err = clientManager.NewAccounts(accCmd)
	if err != nil {
		fmt.Printf("创建新地址失败: %s\n", err.Error())
		return "", errors.New("create address error")
	}
	fmt.Printf("创建新地址成功，数量: %d\n", len(aCcs))

	// uuid
	fmt.Printf("生成唯一标示，数量: %d\n", len(aCcs))
	uniName, err := func()(string, error) {
		uuidv4, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		uuid := uuidv4.String()

		datetime := time.Now().UTC().Format("2006-01-02-15-04-05")
		return coinType + "@" + datetime + "@" + uuid, nil
	}()
	if err != nil {
		l4g.Error("生成唯一标示错误: %s", err.Error())
		return "", err
	}
	fmt.Println("创建唯一标示：%s", uniName)

	// 保存新地址
	err = db.ExportAddress(dataDir, uniName, aCcs)
	if err != nil {
		fmt.Printf("导出新地址失败: %s\n", err.Error())
		return "", errors.New("export address error")
	}
	fmt.Printf("导出新地址成功，唯一标示: %s\n", uniName)

	// 校验一次
	fmt.Printf("开始校验: %s\n", uniName)

	// 校验在线文件
	err = db.VerifyOnlineDBFile(dataDir, uniName, aCcs)
	if err != nil {
		fmt.Printf("校验在线文件失败：%s\n", err.Error())
		return "", errors.New("check online error")
	}

	// 校验离线文件
	err = db.VerifyOfflineDBFile(dataDir, uniName, aCcs)
	if err != nil {
		fmt.Printf("校验离线文件失败：%s\n", err.Error())
		return "", errors.New("check offline error")
	}

	fmt.Println("校验完成")

	return uniName, nil
}

func LoadOnlineAddress(dataDir string, uniName string) ([]*types.Account, error) {
	uniDBName := db.GetOnlineUniDBName(uniName)
	aCcs, err := db.ImportAddress(dataDir, uniDBName)
	if err != nil {
		fmt.Printf("加载地址失败: %s\n", err.Error())
		return nil, err
	}

	return aCcs, nil
}

func LoadOfflineAddress(dataDir string, uniName string) ([]*types.Account, error) {
	uniDBName := db.GetOfflineUniDBName(uniName)
	aCcs, err := db.ImportAddress(dataDir, uniDBName)
	if err != nil {
		fmt.Printf("加载地址失败: %s\n", err.Error())
		return nil, err
	}

	return aCcs, nil
}