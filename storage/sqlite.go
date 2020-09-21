package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"time"
)

type SqliteDB struct {
	db *sql.DB
}

func NewSqlite(dbfile string) (Storage, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}

	createTable(db)

	return &SqliteDB {
		db : db,
	}, nil
}

func createTable(db *sql.DB) error {
	sql := `create table if not exists "indexs" (
		"id" integer primary key autoincrement,
		"key_name" text not null,
		"key_value" text not null,
		"created" datetime
	)`

	_, err := db.Exec(sql)
	return err
}

func (sdb *SqliteDB) AddData(k, v []byte) error {
	sql := `INSERT INTO indexs (key_name, key_value, created) values (?,?,?)`
	stmt, err := sdb.db.Prepare(sql)
	if err != nil {
		return err
	}

	timeStr := time.Now().Format("2006-01-02 15:04:05")
	res, err := stmt.Exec(k, v, timeStr)

	id, err := res.LastInsertId()

	fmt.Println(id)

	return nil
}

func (sdb *SqliteDB) GetData(k []byte) (v []byte, err error) {
	sql := `SELECT * FROM indexs WHERE key_name = ?`
	stmt, err := sdb.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(k)
	if err != nil {
		return nil, err
	}

	data := make([][]byte, 0)
	for rows.Next() {
		var key_name, key_value []byte
		var created string
		rows.Scan(&key_name, &key_value, &created)
		data = append(data, key_value)
	}

	return data[0], nil
}

func (sdb *SqliteDB) DelData(k []byte) error {
	sql := `DELETE FROM indexs WHERE id = ?`
	stmt, err := sdb.db.Prepare(sql)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(k)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (sdb *SqliteDB) Close() error {
	return sdb.db.Close()
}