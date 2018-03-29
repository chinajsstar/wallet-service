package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	userdb "../../user_srv/db"
	_ "github.com/go-sql-driver/mysql"
)

var (
	Url      = "root@tcp(127.0.0.1:3306)/wallet"
	database string
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"readPubKey": "SELECT public_key from %s.%s where license_key = ? limit ? offset ?",
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
	if _, err = d.Exec(userdb.UsersSchema); err != nil {
		log.Fatal(err)
	}

	db = d

	for query, statement := range accountQ {
		prepared, err := db.Prepare(fmt.Sprintf(statement, database, "users"))
		if err != nil {
			log.Fatal(err)
		}
		st[query] = prepared
	}
}

func ReadPubKey(licenseKey string) (string, error) {
	r := st["readPubKey"].QueryRow(licenseKey, 1, 0)

	pubKey := ""
	if err := r.Scan(&pubKey); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("not found")
		}
		return "", err
	}

	return pubKey, nil
}