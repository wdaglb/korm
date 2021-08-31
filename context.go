package korm

import (
	"reflect"
)

type Context struct {
}

type ModelTable interface {
	Table() string
}

type ModelPk interface {
	Pk() string
}

func (m *Context) Model(mod interface{}) *Model {
	model := &Model{}
	model.model = mod

	model.reflectType = reflect.TypeOf(mod)
	for model.reflectType.Kind() == reflect.Slice || model.reflectType.Kind() == reflect.Array || model.reflectType.Kind() == reflect.Ptr {
		model.reflectType = model.reflectType.Elem()
	}
	modelValue := reflect.New(model.reflectType)
	model.pk = "Id"
	if ext, ok := modelValue.Interface().(ModelTable); ok {
		model.table = ext.Table()
	}
	if ext, ok := modelValue.Interface().(ModelPk); ok {
		model.pk = ext.Pk()
	}

	model.builder = NewSqlBuilder(model)
	model.builder.table = model.table
	model.reflectValue = indirect(reflect.ValueOf(mod))

	return model
}
