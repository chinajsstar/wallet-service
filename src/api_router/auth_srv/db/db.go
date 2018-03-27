package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"../user"
	_ "github.com/go-sql-driver/mysql"
)

var (
	Url      = "root@tcp(127.0.0.1:3306)/wallet"
	database string
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"delete": "DELETE from %s.%s where licensekey = ?",
		"create": `INSERT into %s.%s (
				licensekey, username, pubkey, created) 
				values (?, ?, ?, ?)`,
		"read": "SELECT licensekey, username, pubkey, created from %s.%s where licensekey = ?",
	}

	st = map[string]*sql.Stmt{}
)

func Init() {
	var d *sql.DB
	var err error

	parts := strings.Split(Url, "/")
	if len(parts) != 2 {
		panic("Invalid database url")
	}

	if len(parts[1]) == 0 {
		panic("Invalid database name")
	}

	url := parts[0]
	database = parts[1]

	if d, err = sql.Open("mysql", url+"/"); err != nil {
		log.Fatal(err)
	}
	if _, err := d.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
		log.Fatal(err)
	}
	d.Close()
	if d, err = sql.Open("mysql", Url); err != nil {
		log.Fatal(err)
	}
	if _, err = d.Exec(accountSchema); err != nil {
		log.Fatal(err)
	}

	db = d

	for query, statement := range accountQ {
		prepared, err := db.Prepare(fmt.Sprintf(statement, database, "accounts"))
		if err != nil {
			log.Fatal(err)
		}
		st[query] = prepared
	}
}

func Create(user *user.User) error {
	user.Created = time.Now().Unix()
	_, err := st["create"].Exec(user.LicenseKey, user.Username, user.PubKey, user.Created)
	return err
}

func Delete(licensekey string) error {
	_, err := st["delete"].Exec(licensekey)
	return err
}

func Read(licensekey string) (*user.User, error) {
	user := &user.User{}

	r := st["read"].QueryRow(licensekey)
	if err := r.Scan(&user.LicenseKey, &user.Username, &user.PubKey, &user.Created); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return user, nil
}