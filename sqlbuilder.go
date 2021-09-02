package korm

import (
	"fmt"
	"github.com/wdaglb/korm/schema"
	"github.com/wdaglb/korm/utils"
	"strings"
)

type SqlBuilder struct {
	model *Model
	sep string
	p string
	schema *schema.Schema
	data map[string]interface{}
	fields []string
	orders []string
	where *Where
	group []string
	having *Where
	offset *int
	limit *int
	bindParams []interface{}
}

func NewSqlBuilder(model *Model, schema *schema.Schema) *SqlBuilder {
	sq := &SqlBuilder{}
	sq.model = model
	sq.schema = schema
	//if model.Config().Driver == "mysql" {
	//	sq.sep = "@"
	//} else {
	//	sq.sep = ":"
	//}
	sq.sep = ":"
	return sq
}

func (t *SqlBuilder) parseField(field string) string {
	return utils.ParseField(t.model.config.Driver, t.schema.Type, field)
}

func (t *SqlBuilder) bindParam(value interface{}) {
	t.bindParams = append(t.bindParams, value)
}

func (t *SqlBuilder) AddField(str string) *SqlBuilder {
	fields := strings.Split(str, ",")
	for _, f := range fields {
		t.fields = append(t.fields, t.parseField(f))
	}
	return t
}

func (t *SqlBuilder) AddFieldRaw(str string) *SqlBuilder {
	fields := strings.Split(str, ",")
	for _, f := range fields {
		t.fields = append(t.fields, f)
	}
	return t
}

func (t *SqlBuilder) AddWhere(logic string, field string, op interface{}, condition ...interface{}) *SqlBuilder {
	var value interface{}
	if len(condition) == 0 {
		value = op
		op = "="
	} else {
		value = condition[0]
	}
	if t.where == nil {
		t.where = &Where{
			builder: t,
		}
	}
	t.where.AddCondition(WhereCondition{
		Logic: logic,
		Field: t.parseField(field),
		Operator: op.(string),
		Condition: value,
	})

	return t
}

func (t *SqlBuilder) AddOrder(field string, val string) *SqlBuilder {
	t.orders = append(t.orders, t.parseField(field) + " " + val)
	return t
}

func (t *SqlBuilder) AddGroup(name string) *SqlBuilder {
	t.group = append(t.group, t.parseField(name))
	return t
}

func (t *SqlBuilder) AddHaving(logic string, field string, op interface{}, condition ...interface{}) *SqlBuilder {
	var value interface{}
	if len(condition) == 0 {
		value = op
		op = "="
	} else {
		value = condition[0]
	}
	if t.having == nil {
		t.having = &Where{
			builder: t,
		}
	}
	t.having.AddCondition(WhereCondition{
		Logic:     logic,
		Field:     t.parseField(field),
		Operator:  op.(string),
		Condition: value,
	})

	return t
}

func (t *SqlBuilder) ToString() (string, []interface{}) {
	str := ""
	switch t.p {
	case "select":
		str = "SELECT [field] FROM [table]"
		if len(t.fields) == 0 {
			str = strings.ReplaceAll(str, "[field]", "*")
		} else {
			str = strings.ReplaceAll(str, "[field]", strings.Join(t.fields, ","))
		}
	case "insert":
		str = "INSERT INTO [table] ([columns]) VALUES ([values])"
		if t.model.config.Driver == "mssql" {
			str += ";select ID = convert(bigint, SCOPE_IDENTITY())"
		}
		keys := make([]string, 0)
		values := make([]string, 0)
		for k, v := range t.data {
			if k == t.schema.PrimaryKey {
				continue
			}
			t.bindParam(v)
			keys = append(keys, t.parseField(k))
			values = append(values, "?")
		}
		str = strings.ReplaceAll(str, "[columns]", strings.Join(keys, ","))
		str = strings.ReplaceAll(str, "[values]", strings.Join(values, ","))
	case "update":
		str = "UPDATE [table] SET [values]"
		values := make([]string, 0)
		for k, v := range t.data {
			if k == t.schema.PrimaryKey {
				continue
			}
			t.bindParam(v)
			values = append(values, fmt.Sprintf("%s=?", t.parseField(k)))
		}
		str = strings.ReplaceAll(str, "[values]", strings.Join(values, ","))
	case "delete":
		str = "DELETE FROM [table]"
	}
	str = strings.ReplaceAll(str, "[table]", t.schema.TableName)

	switch t.p {
	case "select", "update", "delete":
		if t.where != nil {
			str += " WHERE " + t.where.ToString()
		}
	}
	if len(t.group) > 0 {
		str += " GROUP BY " + strings.Join(t.group, ",")
	}
	if len(t.orders) > 0 {
		str += " ORDER BY " + strings.Join(t.orders, ", ")
	}
	if t.having != nil {
		str += fmt.Sprintf(" HAVING %s", t.having.ToString())
	}

	switch t.model.config.Driver {
	case "mssql":
		if t.offset != nil {
			str += fmt.Sprintf(" OFFSET %d", *t.offset)
			switch t.model.config.Driver {
			case "mssql":
				str += " ROWS FETCH NEXT [limit] ROWS ONLY"
			}
		}
	case "mysql":
		if t.p == "select" {
			if t.limit != nil {
				limit := fmt.Sprintf("%d", *t.limit)
				str += " LIMIT " + limit
			}

			if t.offset != nil {
				str += fmt.Sprintf(" OFFSET %d", *t.offset)
			}
		}
	}

	params := make([]interface{}, len(t.bindParams))
	for v := range t.bindParams {
		params[v] = t.bindParams[v]
	}

	fmt.Printf("sql: %v\n", str)
	return str, params
}
