package schema

import (
	"fmt"
	"github.com/wdaglb/korm/mixins"
	"github.com/wdaglb/korm/relation"
	"github.com/wdaglb/korm/utils"
	"go/ast"
	"reflect"
	"strconv"
	"time"
)

type Schema struct {
	Type reflect.Type
	PrimaryKey string
	TableName string
	Data reflect.Value
	Fields []*Field
	FieldNames map[string]*Field
	Relations map[string]*relation.Relation
}

func NewSchema(data interface{}) *Schema {
	schema := &Schema{}
	schema.Type = reflect.TypeOf(data)
	for schema.Type.Kind() == reflect.Slice || schema.Type.Kind() == reflect.Array || schema.Type.Kind() == reflect.Ptr {
		schema.Type = schema.Type.Elem()
	}
	schema.TableName = utils.Camel2Case(schema.Type.Name())
	schema.Data = utils.Indirect(reflect.ValueOf(data))

	if ext, ok := schema.Data.Interface().(mixins.ModelTable); ok {
		schema.TableName = ext.Table()
	}
	schema.PrimaryKey = "Id"
	if ext, ok := schema.Data.Interface().(mixins.ModelPk); ok {
		schema.PrimaryKey = ext.Pk()
	}
	schema.FieldNames = make(map[string]*Field)
	for i := 0; i < schema.Type.NumField(); i++ {
		if fieldStruct := schema.Type.Field(i); ast.IsExported(fieldStruct.Name) {
			field := schema.AddField(fieldStruct)
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames[field.Name] = field
		}
	}
	return schema
}

func (schema *Schema) AddField(structField reflect.StructField) *Field {
	field := &Field{
		Name:              structField.Name,
		Tag:               structField.Tag,
		StructField:       structField,
		FieldType:         structField.Type,
		IndirectFieldType: utils.IndirectType(structField.Type),
	}

	tagDb := structField.Tag.Get("db")
	if tagDb != "" {
		field.ColumnName = tagDb
	} else {
		field.ColumnName = field.Name
	}
	fieldValue := reflect.New(field.IndirectFieldType)

	switch reflect.Indirect(fieldValue).Kind() {
	case reflect.Bool:
		field.DataType = Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.DataType = Int
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.DataType = Uint
	case reflect.Float32, reflect.Float64:
		field.DataType = Float
	case reflect.String:
		field.DataType = String
	case reflect.Struct:
		if _, ok := fieldValue.Interface().(*time.Time); ok {
			field.DataType = Time
		} else if fieldValue.Type().ConvertibleTo(reflect.TypeOf(time.Time{})) {
			field.DataType = Time
		} else if fieldValue.Type().ConvertibleTo(reflect.TypeOf(&time.Time{})) {
			field.DataType = Time
		}
	case reflect.Array, reflect.Slice:
		if reflect.Indirect(fieldValue).Type().Elem() == reflect.TypeOf(uint8(0)) {
			field.DataType = Bytes
		}
	}

	return field
}

func (schema *Schema) GetFieldName(name string) *Field {
	return schema.FieldNames[name]
}

// 获取字段值
func (schema *Schema) GetFieldValue(name string) interface{} {
	field := schema.FieldNames[name]
	if field == nil {
		return nil
	}
	fieldData := schema.Data.FieldByName(name)
	return fieldData.Interface()
}

// 设置字段值
func (schema *Schema) SetFieldValue(name string, value interface{}) error {
	field := schema.FieldNames[name]
	if field == nil {
		return nil
	}
	fieldData := schema.Data.FieldByName(name)
	return schema.SetStructValue(value, fieldData)
}

// 设置结构值
func (schema *Schema) SetStructValue(src interface{}, dst reflect.Value) error {
	if src == nil {
		return nil
	}
	sv := utils.Indirect(reflect.ValueOf(src))
	if dst.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dst.Type()) {
		dst.Set(sv.Convert(dst.Type()))
		return nil
	}

	switch dst.Kind() {
	case reflect.Ptr:
		if dst.IsValid() {
			dst.Set(reflect.New(dst.Type().Elem()))

			dvt := dst.Interface()

			if scanner, ok := dvt.(mixins.Scanner); ok {
				return scanner.Scan(src)
			}
			return nil
		}
		return schema.SetStructValue(src, dst)
	case reflect.Struct:
		if dst.IsValid() {
			var dvt interface{}
			dst.Set(reflect.New(dst.Type()))

			dvt = dst.Interface()

			if scanner, ok := dvt.(mixins.Scanner); ok {
				return scanner.Scan(src)
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dst.Kind())
		}
		s := utils.AsString(src)
		val, _ := strconv.ParseInt(s, 10, dst.Type().Bits())
		dst.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dst.Kind())
		}
		s := utils.AsString(src)
		val, _ := strconv.ParseUint(s, 10, dst.Type().Bits())
		dst.SetUint(val)
	case reflect.Float32, reflect.Float64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dst.Kind())
		}
		s := utils.AsString(src)
		val, _ := strconv.ParseFloat(s, dst.Type().Bits())
		dst.SetFloat(val)
	case reflect.String:
		dst.SetString(src.(string))
	default:
		fmt.Printf("type: %v\n", dst.Kind())
	}
	return nil
}

// 为数组添加元素
func (schema *Schema) AddArrayItem(data map[string]interface{}) {
	newValue := reflect.New(schema.Type)
	newValue = utils.Indirect(newValue)

	for i := 0; i < len(schema.Fields); i++ {
		field := schema.Fields[i]
		fieldValue := newValue.Field(i)
		if err := schema.SetStructValue(data[field.ColumnName], fieldValue); err != nil {
			fmt.Printf("error: %v\n", err)
		}

		// asValue(ret[colName], p, fieldValue)
	}
	tmp := reflect.Append(schema.Data, newValue)
	schema.Data.Set(tmp)
}
