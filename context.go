package korm

import (
	"fmt"
	"reflect"
)

type Context struct {
	conn string
}

type ModelTable interface {
	Table() string
}

type ModelPk interface {
	Pk() string
}

type TransactionCall func() error

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
	model.config = ctx.Config()
	model.db = ctx.Db()
	model.model = mod

	model.reflectType = reflect.TypeOf(mod)
	for model.reflectType.Kind() == reflect.Slice || model.reflectType.Kind() == reflect.Array || model.reflectType.Kind() == reflect.Ptr {
		model.reflectType = model.reflectType.Elem()
	}
	modelValue := reflect.New(model.reflectType)
	model.pk = "Id"
	if ext, ok := modelValue.Interface().(ModelTable); ok {
		model.table = ext.Table()
	} else {
		model.table = Camel2Case(model.reflectType.Name())
	}
	if ext, ok := modelValue.Interface().(ModelPk); ok {
		model.pk = ext.Pk()
	}

	model.builder = NewSqlBuilder(model)
	model.builder.table = model.table
	model.reflectValue = indirect(reflect.ValueOf(mod))

	return model
}
