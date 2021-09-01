package utils

import (
	"bytes"
	"fmt"
	"github.com/wdaglb/korm/mixins"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 首字母大写
func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// 首字母小写
func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i interface{}) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		_, err := b.Write(val)
		if err != nil {
			return nil
		}
	case rune:
		_, err := b.WriteRune(val)
		if err != nil {
			return nil
		}
	}
	return b
}

func (b *Buffer) append(s string) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()
	_, err := b.WriteString(s)
	if err != nil {
		return nil
	}
	return b
}

// 驼峰式写法转为下划线写法
func Camel2Case(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}
	return buffer.String()
}

// 下划线写法转为驼峰写法
func Case2Camel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	return strings.Replace(name, " ", "", -1)
}

// 解析字段名
func ParseField(driver string, reType reflect.Type, field string) string {
	var (
		p reflect.StructField
		ok bool
	)
	p, ok = reType.FieldByName(field)
	if ok {
		colName := p.Tag.Get("db")
		if colName != "" {
			field = colName
		}
	}

	switch driver {
	case "mssql":
		return fmt.Sprintf("[%s]", field)
	case "mysql":
		return fmt.Sprintf("`%s`", field)
	default:
		return field
	}
}

func AsString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func Indirect(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		return Indirect(value)
	}
	return value
}

func IndirectType(value reflect.Type) reflect.Type {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		return IndirectType(value)
	}
	return value
}


// 转为value
func AsValue(data interface{}, p reflect.Type, value reflect.Value) {
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

// 调用数据修改器
func CallValue(valueOf reflect.Value) interface{} {
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	switch valueOf.Kind() {
	case reflect.Struct:
		typeOf := valueOf.Type()
		method, ok := typeOf.MethodByName("Value")

		if ok {
			val := method.Func.Call([]reflect.Value{valueOf})
			// val := method.Func.Call(nil)
			if len(val) != 2 {
				return nil
			}

			newVal := val[0].Interface()
			return newVal
		}
	}
	return nil
}

// 调用数据获取器
func CallScan(src interface{}, dv reflect.Value) interface{} {
	sv := reflect.ValueOf(src)

	if src == nil {
		return nil
	}
	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	switch dv.Kind() {
	case reflect.Ptr:
		if dv.IsValid() {
			dv.Set(reflect.New(dv.Type().Elem()))

			dvt := dv.Interface()

			if scanner, ok := dvt.(mixins.Scanner); ok {
				_ = scanner.Scan(src)
				return nil
			}
			return nil
		}
		return CallScan(src, dv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := AsString(src)
		val, _ := strconv.ParseInt(s, 10, dv.Type().Bits())
		dv.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := AsString(src)
		val, _ := strconv.ParseUint(s, 10, dv.Type().Bits())
		dv.SetUint(val)
	case reflect.Float32, reflect.Float64:
		if src == nil {
			return fmt.Errorf("converting NULL to %s is unsupported", dv.Kind())
		}
		s := AsString(src)
		val, _ := strconv.ParseFloat(s, dv.Type().Bits())
		dv.SetFloat(val)
	case reflect.String:
		dv.SetString(src.(string))
	default:
		fmt.Printf("type: %v\n", dv.Kind())
	}
	return nil
}

// 结构体转为map
func StructToMap(data interface{}) map[string]interface{} {
	typeOf := IndirectType(reflect.TypeOf(data))
	valueOf := Indirect(reflect.ValueOf(data))
	fieldNum := valueOf.NumField()
	dst := make(map[string]interface{})
	for i := 0; i < fieldNum; i++ {
		typeof := typeOf.Field(i)
		field := valueOf.Field(i)
		if field.Kind() == reflect.Ptr {
			dst[typeof.Name] = CallValue(field)
		} else {
			dst[typeof.Name] = field.Interface()
		}
	}
	return dst
}

// map转为结构体
func MapToStruct(data map[string]interface{}, dst interface{})  {
	typeOf := IndirectType(reflect.TypeOf(dst))
	valueOf := Indirect(reflect.ValueOf(dst))

	// fmt.Printf("vvv: %v\n", dst)
	fieldNum := valueOf.NumField()
	for i := 0; i < fieldNum; i++ {
		typeofItem := typeOf.Field(i)
		valueOfItem := valueOf.Field(i)
		colName := typeofItem.Name
		if typeofItem.Tag.Get("db") != "" {
			colName = typeofItem.Tag.Get("db")
		}
		// val := data[colName]
		CallScan(data[colName], valueOfItem)
		//if valueOfItem.Kind() == reflect.Ptr {
		//	fmt.Printf("colname: %v\n", colName)
		//} else {
		//	data[colName] = valueOfItem.Interface()
		//}
	}
}
