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
	Name string
	ColumnName string
	DataType string
	StructField reflect.StructField
	Tag reflect.StructTag
	TagSetting map[string]string
	FieldType reflect.Type
	IndirectFieldType reflect.Type
}
