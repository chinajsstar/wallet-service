package db

import (
	"fmt"
	"bastionpay_tools/common"
	"bastionpay_api/utils"
	"blockchain_server/types"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	//l4g "github.com/alecthomas/log4go"
)

const tableAccount = `CREATE TABLE account (
 id INTEGER,
 address varchar(255) NOT NULL,
 chiperprikey text NOT NULL,
 primary key(id, address)
);`

func SaveAddress(path string, aCcs []*types.Account) error {
	// open db
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	defer db.Close()

	// create table
	_, err = db.Exec(tableAccount)
	if err != nil {
		return err
	}

	// begin
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into account(id, address, chiperprikey) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, account := range aCcs {
		_, err = stmt.Exec(i, account.Address, account.PrivateKey)
		if err != nil {
			return err
		}
	}
	// commit
	return tx.Commit()
}

func LoadAddress(path string) ([]*types.Account, error) {
	dbExist, err := utils.PathExists(path)
	if dbExist == false{
		return nil, fmt.Errorf("path(%s) not exist", path)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select address, chiperprikey from account")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aCcs []*types.Account
	for rows.Next() {
		acc := &types.Account{}
		err = rows.Scan(&acc.Address, &acc.PrivateKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		aCcs = append(aCcs, acc)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return aCcs, nil
}

func LoadAddressMap(path string) (map[string]*types.Account, error) {
	dbExist, err := utils.PathExists(path)
	if dbExist == false{
		return nil, fmt.Errorf("path(%s) not exist", path)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select address, chiperprikey from account")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accsMap := make(map[string]*types.Account)
	for rows.Next() {
		acc := &types.Account{}
		err = rows.Scan(&acc.Address, &acc.PrivateKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		accsMap[acc.Address] = acc
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return accsMap, nil
}

func ExportLineAddress(path string, accs []*types.Account) (string, error) {
	// check path
	pathExist, _ := utils.PathExists(path)
	if pathExist {
		return "", fmt.Errorf("path(%s) already exist", path)
	}

	// save to path
	err := SaveAddress(path, accs)
	if err != nil {
		return "", err
	}

	// get path md5
	md5, err := common.GetSaltMd5HexByFile(path)
	if err != nil {
		return "", err
	}

	return md5, nil
}

func VerifyLineAddress(path string, oriAccs []*types.Account) (error) {
	aCcs, err := LoadAddress(path)
	if err != nil {
		return err
	}

	if len(oriAccs) != len(aCcs) {
		return fmt.Errorf("count %d-%d is not equal", len(aCcs), len(oriAccs))
	}

	// store to map
	oriAccsMap := make(map[string]string)
	for _, oriAcc := range oriAccs {
		oriAccsMap[oriAcc.Address] = oriAcc.PrivateKey
	}

	for _, acc := range aCcs {
		if pk, ok := oriAccsMap[acc.Address]; !ok || pk != acc.PrivateKey {
			return fmt.Errorf("prikey(%s-%s) is not equal", acc.PrivateKey, pk)
		}
	}

	return nil
}
//
//func ExportAddress(addressDataDir string, uai *common.UniAddressInfo, tmpAddressDbInfo *common.AddressDbInfo, accs []*types.Account) (*common.AddressDbInfo, error) {
//	// online and offline path
//	uniAbsDir := uai.GetUniAbsDir(addressDataDir)
//	tmpDbOnlinePath := uniAbsDir + "/" + uai.GetUniNameAtMd5(tmpAddressDbInfo.OnlineDbMd5) + common.GetOnlineDbNameSuffix()
//	tmpDbOfflinePath := uniAbsDir + "/" + uai.GetUniNameAtMd5(tmpAddressDbInfo.OfflineDbMd5) + common.GetOfflineDbNameSuffix()
//	fmt.Printf("临时在线DB文件名：%s\n", tmpDbOnlinePath)
//	fmt.Printf("临时离线DB文件名：%s\n", tmpDbOfflinePath)
//
//	bl, _ := utils.PathExists(tmpDbOnlinePath)
//	bl2, _ := utils.PathExists(tmpDbOfflinePath)
//	if bl || bl2 {
//		l4g.Error("DB文件已经存在: %s-%s", tmpDbOnlinePath, tmpDbOfflinePath)
//		return nil, errors.New("path exist")
//	}
//
//	// db func
//	dbFunc := func(path string, onLine bool, priKey string) error {
//		// open db
//		db, err := sql.Open("sqlite3", path)
//		if err != nil {
//			l4g.Error("打开DB失败：%s", err.Error())
//			return err
//		}
//		defer db.Close()
//
//		// create table
//		_, err = db.Exec(tableAccount)
//		if err != nil {
//			l4g.Error("创建DB table失败 %q: %s", err, tableAccount)
//			return err
//		}
//
//		// begin
//		tx, err := db.Begin()
//		if err != nil {
//			fmt.Println(err)
//			return err
//		}
//		stmt, err := tx.Prepare("insert into account(id, address, chiperprikey) values(?, ?, ?)")
//		if err != nil {
//			log.Println(err)
//			return err
//		}
//		defer stmt.Close()
//
//		for i, account := range accs {
//			if onLine {
//				_, err = stmt.Exec(i, account.Address, fmt.Sprintf("%s", priKey))
//				//_, err = stmt.Exec(i, account.Address, fmt.Sprintf("%s@%d", uniName, i))
//			}else{
//				_, err = stmt.Exec(i, account.Address, account.PrivateKey)
//			}
//
//			if err != nil {
//				fmt.Println(err)
//				return err
//			}
//		}
//
//		// commit
//		return tx.Commit()
//	}
//
//	// 先保存offline
//	if err := dbFunc(tmpDbOfflinePath, false, ""); err != nil{
//		return nil, err
//	}
//	fmt.Printf("保存离线DB成功\n")
//
//	offlineDbMd5, err := common.GetSaltMd5HexByFile(tmpDbOfflinePath)
//	if err != nil {
//		return nil, err
//	}
//	offlineUniName := uai.GetUniNameAtMd5(offlineDbMd5)
//
//	// 在保存online
//	if err := dbFunc(tmpDbOnlinePath, true, offlineUniName); err != nil{
//		return nil, err
//	}
//
//	fmt.Printf("保存在线DB成功\n")
//	onlineDbMd5, err := common.GetSaltMd5HexByFile(tmpDbOnlinePath)
//	if err != nil {
//		return nil, err
//	}
//
//	realAddressDbInfo := &common.AddressDbInfo{OfflineDbMd5:offlineDbMd5, OnlineDbMd5:onlineDbMd5}
//	return realAddressDbInfo, nil
//}