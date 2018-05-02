package db

import (
	"errors"
	"api_router/base/utils"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	l4g "github.com/alecthomas/log4go"
)

type DBInstance struct{
	Db *sql.DB
}

func (dbi *DBInstance)Open(dataDir string, uniDBName string) (error) {
	dbPath := dataDir + "/" + uniDBName

	bl, err := utils.PathExists(dbPath)
	if bl == false{
		l4g.Error("文件不存在: %s", dbPath)
		return errors.New("path not exist")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		l4g.Error("打开DB文件失败：%s", err.Error())
		return err
	}

	dbi.Db = db
	return nil
}

func (dbi *DBInstance)Close()  {
	dbi.Db.Close()
}