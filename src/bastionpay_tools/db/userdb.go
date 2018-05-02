package db

import (
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

var(
	UserLevel_Operator = 1	// 操作员
	UserLevel_Admin = 2		// 管理员
)

type ToolUser struct{
	Name string
	Level int
}

const tableUser = `CREATE TABLE user (
 id INTEGER,
 name varchar(255) NOT NULL,
 salt varchar(16) NOT NULL,
 password varchar(255) NOT NULL,
 level INTEGER NOT NULL DEFAULT 1,
 primary key(id, name)
);`

func (dbi *DBInstance)ExistUser(name string) (bool, error) {
	stmt, err := dbi.Db.Prepare("select id from user where name = ?")
	if err != nil {
		return true, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(name)
	if err != nil {
		return true, err
	}
	defer rows.Close()

	if !rows.Next() {
		// not exist
		return false, nil
	}

	// exist
	return true, nil
}

func (dbi *DBInstance)AddUser(toolUser *ToolUser, salt, password string) (error) {
	exist, err := dbi.ExistUser(toolUser.Name)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("user name is exist")
	}

	stmt, err := dbi.Db.Prepare("insert into user(name, salt, password, level) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(toolUser.Name, salt, password, toolUser.Level)
	return err
}

func (dbi *DBInstance)DeleteUser(name string) (error) {
	exist, err := dbi.ExistUser(name)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("user name is not exist")
	}

	stmt, err := dbi.Db.Prepare("delete from user where name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	return err
}

func (dbi *DBInstance)GetUser(name string) (*ToolUser, string, string, error) {
	exist, err := dbi.ExistUser(name)
	if err != nil {
		return nil, "", "", err
	}
	if !exist {
		return nil, "", "", errors.New("user name is not exist")
	}

	stmt, err := dbi.Db.Prepare("select salt, password, level from user where name = ?")
	if err != nil {
		return nil, "", "", err
	}
	defer stmt.Close()

	rows, err := stmt.Query(name)
	if err != nil {
		return nil, "", "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, "", "", errors.New("row next is nil")
	}

	var salt, password string
	toolUser := &ToolUser{}
	err = rows.Scan(&salt, &password, &toolUser.Level)
	if err != nil {
		return nil, "", "", err
	}

	toolUser.Name = name
	return toolUser, salt, password, nil
}