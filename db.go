package korm

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type kdb struct {
	db *sql.DB
	config *Config
	dbConf *DbConfig
	currentKey *Queue
	txCount int
	tx map[int]*sql.Tx
}

func NewDb(config *Config, dbConf *DbConfig) (*kdb, error) {
	var (
		db *sql.DB
	)
	dsn := configToDsn(dbConf)
	if dsn == "" {
		return nil, fmt.Errorf("unsupported driver: %s", dbConf.Driver)
	}

	fmt.Printf("conn: %s\n", dsn)
	db, err := sql.Open(dbConf.Driver, dsn)
	if err != nil {
		return nil, err
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 3600 * 2
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	kdb := &kdb{}
	kdb.db = db
	kdb.config = config
	kdb.dbConf = dbConf
	kdb.currentKey = newQueue()
	kdb.tx = make(map[int]*sql.Tx)
	return kdb, nil
}

// 申请空闲事务标识
func (t *kdb) getKey() int {
	for k, v := range t.tx {
		if v == nil {
			return k
		}
	}
	t.txCount++
	return t.txCount
}

// Begin 开启一个事务, 返回事务标识
func (t *kdb) Begin() (int, error) {
	var (
		err error
		tx *sql.Tx
	)
	tx, err = t.db.Begin()
	if err != nil {
		return 0, err
	}
	k := t.getKey()
	t.currentKey.push(k)
	t.tx[k] = tx
	return k, nil
}

// Commit 提交标识对应的事务
func (t *kdb) Commit(key int) error {
	fmt.Printf("commit: %v\n", key)
	if t.tx[key] == nil {
		return fmt.Errorf("txKey not exist: %v", key)
	}
	defer func() {
		if t.tx[key] != nil {
			t.currentKey.pop()
		}
		t.tx[key] = nil
	}()
	return t.tx[key].Commit()
}

// Rollback 回滚标识对应的事务
func (t *kdb) Rollback(key int) error {
	if t.tx[key] == nil {
		return fmt.Errorf("txKey not exist: %v", key)
	}
	defer func() {
		if t.tx[key] != nil {
			t.currentKey.pop()
		}
		t.tx[key] = nil
	}()
	return t.tx[key].Rollback()
}

// 当前操作事务
func (t *kdb) current() *sql.Tx {
	key := t.currentKey.get()
	if key == nil {
		return nil
	}
	fmt.Printf("当前执行事务标识：%v\n", key)
	return t.tx[key.(int)]
}

// Handler 获取db实例
func (t *kdb) Handler() *sql.DB {
	return t.db
}

func (t *kdb) Exec(query string, args ...interface{}) (sql.Result, error) {
	if c := t.current(); c != nil {
		return c.Exec(query, args...)
	}
	return t.db.Exec(query, args...)
}

func (t *kdb) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if c := t.current(); c != nil {
		return c.ExecContext(ctx, query, args)
	}
	return t.db.ExecContext(ctx, query, args)
}

func (t *kdb) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if c := t.current(); c != nil {
		return c.QueryContext(ctx, query, args...)
	}
	return t.db.QueryContext(ctx, query, args...)
}

func (t *kdb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if c := t.current(); c != nil {
		return c.Query(query, args...)
	}
	return t.db.Query(query, args...)
}

func (t *kdb) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if c := t.current(); c != nil {
		return c.QueryRowContext(ctx, query, args...)
	}
	return t.db.QueryRowContext(ctx, query, args...)
}

func (t *kdb) QueryRow(query string, args ...interface{}) *sql.Row {
	if c := t.current(); c != nil {
		return c.QueryRow(query, args...)
	}
	return t.db.QueryRow(query, args...)
}

func (t *kdb) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if c := t.current(); c != nil {
		return c.PrepareContext(ctx, query)
	}
	return t.db.PrepareContext(ctx, query)
}

func (t *kdb) Prepare(query string) (*sql.Stmt, error) {
	if c := t.current(); c != nil {
		return c.Prepare(query)
	}
	return t.db.Prepare(query)
}
