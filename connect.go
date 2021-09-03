package korm

import (
	"fmt"
)

var (
	mainConnect *connect
)

type connect struct {
	config Config
	dbList map[string]*kdb
}

// 讲dbConfig转为dsn字符
func configToDsn(config DbConfig) string {
	switch config.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Database)
	case "mssql":
		return fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", config.Host, config.Database, config.User, config.Pass, config.Port)
	}
	return ""
}

// 连接数据库
func Connect(config Config) *connect {
	mainConnect = &connect{}
	mainConnect.config = config
	if mainConnect.config.DefaultConn == "" {
		mainConnect.config.DefaultConn = "default"
	}
	mainConnect.dbList = make(map[string]*kdb)
	return mainConnect
}

// 添加数据库连接
func (c *connect) AddDb(config DbConfig) error {
	if config.Conn == "" {
		config.Conn = config.Database
	}
	v, err := NewDb(c.config, config)
	if err != nil {
		return err
	}
	c.dbList[config.Conn] = v
	return nil
}

// 关闭所有连接
func Close() {
	//for _, db := range dbli {
	//	_ = db.db.db.Close()
	//}
}
