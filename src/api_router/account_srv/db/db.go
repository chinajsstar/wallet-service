package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	_ "github.com/go-sql-driver/mysql"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
	"bastionpay_api/apibackend/v1/backend"
)

var (
	//Url      = "root:root@tcp(127.0.0.1:3306)/wallet"
	Url      = ""//"root@tcp(127.0.0.1:3306)/wallet"
	database string
	usertable = "user_property"
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"register": `INSERT into %s.%s (
				user_name, user_mobile, user_email,
				user_key, user_class, 
				public_key, source_ip, callback_url, level, is_frozen,
				create_time, update_time) 
				values (?, ?, ?,
				?, ?,
				?, ?, ?, ?, ?,
				?, ?)`,
		"delete": "DELETE from %s.%s where user_key = ?",

		"updateProfile":          "UPDATE %s.%s set public_key = ?, source_ip = ?, callback_url = ?, update_time = ? where user_key = ?",
		"readProfile":         	  "SELECT public_key, source_ip, callback_url from %s.%s where user_key = ?",

		"updateFrozen":           "UPDATE %s.%s set is_frozen = ? where user_key = ?",
		"readFrozen":         	  "SELECT is_frozen from %s.%s where user_key = ?",

		"listUsers":              "SELECT id, user_name, user_mobile, user_email, user_key, user_class, level, is_frozen from %s.%s order by id desc limit ?, ?",
		"listUsersCount":         "SELECT count(*) from %s.%s",
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
	//	l4g.Crashf(err.Error())
	//}
	// do not create db auto
	//if _, err := d.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
	//	l4g.Crashf(err.Error())
	//}
	//d.Close()

	if d, err = sql.Open("mysql", Url); err != nil {
		l4g.Crashf(err.Error())
	}
	// http://www.01happy.com/golang-go-sql-drive-mysql-connection-pooling/
	d.SetMaxOpenConns(2000)
	d.SetMaxIdleConns(1000)
	d.Ping()
	// do not create table auto
	//if _, err = d.Exec(UsersSchema); err != nil {
	//	l4g.Crashf(err.Error())
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

func Register(userRegister *backend.ReqUserRegister, userKey string) error {
	var datetime = time.Now().UTC()
	datetime.Format(time.RFC3339)
	_, err := st["register"].Exec(
		userRegister.UserName, userRegister.UserMobile, userRegister.UserEmail,
		userKey, userRegister.UserClass,
		"", "", "", userRegister.Level, userRegister.IsFrozen,
		datetime, datetime)
	return err
}

func Delete(userKey string) error {
	_, err := st["delete"].Exec(userKey)
	return err
}

func UpdateProfile(subUserKey string, userUpdateProfile *backend.ReqUserUpdateProfile) error {
	var datetime = time.Now().UTC()
	datetime.Format(time.RFC3339)
	_, err := st["updateProfile"].Exec(userUpdateProfile.PublicKey, userUpdateProfile.SourceIP, userUpdateProfile.CallbackUrl, datetime, subUserKey)
	return err
}

func ReadProfile(userKey string) (*backend.AckUserReadProfile, error) {
	var r *sql.Rows
	var err error

	r, err = st["readProfile"].Query(userKey)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if !r.Next() {
		return nil, errors.New("row no next")
	}

	ackUserReadProfile := &backend.AckUserReadProfile{}
	if err := r.Scan(&ackUserReadProfile.PublicKey, &ackUserReadProfile.SourceIP, &ackUserReadProfile.CallbackUrl); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no rows")
		}
		return nil, err
	}
	if r.Err() != nil {
		return nil, err
	}

	ackUserReadProfile.UserKey = userKey
	return ackUserReadProfile, nil
}

func UpdateFrozen(userKey string, frozen int) error {
	_, err := st["updateFrozen"].Exec(frozen, userKey)
	return err
}

