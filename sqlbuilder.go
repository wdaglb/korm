package korm

import (
	"fmt"
	"github.com/wdaglb/korm/schema"
	"github.com/wdaglb/korm/utils"
	"strings"
)

type SqlBuilder struct {
	model        *Model
	sep          string
	p            string
	schema       *schema.Schema
	data         map[string]interface{}
	fields       []SqlField
	ignoreFields []string
	rawFields    []string
	resultFields map[string]string
	clearField   bool
	orders       []string
	where        *Where
	group        []string
	having       *Where
	offset       *int
	limit        *int
	bindParams   []interface{}
}

type SqlField struct {
	Name  string
	IsRaw bool
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

func (t *SqlBuilder) parseField(field string, raw bool) string {
	return utils.ParseField(t.model.db.dbConf.Driver, t.schema.Type, field, raw)
}

func (t *SqlBuilder) bindParam(value interface{}) {
	t.bindParams = append(t.bindParams, value)
}

func (t *SqlBuilder) AddField(str string) *SqlBuilder {
	fields := strings.Split(str, ",")
	for _, f := range fields {
		t.fields = append(t.fields, SqlField{
			Name:  f,
			IsRaw: false,
		})
	}
	return t
}

func (t *SqlBuilder) AddFieldRaw(str string) *SqlBuilder {
	fields := strings.Split(str, ",")
	for _, f := range fields {
		t.rawFields = append(t.rawFields, f)
	}
	return t
}

// 忽略字段
func (t *SqlBuilder) IgnoreField(str string) *SqlBuilder {
	fields := strings.Split(str, ",")
	for _, f := range fields {
		t.ignoreFields = append(t.ignoreFields, f)
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
		Logic:     logic,
		Field:     t.parseField(field, false),
		Operator:  op.(string),
		Condition: value,
	})

	return t
}

func (t *SqlBuilder) AddOrder(field string, val string) *SqlBuilder {
	t.orders = append(t.orders, t.parseField(field, false)+" "+val)
	return t
}

func (t *SqlBuilder) AddOrderRaw(field string, val string) *SqlBuilder {
	t.orders = append(t.orders, field+" "+val)
	return t
}

func (t *SqlBuilder) AddGroup(name string) *SqlBuilder {
	t.group = append(t.group, t.parseField(name, false))
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
		Field:     t.parseField(field, false),
		Operator:  op.(string),
		Condition: value,
	})

	return t
}

func (t *SqlBuilder) GetTable() string {
	var table string
	switch t.model.db.dbConf.Driver {
	case "mssql":
		table = fmt.Sprintf("[%s]", t.model.db.dbConf.TablePrefix+t.schema.TableName)
	case "mysql":
		table = fmt.Sprintf("`%s`", t.model.db.dbConf.TablePrefix+t.schema.TableName)
	}
	return table
}

func (t *SqlBuilder) ToString() (string, []interface{}) {
	str := ""
	switch t.p {
	case "select":
		str = "SELECT [field] FROM [table]"
		if t.model.db.dbConf.Driver == "mssql" && t.limit != nil && t.offset == nil {
			str = strings.ReplaceAll(str, "SELECT ", fmt.Sprintf("SELECT TOP %d ", *t.limit))
		}
		var (
			fs  []SqlField
			fsv []string
		)
		if !t.clearField {
			if len(t.fields) > 0 {
				fs = t.fields
			} else {
				for _, f := range t.schema.Fields {
					if f.DataType == "" {
						continue
					}
					fs = append(fs, SqlField{
						Name:  f.Name,
						IsRaw: false,
					})
				}
			}
		}

		t.resultFields = make(map[string]string, 0)
		for _, v := range fs {
			if utils.InStrArray(t.ignoreFields, v.Name) {
				continue
			}
			fsv = append(fsv, t.parseField(v.Name, v.IsRaw))
			vf, _ := utils.ParseFieldDb(t.schema.Type, v.Name)
			t.resultFields[v.Name] = vf
		}
		for _, v := range t.rawFields {
			fsv = append(fsv, v)
			vf, _ := utils.ParseFieldDb(t.schema.Type, v)
			t.resultFields[v] = vf
		}

		str = strings.ReplaceAll(str, "[field]", strings.Join(fsv, ","))
	case "insert":
		str = "INSERT INTO [table] ([columns]) VALUES ([values])"
		if t.model.db.dbConf.Driver == "mssql" {
			str += ";select ID = convert(bigint, SCOPE_IDENTITY())"
		}
		keys := make([]string, 0)
		values := make([]string, 0)
		var (
			fs []SqlField
		)
		if len(t.fields) > 0 {
			fs = t.fields
		} else {
			for _, f := range t.schema.Fields {
				if f.DataType == "" {
					continue
				}
				fs = append(fs, SqlField{
					Name:  f.Name,
					IsRaw: false,
				})
			}
		}

		for _, k := range fs {
			if k.Name == t.schema.PrimaryKey {
				continue
			}
			f := t.schema.FieldNames[k.Name]
			if f.DataType == "" {
				continue
			}
			if utils.InStrArray(t.ignoreFields, k.Name) {
				continue
			}
			t.bindParam(t.data[k.Name])
			keys = append(keys, t.parseField(k.Name, k.IsRaw))
			values = append(values, "?")
		}
		str = strings.ReplaceAll(str, "[columns]", strings.Join(keys, ","))
		str = strings.ReplaceAll(str, "[values]", strings.Join(values, ","))
	case "update":
		str = "UPDATE [table] SET [values]"
		values := make([]string, 0)

		var (
			fs []SqlField
		)
		if len(t.fields) > 0 {
			fs = t.fields
		} else {
			for _, f := range t.schema.Fields {
				if f.DataType == "" {
					continue
				}
				fs = append(fs, SqlField{
					Name:  f.Name,
					IsRaw: false,
				})
			}
		}

		for _, k := range fs {
			if k.Name == t.schema.PrimaryKey {
				continue
			}
			f := t.schema.FieldNames[k.Name]
			if f.DataType == "" {
				continue
			}
			if utils.InStrArray(t.ignoreFields, k.Name) {
				continue
			}
			t.bindParam(t.data[k.Name])
			values = append(values, fmt.Sprintf("%s=?", t.parseField(k.Name, k.IsRaw)))
		}
		str = strings.ReplaceAll(str, "[values]", strings.Join(values, ","))
	case "delete":
		str = "DELETE FROM [table]"
	}
	str = strings.ReplaceAll(str, "[table]", t.GetTable())

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

	switch t.model.db.dbConf.Driver {
	case "mssql":
		if t.offset != nil {
			str += fmt.Sprintf(" OFFSET %d", *t.offset)
			str += fmt.Sprintf(" ROWS FETCH NEXT %d ROWS ONLY", *t.limit)
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

	if t.model.db.config.PrintSql {
		fmt.Printf("sql: %v\n", str)
	}

	params := make([]interface{}, len(t.bindParams))
	for v := range t.bindParams {
		params[v] = t.bindParams[v]
	}

	// fmt.Printf("sql: %v\n", str)
	return str, params
}
