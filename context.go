package korm

import (
	"database/sql"
	"fmt"
	"github.com/wdaglb/korm/schema"
)

type Context struct {
	conn string
	events map[string][]EventCallback
}

type TransactionCall func() error

func NewContext() *Context {
	ctx := &Context{}
	ctx.events = make(map[string][]EventCallback)
	RegisterCallback(ctx)
	return ctx
}

func UseContext(conn string) *Context {
	newCtx := NewContext()
	newCtx.conn = conn
	return newCtx
}

// 使用连接名
func (ctx *Context) Use(conn string) *Context {
	newCtx := NewContext()
	newCtx.conn = conn
	return newCtx
}

// 取得当前连接sql.DB实例
func (ctx *Context) Db() *kdb {
	if ctx.conn == "" {
		return mainConnect.dbList[mainConnect.config.DefaultConn]
	}
	return mainConnect.dbList[ctx.conn]
}

// 监听查询后事件
func (ctx *Context) OnEventQueryAfter(callback EventCallback) *Context {
	event := "query_after"
	ctx.events[event] = append(ctx.events[event], callback)
	return ctx
}

// 监听插入后事件
func (ctx *Context) OnInsertAfterCallback(callback EventCallback) *Context {
	event := "insert_after"
	ctx.events[event] = append(ctx.events[event], callback)
	return ctx
}

// 监听更新后事件
func (ctx *Context) OnUpdateAfterCallback(callback EventCallback) *Context {
	event := "update_after"
	ctx.events[event] = append(ctx.events[event], callback)
	return ctx
}

// 监听删除后事件
func (ctx *Context) OnDeleteAfterCallback(callback EventCallback) *Context {
	event := "delete_after"
	ctx.events[event] = append(ctx.events[event], callback)
	return ctx
}

// 调用指定事件
func (ctx *Context) emitEvent(event string, params *CallbackParams) (err error) {
	for _, fun := range ctx.events[event] {
		err = fun(params)
		if err != nil {
			return
		}
	}
	return
}

// 事务处理
func (ctx *Context) Transaction(call TransactionCall) error {
	key, err := ctx.Db().Begin()
	if err != nil {
		return fmt.Errorf("transaction enable fail: %v", err)
	}

	defer func() {
		fmt.Printf("释放tx: %v\n", key)
		if err := recover(); err != nil {
			_ = ctx.Db().Rollback(key)
		}
	}()

	err = call()
	if err != nil {
		_ = ctx.Db().Rollback(key)
		return err
	}

	err = ctx.Db().Commit(key)
	if err != nil {
		return err
	}
	return nil
}

// 新的模型实例
func (ctx *Context) Model(mod interface{}) *Model {
	model := &Model{}
	model.context = ctx
	model.db = ctx.Db()
	model.model = mod
	model.schema = schema.NewSchema(mod)
	if len(model.schema.WithList) > 0 {
		model.withList = append(model.withList, model.schema.WithList...)
	}

	model.builder = NewSqlBuilder(model, model.schema)

	return model
}

// query
func (ctx *Context) Query(sqlStr string, params ...interface{}) (*sql.Rows, error) {
	var (
		err error
		stmt *sql.Stmt
	)
	db := ctx.Db()
	stmt, err = db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, fmt.Errorf("query fail: %v", err)
	}
	return rows, nil
}

// exec
func (ctx *Context) Exec(sqlStr string, params ...interface{}) (sql.Result, error) {
	var (
		err error
		stmt *sql.Stmt
	)
	db := ctx.Db()
	stmt, err = db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Exec(params...)
	if err != nil {
		return nil, fmt.Errorf("exec fail: %v", err)
	}
	return rows, nil
}
