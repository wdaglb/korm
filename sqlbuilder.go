package korm

import (
	"fmt"
	"strings"
)

type SqlBuilder struct {
	model *Model
	sep string
	p string
	table string
	data map[string]interface{}
	fields []string
	orders []string
	where []string
	group []string
	having []string
	offset *int
	limit *int
	bindParams []interface{}
}

func NewSqlBuilder(model *Model) *SqlBuilder {
	sq := &SqlBuilder{}
	sq.model = model
	//if model.Config().Driver == "mysql" {
	//	sq.sep = "@"
	//} else {
	//	sq.sep = ":"
	//}
	sq.sep = ":"
	sq.where = make([]string, 0)
	return sq
}

func (t *SqlBuilder) parseField(field string) string {
	return parseField(t.model.config.Driver, t.model.reflectType, field)
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
	t.bindParam(value)
	if len(t.where) == 0 {
		logic = ""
	} else {
		logic = fmt.Sprintf(" %s ", strings.ToUpper(logic))
	}
	t.where = append(t.where, fmt.Sprintf("%s%s%s?", logic, t.parseField(field), op))

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
	t.bindParam(value)
	if len(t.where) == 0 {
		logic = ""
	} else {
		logic = fmt.Sprintf(" %s ", strings.ToUpper(logic))
	}
	t.having = append(t.having, fmt.Sprintf("%s%s%s?", logic, t.parseField(field), op))

	return t
}

func (t *SqlBuilder) toWhere() string {
	return strings.Join(t.where, "")
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
			if k == t.model.pk {
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
			if k == t.model.pk {
				continue
			}
			t.bindParam(v)
			values = append(values, fmt.Sprintf("%s=?", t.parseField(k)))
		}
		str = strings.ReplaceAll(str, "[values]", strings.Join(values, ","))
	case "delete":
		str = "DELETE FROM [table]"
	}
	str = strings.ReplaceAll(str, "[table]", t.table)

	switch t.p {
	case "select", "update", "delete":
		if len(t.where) > 0 {
			str += " WHERE " + t.toWhere()
		}
	}
	if len(t.group) > 0 {
		str += " GROUP BY " + strings.Join(t.group, ",")
	}
	if len(t.orders) > 0 {
		str += " ORDER BY " + strings.Join(t.orders, ", ")
	}
	if len(t.having) > 0 {
		str += fmt.Sprintf(" HAVING %s", strings.Join(t.having, ""))
	}
	if t.offset != nil {
		str += fmt.Sprintf(" OFFSET %d", *t.offset)
		switch t.model.config.Driver {
		case "mssql":
			str += " ROWS FETCH NEXT [limit] ROWS ONLY"
		}
	}
	limit := "1"
	if t.limit != nil {
		limit = fmt.Sprintf("%d", *t.limit)
	}
	str = strings.ReplaceAll(str, "[limit]", limit)

	params := make([]interface{}, len(t.bindParams))
	for v := range t.bindParams {
		params[v] = t.bindParams[v]
	}

	return str, params
}