func ReadFrozen(userKey string) (int, error) {
	var r *sql.Rows
	var err error
	var isFrozen int

	r, err = st["readFrozen"].Query(userKey)
	if err != nil {
		return -1, err
	}
	defer r.Close()

	if !r.Next() {
		return -1, errors.New("row no next")
	}

	if err := r.Scan(&isFrozen); err != nil {
		if err == sql.ErrNoRows {
			return -1, errors.New("no rows")
		}
		return -1, err
	}
	if r.Err() != nil {
		return -1, err
	}

	return isFrozen, nil
}

func ListUsers(beginIndex int, pageNum int) (*backend.AckUserList, error) {
	var r *sql.Rows
	var err error

	r, err = st["listUsers"].Query(beginIndex, pageNum)

	if err != nil {
		return nil, err
	}
	defer r.Close()

	ul := &backend.AckUserList{}
	for r.Next()  {
		up := backend.UserBasic{}
		if err := r.Scan(&up.Id, &up.UserName, &up.UserMobile, &up.UserEmail, &up.UserKey, &up.UserClass, &up.Level, &up.IsFrozen); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			continue
		}

		ul.Data = append(ul.Data, up)
	}

	if r.Err() != nil {
		return nil, err
	}

	return ul, nil
}

func ListUserCount() (int, error) {
	var r *sql.Rows
	var err error

	r, err = st["listUsersCount"].Query()
	if err != nil {
		return 0, err
	}
	defer r.Close()

	if !r.Next() {
		return 0, errors.New("row no next")
	}

	var count int
	if err := r.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("no rows")
		}
		return 0, err
	}
	if r.Err() != nil {
		return 0, err
	}

	return count, nil
}

func buildUserBasicCondition(conds map[string]interface{}) (string, []interface{}) {
	var statement string
	var condition []interface{}

	for k, v := range conds {
		statement += " and " + k + " = ?"
		condition = append(condition, v)
	}

	wholeStatement := ""
	if statement != ""{
		wholeStatement = " where true"
		wholeStatement += statement
	}

	return wholeStatement, condition
}

func ListUserCountByBasic(conds map[string]interface{}) (int, error) {
	var r *sql.Rows
	var err error

	basestatement := "SELECT count(*) from %s.%s"
	statement, conditions := buildUserBasicCondition(conds)
	basestatement += statement

	prepared, err := db.Prepare(fmt.Sprintf(basestatement, database, usertable))
	if err != nil {
		return 0, err
	}

	r, err = prepared.Query(conditions...)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	if !r.Next() {
		return 0, errors.New("row no next")
	}

	var count int
	if err := r.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("no rows")
		}
		return 0, err
	}
	if r.Err() != nil {
		return 0, err
	}

	return count, nil
}

func ListUsersByBasic(beginIndex int, pageNum int, conds map[string]interface{}) (*backend.AckUserList, error) {
	var r *sql.Rows
	var err error

	basestatement := "SELECT id, user_name, user_mobile, user_email, user_key, user_class, level, is_frozen from %s.%s"
	statement, conditions := buildUserBasicCondition(conds)
	basestatement += statement
	basestatement += " order by id desc limit ?, ?"
	conditions = append(conditions, beginIndex)
	conditions = append(conditions, pageNum)

	fmt.Println(basestatement)
	fmt.Println(conditions...)

	prepared, err := db.Prepare(fmt.Sprintf(basestatement, database, usertable))
	if err != nil {
		return nil, err
	}

	r, err = prepared.Query(conditions...)

	//r, err = st["listUsers"].Query(beginIndex, pageNum)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	ul := &backend.AckUserList{}
	for r.Next()  {
		up := backend.UserBasic{}
		if err := r.Scan(&up.Id, &up.UserName, &up.UserMobile, &up.UserEmail, &up.UserKey, &up.UserClass, &up.Level, &up.IsFrozen); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			continue
		}

		ul.Data = append(ul.Data, up)
	}

	if r.Err() != nil {
		return nil, err
	}

	return ul, nil
}