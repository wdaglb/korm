package korm

import (
	"fmt"
	"reflect"
	"strings"
)

type Where struct {
	builder *SqlBuilder
	list []WhereCondition
}

type WhereCondition struct {
	Logic string
	Field string
	Operator string
	Condition interface{}
}

// 添加条件
func (t *Where) AddCondition(cond WhereCondition) {
	t.list = append(t.list, cond)
}

// 解析运算符
func (t *Where) parseOperator(v WhereCondition) string {
	if v.Operator == "in" || v.Operator == "not in" {
		typeOf := reflect.TypeOf(v.Condition)
		values := make([]string, 0)
		if typeOf.Kind() == reflect.Slice {
			valueOf := reflect.ValueOf(v.Condition)
			for i := 0; i < valueOf.Len(); i++ {
				val := valueOf.Index(i).Interface()
				t.builder.bindParam(val)
				values = append(values, "?")
			}
		}
		return fmt.Sprintf("%s %s(%s)", v.Field, v.Operator, strings.Join(values, ","))
	}

	t.builder.bindParam(v.Condition)
	return fmt.Sprintf("%s%s?", v.Field, v.Operator)
}

func (t *Where) ToString() string {
	str := ""
	for i, v := range t.list {
		if i == 0 {
			str += t.parseOperator(v)
		} else {
			str += " " + v.Logic + " " + t.parseOperator(v)
		}
	}
	return str
}
