package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"../user"
	_ "github.com/go-sql-driver/mysql"
	l4g "github.com/alecthomas/log4go"
	"../../base/config"
)

var (
	//Url      = "root:root@tcp(127.0.0.1:3306)/wallet"
	Url      = ""//"root@tcp(127.0.0.1:3306)/wallet"
	database string
	usertable = "user_property"
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"create": `INSERT into %s.%s (
				user_key, user_name, user_class, phone, email, 
				salt, password, google_auth, 
				public_key, callback_url, level,
				last_login_time, last_login_ip, last_login_mac,
				create_time, update_time, is_frozen,
				time_zone, country, language) 
				values (?, ?, ?, ?, ?,
				?, ?, ?,
				?, ?, ?, 
				?, ?, ?,
				?, ?, ?,
				?, ?, ?)`,
		"delete": "DELETE from %s.%s where user_key = ?",

		"updatePassword":         "UPDATE %s.%s set salt = ?, password = ?, update_time = ? where user_key = ?",
		"frozen":         	  	  "UPDATE %s.%s set is_frozen = ? where user_key = ?",
		"level":         	      "UPDATE %s.%s set level = ? where user_key = ?",
		"readProfile":            "SELECT user_key, user_name, phone, email from %s.%s where user_key = ?",
		"readPassword":           "SELECT salt, password from %s.%s where user_key = ?",
		"searchUsername":         "SELECT user_key, user_name, phone, email, salt, password from %s.%s where user_name = ? limit ? offset ?",
		"searchPhone":         	  "SELECT user_key, user_name, phone, email, salt, password from %s.%s where phone = ? limit ? offset ?",
		"searchEmail":            "SELECT user_key, user_name, phone, email, salt, password from %s.%s where email = ? limit ? offset ?",

		"listUsers":              "SELECT id, user_key, user_name, user_class, phone, email from %s.%s where id < ? order by id desc limit ?",
		"listUsers2":             "SELECT id, user_key, user_name, user_class, phone, email from %s.%s order by id desc limit ?",
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

func Create(user *user.ReqUserCreate, userKey string, salt string, password string) error {
	var datetime = time.Now().UTC()
	datetime.Format(time.RFC3339)
	_, err := st["create"].Exec(
		userKey, user.UserName, user.UserClass, user.Phone, user.Email,
		salt, password, user.GoogleAuth,
		user.PublicKey, user.CallbackUrl, user.Level,
		datetime, "", "",
		datetime, datetime, 0,
		user.TimeZone, user.Country, user.Language)
	return err
}

func Delete(userKey string) error {
	_, err := st["delete"].Exec(userKey)
	return err
}

func UpdatePassword(userKey string, salt string, password string) error {
	var datetime = time.Now().UTC()
	datetime.Format(time.RFC3339)
	_, err := st["updatePassword"].Exec(salt, password, datetime, userKey)
	return err
}

func Frozen(userKey string, frozen rune) error {
	_, err := st["frozen"].Exec(frozen, userKey)
	return err
}

func Level(userKey string, level int) error {
	_, err := st["level"].Exec(level, userKey)
	return err
}

func ReadPassword(userName, phone, email string) (*user.AckUserLogin, string, string, error) {
	var r *sql.Rows
	var err error

	//phoneReg := `^((\+86)|(86))?(13)\d{9}$`
	//emailReg := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	//phoneRgx := regexp.MustCompile(phoneReg)
	//emailRgx := regexp.MustCompile(emailReg)

	if len(userName) > 0 {
		r, err = st["searchUsername"].Query(userName, 1, 0)
	} else if len(phone) > 0 {
		r, err = st["searchPhone"].Query(phone, 1, 0)
	} else if len(email) > 0 {
		r, err = st["searchEmail"].Query(email, 1, 0)
	} else {
		return nil, "", "", errors.New("username, phone or email cannot be blank")
	}

	if err != nil {
		return nil, "", "", err
	}
	defer r.Close()

	if !r.Next() {
		return nil, "", "", errors.New("not found")
	}

	var salt, pass string
	user := &user.AckUserLogin{}
	if err := r.Scan(&user.UserKey, &user.UserName, &user.Phone, &user.Email, &salt, &pass); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", "", errors.New("not found")
		}
		return nil, "", "", err
	}
	if r.Err() != nil {
		return nil, "", "", err
	}

	return user, salt, pass, nil
}

func ListUsers(id int, num int) (*user.AckUserList, error) {
	var r *sql.Rows
	var err error

	if id < 0 {
		r, err = st["listUsers2"].Query(num)
	}else{
		r, err = st["listUsers"].Query(id, num)
	}

	if err != nil {
		return nil, err
	}
	defer r.Close()

	ul := &user.AckUserList{}
	for r.Next()  {
		up := user.UserProfile{}
		if err := r.Scan(&up.Id, &up.UserKey, &up.UserName, &up.UserClass, &up.Phone, &up.Email); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			continue
		}

		ul.Users = append(ul.Users, up)
	}

	if r.Err() != nil {
		return nil, err
	}

	return ul, nil
}