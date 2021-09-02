package schema

import (
	"reflect"
)

const (
	Bool = "bool"
	Int = "int"
	Uint = "uint"
	Float = "float"
	String = "string"
	Time = "time"
	Bytes = "bytes"
)

type Field struct {
	Schema *Schema
	Name string
	ColumnName string
	DataType string
	StructField reflect.StructField
	Tag reflect.StructTag
	TagSetting map[string]string
	FieldType reflect.Type
	IndirectFieldType reflect.Type
}

func (field *Field) GetColumnName() string {
	col := field.Tag.Get("db")
	if col == "" {
		return field.Name
	}
	return col
}

func (field *Field) GetPrimaryName() string {
	val := field.Tag.Get("pk")
	if val == "" {
		return field.Schema.PrimaryKey
	}
	return val
}

func (field *Field) GetForeignName() string {
	val := field.Tag.Get("fk")
	if val == "" {
		return field.Name + field.Schema.PrimaryKey
	}
	return val
}

func (field *Field) GetPrimaryKey() string {
	val := field.Tag.Get("pk")
	if val == "" {
		return field.Schema.FieldNameToColumnName(field.Schema.PrimaryKey)
	}
	return field.Schema.FieldNameToColumnName(val)
}

func (field *Field) GetForeignKey() string {
	val := field.Tag.Get("fk")
	if val == "" {
		return field.Schema.FieldNameToColumnName(field.Name + field.Schema.PrimaryKey)
	}
	return field.Schema.FieldNameToColumnName(val)
}
