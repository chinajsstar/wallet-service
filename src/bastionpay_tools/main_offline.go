package main

import (
	"fmt"
	"strings"
	"api_router/base/utils"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"bastionpay_tools/tools"
)

func testSqlite3()  {
	fmt.Println("test...")
	//os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
	}
	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("こんにちわ世界%03d", i))
		if err != nil {
			fmt.Println(err)
		}
	}
	tx.Commit()

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func main()  {
	curDir, _ := utils.GetCurrentDir()
	dataDir := curDir + "/data"
	err := os.Mkdir(dataDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("创建数据目录失败：%s", err.Error())
		return
	}

	//testSqlite3()
	//return

	ol := &tools.OffLine{}
	err = ol.Start(dataDir)
	if err != nil {
		fmt.Printf("启动离线工具失败：%s", err.Error())
		return
	}

	for {
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			fmt.Println("I do quit")
			break;
		}else{
			ol.Execute(argv)
		}
	}
}