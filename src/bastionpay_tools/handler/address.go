package handler

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/db"
	//l4g "github.com/alecthomas/log4go"
	"fmt"
	"bastionpay_tools/common"
	"os"
)

func newMaxCountAddress(clientManager *service.ClientManager, coinType string, count uint32) ([]*types.Account, error) {
	var aCcsAll []*types.Account

	left := count

	var realCount uint32
	for left > 0 {
		realCount = common.MaxCountAddress
		if left < common.MaxCountAddress {
			realCount = left
		}

		accCmd := service.NewAccountCmd("message id", coinType, uint32(realCount))
		aCcs, err := clientManager.NewAccounts(accCmd)
		if err != nil {
			return nil, err
		}
		aCcsAll = append(aCcsAll, aCcs...)

		left -= realCount
	}

	return aCcsAll, nil
}

// 返回唯一标示
func NewAddress(clientManager *service.ClientManager, addrDir string, uniAddrInfo *common.UniAddressInfo, count uint32)(*common.UniAddressDbInfo, error) {
	var err error

	// 生成唯一
	uniAddrDbInfo, err := common.NewUniAddressDbInfo()
	if err != nil {
		return nil, err
	}

	// 生成离线
	fmt.Println("生成离线地址...")
	aCcsOffline, err := newMaxCountAddress(clientManager, uniAddrInfo.CoinType, count)
	if err != nil {
		return nil, err
	}

	// 保存离线
	fmt.Println("保存离线地址...")
	offlineTmpPath := uniAddrInfo.GetUniAbsDir(addrDir) + "/" + "offline.db"
	uniAddrDbInfo.OfflineMd5, err = db.ExportLineAddress(offlineTmpPath, aCcsOffline)
	if err != nil {
		return nil, err
	}

	// 生成在线
	fmt.Println("导出在线地址...")
	var aCcsOnline []*types.Account
	for _, acc := range aCcsOffline{
		accOnline := &types.Account{Address:acc.Address, PrivateKey:uniAddrInfo.GetUniName() + "@" + uniAddrDbInfo.GetUniNameOffline()}
		aCcsOnline = append(aCcsOnline, accOnline)
	}

	// 保存在线
	fmt.Println("保存在线地址...")
	onlineTmpPath := uniAddrInfo.GetUniAbsDir(addrDir) + "/" + "online.db"
	uniAddrDbInfo.OnlineMd5, err = db.ExportLineAddress(onlineTmpPath, aCcsOnline)
	if err != nil {
		return nil, err
	}

	// 校验离线
	fmt.Println("校验离线地址...")
	offlineDbMd5_1, err := common.GetSaltMd5HexByFile(offlineTmpPath)
	if err != nil {
		return nil, err
	}
	if uniAddrDbInfo.OfflineMd5 != offlineDbMd5_1 {
		fmt.Println("离线地址MD5错误...")
		return nil, err
	}

	err = db.VerifyLineAddress(offlineTmpPath, aCcsOffline)
	if err != nil {
		fmt.Println("离线地址私钥有错误...")
		return nil, err
	}

	// 校验在线
	fmt.Println("校验在线地址...")
	onlineDbMd5_1, err := common.GetSaltMd5HexByFile(onlineTmpPath)
	if err != nil {
		return nil, err
	}
	if uniAddrDbInfo.OnlineMd5 != onlineDbMd5_1 {
		fmt.Println("在线地址MD5错误...")
		return nil, err
	}

	err = db.VerifyLineAddress(onlineTmpPath, aCcsOnline)
	if err != nil {
		fmt.Println("在线地址私钥有错误...")
		return nil, err
	}

	fmt.Println("校验成功")

	// 重命名
	fmt.Println("重命名")
	uniAbsDir := uniAddrInfo.GetUniAbsDir(addrDir)
	realOnlineTmpPath := uniAbsDir + "/" + uniAddrInfo.GetUniName() + "@" + uniAddrDbInfo.GetUniNameOnline() + "@" + common.GetOnlineDbNameSuffix()
	realOfflineTmpPath := uniAbsDir + "/" + uniAddrInfo.GetUniName() + "@" + uniAddrDbInfo.GetUniNameOffline() + "@" + common.GetOfflineDbNameSuffix()

	err = os.Rename(onlineTmpPath, realOnlineTmpPath)
	if err != nil {
		fmt.Printf("重命名在线文件失败：%s\n", err.Error())
		return nil, err
	}

	err = os.Rename(offlineTmpPath, realOfflineTmpPath)
	if err != nil {
		fmt.Printf("重命名离线文件失败：%s\n", err.Error())
		return nil, err
	}

	fmt.Printf("==============End NewAddress================\n")
	return uniAddrDbInfo, nil
}

func LoadAddress(uniAbsDir string, uniDbName string) ([]*types.Account, error) {
	aCcs, err := db.LoadAddress(uniAbsDir + "/" + uniDbName)
	if err != nil {
		return nil, err
	}

	return aCcs, nil
}