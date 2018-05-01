package db

import (
	"fmt"
	"log"
	"errors"
	"api_router/base/utils"
	"blockchain_server/types"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	l4g "github.com/alecthomas/log4go"
)

const tableAccount = `CREATE TABLE account (
 id INTEGER,
 address varchar(255) NOT NULL,
 chiperprikey text NOT NULL,
 primary key(id, address)
);`

const (
	OnlineTag = "-online"
	OfflineTag = "-offline"
)

type DBInstance struct{
	db *sql.DB
}

func Open(dataDir string, uniDBName string) (*DBInstance, error) {
	dbPath := dataDir + "/" + uniDBName
	fmt.Println("正在读取地址信息：%s", dbPath)

	bl, err := utils.PathExists(dbPath)
	if bl == false{
		l4g.Error("文件不存在: %s", dbPath)
		return nil, errors.New("path not exist")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		l4g.Error("打开DB文件失败：%s", err.Error())
		return nil, err
	}

	dbi := &DBInstance{db:db}
	return dbi, nil
}
func (dbi *DBInstance)Close()  {
	dbi.db.Close()
}

func (dbi *DBInstance)QueryAddress(indexs []int) (map[string]*types.Account, error) {
	stmt, err := dbi.db.Prepare("select address, chiperprikey from account where id = ?")
	if err != nil {
		l4g.Error("Prepare DB失败：%s", err.Error())
		return nil, err
	}
	defer stmt.Close()

	accsMap := make(map[string]*types.Account)
	for _, index := range indexs {
		acc := &types.Account{}
		err = stmt.QueryRow(index).Scan(&acc.Address, &acc.PrivateKey)
		if err != nil {
			l4g.Error("QueryRow失败(%d)：%s", index, err.Error())
			return nil, err
		}

		accsMap[acc.Address] = acc
	}

	return accsMap, nil
}

func GetOnlineUniDBName(uniName string) string {
	return uniName + OnlineTag + ".db"
}

func GetOfflineUniDBName(uniName string) string {
	return uniName + OfflineTag + ".db"
}

