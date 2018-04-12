package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	accountdb "../../account_srv/db"
	_ "github.com/go-sql-driver/mysql"
	"../../account_srv/user"
	l4g "github.com/alecthomas/log4go"
)

var (
	Url      = "root@tcp(127.0.0.1:3306)/wallet"
	database string
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"readUserLevel": "SELECT level, is_frozen, public_key from %s.%s where license_key = ? limit ? offset ?",
	}

	st = map[string]*sql.Stmt{}
)

func init() {
	var d *sql.DB
	var err error

	parts := strings.Split(Url, "/")
	if len(parts) != 2 {
		l4g.Crash("Invalid database url")
	}

	if len(parts[1]) == 0 {
		l4g.Crash("Invalid database name")
	}

	url := parts[0]
	database = parts[1]

	if d, err = sql.Open("mysql", url+"/"); err != nil {
		l4g.Crash(err)
	}
	// http://www.01happy.com/golang-go-sql-drive-mysql-connection-pooling/
	d.SetMaxOpenConns(2000)
	d.SetMaxIdleConns(1000)
	d.Ping()

	if _, err := d.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
		l4g.Crash(err)
	}
	d.Close()
	if d, err = sql.Open("mysql", Url); err != nil {
		l4g.Crash(err)
	}
	if _, err = d.Exec(accountdb.UsersSchema); err != nil {
		l4g.Crash(err)
	}

	db = d

	for query, statement := range accountQ {
		prepared, err := db.Prepare(fmt.Sprintf(statement, database, "users"))
		if err != nil {
			l4g.Crash(err)
		}
		st[query] = prepared
	}
}

func ReadUserLevel(licenseKey string) (*user.UserLevel, error) {
	r := st["readUserLevel"].QueryRow(licenseKey, 1, 0)

	ul := &user.UserLevel{}
	if err := r.Scan(&ul.Level, &ul.IsFrozen, &ul.PublicKey); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return ul, nil
}