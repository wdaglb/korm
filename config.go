package korm

type Config struct {
	Conn string // 连接名, 保持单个应用唯一即可
	Driver string
	Host string
	User string
	Pass string
	Port int
	Database string
	MaxOpenConns int
	MaxIdleConns int
}