func ExportAddress(dataDir string, uniName string, accs []*types.Account) (error) {
	// online and offline path
	dbOnlinePath := dataDir + "/" + GetOnlineUniDBName(uniName)
	dbOfflinePath := dataDir + "/" + GetOfflineUniDBName(uniName)
	fmt.Printf("在线DB文件名：%s\n", dbOnlinePath)
	fmt.Printf("离线DB文件名：%s\n", dbOfflinePath)

	bl, _ := utils.PathExists(dbOnlinePath)
	bl2, _ := utils.PathExists(dbOfflinePath)
	if bl || bl2 {
		l4g.Error("DB文件已经存在: %s-%s", dbOnlinePath, dbOfflinePath)
		return errors.New("path exist")
	}

	// db func
	dbFunc := func(path string, onLine bool) error {
		// open db
		db, err := sql.Open("sqlite3", path)
		if err != nil {
			l4g.Error("打开DB失败：%s", err.Error())
			return err
		}
		defer db.Close()

		// create table
		_, err = db.Exec(tableAccount)
		if err != nil {
			l4g.Error("创建DB table失败 %q: %s", err, tableAccount)
			return err
		}

		// begin
		tx, err := db.Begin()
		if err != nil {
			fmt.Println(err)
			return err
		}
		stmt, err := tx.Prepare("insert into account(id, address, chiperprikey) values(?, ?, ?)")
		if err != nil {
			log.Println(err)
			return err
		}
		defer stmt.Close()

		for i, account := range accs {
			if onLine {
				_, err = stmt.Exec(i, account.Address, fmt.Sprintf("%s", uniName))
				//_, err = stmt.Exec(i, account.Address, fmt.Sprintf("%s@%d", uniName, i))
			}else{
				_, err = stmt.Exec(i, account.Address, account.PrivateKey)
			}

			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		// commit
		return tx.Commit()
	}

	if err := dbFunc(dbOnlinePath, true); err != nil{
		return err
	}
	fmt.Printf("保存在线DB成功\n")

	if err := dbFunc(dbOfflinePath, false); err != nil{
		return err
	}
	fmt.Printf("保存离线DB成功\n")

	return nil
}

func ImportAddress(dataDir string, uniDBName string) ([]*types.Account, error) {
	dbPath := dataDir + "/" + uniDBName
	fmt.Println("正在读取地址信息：", dbPath)

	bl, err := utils.PathExists(dbPath)
	if bl == false{
		l4g.Error("文件不存在: %s", dbPath)
		return nil, errors.New("path not exist")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		l4g.Error("打开DB文件失败：%s", err.Error())
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select address, chiperprikey from account")
	if err != nil {
		l4g.Error("查询地址失败：%s", err.Error())
		return nil, err
	}
	defer rows.Close()

	var accs []*types.Account
	for rows.Next() {
		acc := &types.Account{}
		err = rows.Scan(&acc.Address, &acc.PrivateKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		accs = append(accs, acc)
	}

	err = rows.Err()
	if err != nil {
		l4g.Error("读取地址失败：%s", err.Error())
		return nil, err
	}

	fmt.Println("读取地址完成")
	return accs, nil
}

func ImportAddressMap(dataDir string, uniDBName string) (map[string]*types.Account, error) {
	dbPath := dataDir + "/" + uniDBName
	fmt.Println("正在读取地址信息：", dbPath)

	bl, err := utils.PathExists(dbPath)
	if bl == false{
		l4g.Error("文件不存在: %s", dbPath)
		return nil, errors.New("path not exist")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		l4g.Error("打开DB文件失败：%s", err.Error())
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select address, chiperprikey from account")
	if err != nil {
		l4g.Error("查询地址失败：%s", err.Error())
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
		l4g.Error("读取地址失败：%s", err.Error())
		return nil, err
	}

	fmt.Println("读取地址完成: ", len(accsMap))
	return accsMap, nil
}

// 校验在线文件
func VerifyOnlineDBFile(dataDir string, uniName string, oriAccs []*types.Account) (error) {
	uniOnlineDBName := GetOnlineUniDBName(uniName)
	fmt.Printf("校验在线文件: %s\n", uniOnlineDBName)
	aCcsOnline, err := ImportAddress(dataDir, uniOnlineDBName)
	if err != nil {
		fmt.Printf("校验在线文件失败: %s\n", err.Error())
		return err
	}

	if len(oriAccs) != len(aCcsOnline) {
		fmt.Printf("校验在线文件失败，数量不一致: %d-%d\n", len(oriAccs), len(aCcsOnline))
		return errors.New("数量不一致")
	}

	for _, acc := range aCcsOnline {
		if acc.PrivateKey != uniName {
			fmt.Printf("校验在线文件失败，地址对应的标示不正确: %s-%s\n", acc.PrivateKey, uniName)
			return errors.New("地址标示不对应")
		}
	}
	fmt.Println("校验在线文件完成")
	return nil
}

// 校验离线文件
func VerifyOfflineDBFile(dataDir string, uniName string, oriAccs []*types.Account) (error) {
	oriAccsMap := make(map[string]string)
	for _, oriAcc := range oriAccs {
		oriAccsMap[oriAcc.Address] = oriAcc.PrivateKey
	}

	uniOfflineDBName := GetOfflineUniDBName(uniName)
	fmt.Printf("校验离线文件: %s\n", uniOfflineDBName)
	aCcsOffline, err := ImportAddress(dataDir, uniOfflineDBName)
	if err != nil {
		fmt.Printf("校验离线文件失败: %s\n", err.Error())
		return err
	}

	if len(oriAccs) != len(aCcsOffline) {
		fmt.Printf("校验离线文件失败，数量不一致: %d-%d\n", len(oriAccs), len(aCcsOffline))
		return errors.New("数量不一致")
	}

	for _, acc := range aCcsOffline {
		if pk, ok := oriAccsMap[acc.Address]; !ok || pk != acc.PrivateKey {
			fmt.Printf("校验离线文件失败，地址对应的加密私钥不正确: %s-%s\n", acc.PrivateKey, pk)
			return errors.New("地址加密私钥不对应")
		}
	}
	fmt.Println("校验离线文件完成")
	return nil
}