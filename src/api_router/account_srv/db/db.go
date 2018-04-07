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
	//Url      = "root:root@tcp(127.0.0.1:3306)/wallet"
	Url      = "root@tcp(127.0.0.1:3306)/wallet"
	database string
	db       *sql.DB

	q = map[string]string{}

	accountQ = map[string]string{
		"delete": "DELETE from %s.%s where id = ?",
		"deleteByLicenseKey": "DELETE from %s.%s where license_key = ?",
		"create": `INSERT into %s.%s (
				user_name, phone, email, 
				salt, password, google_auth, 
				license_key, public_key, level,
				last_login_time, last_login_ip, last_login_mac,
				create_time, update_time,
				time_zone, country, language) 
				values (?, ?, ?, 
				?, ?, ?,
				?, ?, ?,
				?, ?, ?,
				?, ?,
				?, ?, ?)`,
		"updateProfile":          "UPDATE %s.%s set user_name = ?, phone = ?, email = ?, update_time = ? where id = ?",
		"updatePassword":         "UPDATE %s.%s set salt = ?, password = ?, update_time = ? where id = ?",
		"frozen":         	  	  "UPDATE %s.%s set is_frozen = ? where id = ?",
		"level":         	      "UPDATE %s.%s set level = ? where id = ?",
		"readProfile":            "SELECT id, license_key, user_name, phone, email from %s.%s where id = ?",
		"readPassword":           "SELECT salt, password from %s.%s where id = ?",
		"searchUsername":         "SELECT id, license_key, user_name, phone, email, salt, password from %s.%s where user_name = ? limit ? offset ?",
		"searchPhone":         	  "SELECT id, license_key, user_name, phone, email, salt, password from %s.%s where phone = ? limit ? offset ?",
		"searchEmail":            "SELECT id, license_key, user_name, phone, email, salt, password from %s.%s where email = ? limit ? offset ?",

		"listUsers":              "SELECT id, license_key, user_name, phone, email from %s.%s where id < ? order by id desc limit ?",
		"listUsers2":             "SELECT id, license_key, user_name, phone, email from %s.%s order by id desc limit ?",
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
	if _, err = d.Exec(UsersSchema); err != nil {
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

func Create(user *user.ReqUserCreate, licenseKey string, salt string, password string) error {
	var datetime = time.Now().Local()
	datetime.Format(time.RFC3339)
	_, err := st["create"].Exec(
		user.UserName, user.Phone, user.Email,
		salt, password, user.GoogleAuth,
		licenseKey, user.PublicKey, user.Level,
		datetime, "", "",
		datetime, datetime,
		user.TimeZone, user.Country, user.Language)
	return err
}

func Delete(id int) error {
	_, err := st["delete"].Exec(id)
	return err
}

func DeleteByLicenseKey(licenseKey string) error {
	_, err := st["deleteByLicenseKey"].Exec(licenseKey)
	return err
}

func UpdateProfile(id int, userName string, phone string, email string) error {
	_, err := st["updateProfile"].Exec(userName, phone, email, time.Now().Unix(), id)
	return err
}

func UpdatePassword(id int, salt string, password string) error {
	_, err := st["updatePassword"].Exec(salt, password, time.Now().Unix(), id)
	return err
}

func Frozen(id int, frozen rune) error {
	_, err := st["frozen"].Exec(frozen, id)
	return err
}

func Level(id int, level int) error {
	_, err := st["level"].Exec(level, id)
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
	if err := r.Scan(&user.Id, &user.LicenseKey, &user.UserName, &user.Phone, &user.Email, &salt, &pass); err != nil {
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
		if err := r.Scan(&up.Id, &up.LicenseKey, &up.UserName, &up.Phone, &up.Email); err != nil {
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