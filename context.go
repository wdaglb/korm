package korm

import (
	"fmt"
	"github.com/wdaglb/korm/schema"
)

type Context struct {
	conn string
	queryAfter []QueryAfterCallback
}

type TransactionCall func() error

func NewContext() *Context {
	ctx := &Context{}
	RegisterCallback(ctx)
	return ctx
}

// 使用连接名
func (ctx *Context) Use(conn string) *Context {
	ctx.conn = conn
	return ctx
}

// 取得当前连接sql.DB实例
func (ctx *Context) Db() *kdb {
	if ctx.conn == "" {
		return dbMaps[defaultConn].db
	}
	return dbMaps[ctx.conn].db
}

// 取得当前模型使用的配置
func (ctx *Context) Config() Config {
	if ctx.conn == "" {
		return dbMaps[defaultConn].config
	}
	return dbMaps[ctx.conn].config
}

// 添加查询后事件
func (ctx *Context) AddQueryAfterCallback(callback QueryAfterCallback) *Context {
	ctx.queryAfter = append(ctx.queryAfter, callback)
	return ctx
}

func (ctx *Context) callQueryAfterCallbacks(params *CallbackParams) (err error) {
	for _, fun := range ctx.queryAfter {
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
	model.config = ctx.Config()
	model.db = ctx.Db()
	model.model = mod
	model.schema = schema.NewSchema(mod)
	if len(model.schema.WithList) > 0 {
		model.withList = append(model.withList, model.schema.WithList...)
	}

	model.builder = NewSqlBuilder(model, model.schema)

	return model
}
