package handler

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/db"
	//l4g "github.com/alecthomas/log4go"
	"fmt"
	"strconv"
	"errors"
	"encoding/json"
)

type NewAddressCmd struct{
	CoinType string `json:"cointype"`
	Count int `json:"count"`
}
type NewAddressRes struct{
	UniName string `json:"uniname"`
	OnlineDBPath string `json:"onlinedbpath"`
	OfflineDBPath string `json:"offlinedbpath"`
}
func NewAddressByCmd(clientManager *service.ClientManager, dataDir string, cmd *NewAddressCmd) (*NewAddressRes, error) {
	var err error

	t := cmd.CoinType
	count := cmd.Count

	fmt.Printf("==============NewAddressByCmd================\n")
	fmt.Printf("您正在创建新地址，类型：%s, 数量为：%d\n", t, count)

	// 创建新地址
	accCmd := service.NewAccountCmd("message id", t, uint32(count))
	var aCcs []*types.Account
	aCcs, err = clientManager.NewAccounts(accCmd)
	if err != nil {
		fmt.Printf("创建新地址失败: %s\n", err.Error())
		return nil, errors.New("create address error")
	}
	fmt.Printf("创建新地址成功，数量: %d\n", len(aCcs))

	// 保存新地址
	uniName, err := db.ExportAddress(dataDir, aCcs)
	if err != nil {
		fmt.Printf("导出新地址失败: %s\n", err.Error())
		return nil, errors.New("export address error")
	}
	fmt.Printf("导出新地址成功，标示: %s\n", uniName)

	// 校验一次
	fmt.Printf("开始校验: %s\n", uniName)

	// 校验在线文件
	err = db.VerifyOnlineDBFile(dataDir, uniName, aCcs)
	if err != nil {
		fmt.Printf("校验在线文件失败：%s\n", err.Error())
		return nil, errors.New("check online error")
	}

	// 校验离线文件
	err = db.VerifyOfflineDBFile(dataDir, uniName, aCcs)
	if err != nil {
		fmt.Printf("校验离线文件失败：%s\n", err.Error())
		return nil, errors.New("check offline error")
	}

	fmt.Println("校验完成")

	res := &NewAddressRes{}
	res.UniName = uniName
	res.OnlineDBPath = dataDir + "/" + db.GetOnlineUniDBName(uniName)
	res.OfflineDBPath = dataDir + "/" + db.GetOfflineUniDBName(uniName)

	return res, nil
}

func NewAddress(clientManager *service.ClientManager, dataDir string, argv []string) (string, error) {
	if len(argv) != 3 {
		fmt.Println("正确格式：newaddress 类型 数量")
		return "", errors.New("command error")
	}
	t := argv[1]
	count, err := strconv.Atoi(argv[2])
	if err != nil {
		fmt.Println("错误数量")
		fmt.Println("正确格式：newaddress 类型 数量")
		return "", errors.New("command error")
	}

	cmd := NewAddressCmd{}
	cmd.CoinType = t
	cmd.Count = count
	res, err := NewAddressByCmd(clientManager, dataDir, &cmd)
	if err != nil {
		return "", nil
	}

	resb, err := json.Marshal(res)
	if err != nil {
		return "", nil
	}

	return string(resb), nil
}

func LoadOnlineAddress(dataDir string, argv []string) ([]*types.Account, error) {
	if len(argv) != 2 {
		fmt.Println("正确格式：loadonlineaddress 唯一标示")
		return nil, errors.New("command is error")
	}

	uniName := argv[1]
	uniDBName := db.GetOnlineUniDBName(uniName)
	aCcs, err := db.ImportAddress(dataDir, uniDBName)
	if err != nil {
		fmt.Printf("加载地址失败: %s\n", err.Error())
		return nil, err
	}

	return aCcs, nil
}

func LoadOfflineAddress(dataDir string, argv []string) ([]*types.Account, error) {
	if len(argv) != 2 {
		fmt.Println("正确格式：loadofflineaddress 唯一标示")
		return nil, errors.New("command is error")
	}

	uniName := argv[1]
	uniDBName := db.GetOfflineUniDBName(uniName)
	aCcs, err := db.ImportAddress(dataDir, uniDBName)
	if err != nil {
		fmt.Printf("加载地址失败: %s\n", err.Error())
		return nil, err
	}

	return aCcs, nil
}