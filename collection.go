package korm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Collection struct {
	model *Model
	data map[string]interface{}
}

func (c *Collection) SetData(data map[string]interface{}) *Collection {
	c.data = data
	return c
}

func (c *Collection) colName(p reflect.StructField) string {
	tag := p.Tag
	str := tag.Get("db")
	if str != "" {
		return str
	}
	return p.Name
}

func (c *Collection) indirectValue(target reflect.Value, p reflect.Type) reflect.Value {
	for target.Kind() == reflect.Ptr {
		target.Set(reflect.New(p))
		target = target.Elem()
	}
	return target
}

func (c *Collection) indirectType(target reflect.Type) reflect.Type {
	for target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	return target
}

// 转为value
func asValue(data interface{}, p reflect.Type, value reflect.Value) {
	switch p.Name() {
	case "Time":
		t, _ := time.Parse(time.RFC3339, data.(string))
		newVal := reflect.ValueOf(t)
		value.Set(newVal)
		return
	}

	switch p.Kind() {
	case reflect.Int64, reflect.Int:
		val, _ := strconv.ParseInt(data.(string), 10, 64)
		value.SetInt(val)
	case reflect.String:
		value.SetString(data.(string))
	default:
		fmt.Printf("type: %v\n", p)
	}
}

func (c *Collection) ToMap(dst *map[string]interface{}) {
	dst = &c.data
}

func (c *Collection) ToStruct() {
	mapToStruct(c.data, c.model.model)
	//fieldNum := c.model.reflectType.NumField()
	//if fieldNum == 0 {
	//	return
	//}
	//
	//for i := 0; i < fieldNum; i++ {
	//	field := c.model.reflectType.Field(i)
	//	colName := c.colName(field)
	//
	//	if c.data[colName] == nil {
	//		continue
	//	}
	//	p := c.indirectType(field.Type)
	//	value := c.indirectValue(c.model.reflectValue.Field(i), p)
	//
	//	asValue(c.data[colName], p, value)
	//}
	// fmt.Printf("struct: %v\n", c.context.model)
}
