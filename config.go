package korm

type Config struct {
	DefaultConn string
	MaxOpenConns int // 最大打开连接数
	MaxIdleConns int // 最大空闲连接数
	ConnMaxLifetime int // 保持连接时间
	PrintSql bool
}

type DbConfig struct {
	Conn string // 连接名, 保持单个应用唯一即可
	Driver string
	Host string
	User string
	Pass string
	Port int
	Database string
	TablePrefix string
}
