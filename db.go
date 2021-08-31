package korm

import (
	"context"
	"database/sql"
	"fmt"
)

type kdb struct {
	db *sql.DB
	currentKey *Queue
	txCount int
	tx map[int]*sql.Tx
}

func NewDb(db *sql.DB) *kdb {
	d := &kdb{}
	d.db = db
	d.tx = make(map[int]*sql.Tx)
	d.currentKey = newQueue()
	return d
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

// 开启一个事务, 返回事务标识
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

// 提交标识对应的事务
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

// 回滚标识对应的事务
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
	key := t.currentKey.get().(int)
	if key == 0 {
		return nil
	}
	fmt.Printf("当前执行事务标识：%v\n", key)
	return t.tx[key]
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
