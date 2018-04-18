package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"api_router/account_srv/user"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
)

var (
	Url      = ""//"root@tcp(127.0.0.1:3306)/wallet"
	database string
	usertable = "user_property"
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"readUserLevel": "SELECT level, is_frozen, public_key from %s.%s where user_key = ? limit ? offset ?",
	}

	st = map[string]*sql.Stmt{}
)

func Init(configPath string) {
	var d *sql.DB
	var err error

	err = config.LoadJsonNode(configPath, "db", &Url)
	if err != nil {
		l4g.Crashf("", err)
	}

	parts := strings.Split(Url, "/")
	if len(parts) != 2 {
		l4g.Crashf("Invalid database url")
	}

	if len(parts[1]) == 0 {
		l4g.Crashf("Invalid database name")
	}

	//url := parts[0]
	database = parts[1]

	//if d, err = sql.Open("mysql", url+"/"); err != nil {
	//	l4g.Crashf(err)
	//}
	//if _, err := d.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
	//	l4g.Crashf(err)
	//}
	//d.Close()
	if d, err = sql.Open("mysql", Url); err != nil {
		l4g.Crashf("", err)
	}
	// http://www.01happy.com/golang-go-sql-drive-mysql-connection-pooling/
	d.SetMaxOpenConns(2000)
	d.SetMaxIdleConns(1000)
	d.Ping()
	//if _, err = d.Exec(accountdb.UsersSchema); err != nil {
	//	l4g.Crash(err)
	//}

	db = d

	for query, statement := range accountQ {
		prepared, err := db.Prepare(fmt.Sprintf(statement, database, usertable))
		if err != nil {
			l4g.Crashf("", err)
		}
		st[query] = prepared
	}
}

func ReadUserLevel(userKey string) (*user.UserLevel, error) {
	r := st["readUserLevel"].QueryRow(userKey, 1, 0)

	ul := &user.UserLevel{}
	if err := r.Scan(&ul.Level, &ul.IsFrozen, &ul.PublicKey); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return ul, nil
}