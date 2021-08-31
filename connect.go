package korm

import (
	"database/sql"
	"fmt"
)

var (
	defaultConn = "default"
	dbMaps map[string]*connect
)

type connect struct {
	db *kdb
	config Config
}

func init()  {
	dbMaps = make(map[string]*connect, 0)
}

// 设置默认使用的连接名
func SetDefaultConn(conn string)  {
	defaultConn = conn
}

// 讲struct转为dsn字符
func configToDsn(config Config) string {
	switch config.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Database)
	case "mssql":
		return fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", config.Host, config.Database, config.User, config.Pass, config.Port)
	}
	return ""
}

// 连接数据库
func Connect(config Config) (err error) {
	var (
		db *sql.DB
	)
	dsn := configToDsn(config)
	if dsn == "" {
		return fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	fmt.Printf("conn: %s\n", dsn)
	db, err = sql.Open(config.Driver, dsn)
	if err != nil {
		return
	}

	kdb := NewDb(db)
	conn := &connect{
		db: kdb,
		config: config,
	}
	if config.Conn == "" {
		dbMaps[defaultConn] = conn
	} else {
		dbMaps[config.Conn] = conn
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}

	return
}

// 关闭所有连接
func Close() {
	for _, db := range dbMaps {
		_ = db.db.db.Close()
	}
}
