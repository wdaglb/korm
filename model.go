package korm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/wdaglb/korm/schema"
	"github.com/wdaglb/korm/utils"
	"reflect"
)

type Model struct {
	db *kdb
	model interface{}
	context *Context
	builder *SqlBuilder
	schema *schema.Schema
	collection *Collection
	withList []string

	relationData map[string][]*relation
	relationMap map[string]*relation
	cancelTogethers []string // 取消关联数据同步操作
}

// 转为map
func (m *Model) toMap(rows *sql.Rows) (map[string]interface{}, error) {
	row := make(map[string]interface{})
	fields, _ := rows.Columns()
	values := make([][]byte, len(fields))
	scans := make([]interface{}, len(fields))
	for i := range values {
		scans[i] = &values[i]
	}
	if err := rows.Scan(scans...); err != nil {
		return nil, err
	}
	for k, v := range values {
		key := fields[k]
		row[key] = string(v)
	}

	return row, nil
}

// 关联加载
func (m *Model) With(list ...string) *Model {
	temps := make([]string, 0)
	if len(m.withList) > 0 {
		for _, v := range m.withList {
			for _, v2 := range list {
				if v2 != v {
					temps = append(temps, v2)
				}
			}
		}
	} else {
		temps = list
	}
	m.withList = append(m.withList, temps...)
	return m
}

// 取消关联数据同步操作
func (m *Model) CancelTogether(list ...string) *Model {
	m.cancelTogethers = append(m.cancelTogethers, list...)
	return m
}

func (m *Model) Field(str string) *Model {
	m.builder.AddField(str)
	return m
}

func (m *Model) FieldRaw(str string) *Model {
	m.builder.AddFieldRaw(str)
	return m
}

func (m *Model) IgnoreField(str ...string) *Model {
	for _, v := range str {
		m.builder.IgnoreField(v)
	}
	return m
}

func (m *Model) Where(field string, op interface{}, condition ...interface{}) *Model {
	m.builder.AddWhere("and", field, op, condition...)
	return m
}

func (m *Model) WhereOr(field string, op interface{}, condition ...interface{}) *Model {
	m.builder.AddWhere("or", field, op, condition...)
	return m
}

func (m *Model) Group(name string) *Model {
	m.builder.AddGroup(name)
	return m
}

func (m *Model) OrderByDesc(field ...string) *Model {
	for _, f := range field {
		m.builder.AddOrder(f, "DESC")
	}
	return m
}

func (m *Model) OrderByAsc(field ...string) *Model {
	for _, f := range field {
		m.builder.AddOrder(f, "ASC")
	}
	return m
}

func (m *Model) OrderRawByDesc(field ...string) *Model {
	for _, f := range field {
		m.builder.AddOrderRaw(f, "DESC")
	}
	return m
}

func (m *Model) OrderRawByAsc(field ...string) *Model {
	for _, f := range field {
		m.builder.AddOrderRaw(f, "ASC")
	}
	return m
}

func (m *Model) Offset(val int) *Model {
	m.builder.offset = &val
	return m
}

func (m *Model) Limit(val int) *Model {
	m.builder.limit = &val
	return m
}

func (m *Model) Having(field string, op interface{}, condition ...interface{}) *Model {
	m.builder.AddWhere("and", field, op, condition...)
	return m
}

func (m *Model) HavingOr(field string, op interface{}, condition ...interface{}) *Model {
	m.builder.AddWhere("or", field, op, condition...)
	return m
}

// 获取一行数据
func (m *Model) Find() *Collection {
	m.collection = NewCollection()

	m.collection.Type = "find"
	if m.schema.TableName == "" {
		return m.collection.SetError(errors.New("table is not set"))
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("prepare fail: %v", err))
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("query fail: %v", err))
	}
	defer rows.Close()

	if !rows.Next() {
		return m.collection.SetExist(false)
	}

	ret, err := m.toMap(rows)
	if err != nil {
		return m.collection.SetExist(true).SetError(fmt.Errorf("res to map fail: %v", err))
	}

	for k, v := range m.schema.FieldNames {
		if err := m.schema.SetFieldValue(k, ret[v.ColumnName]); err != nil {
			return m.collection.SetExist(false).SetError(err)
		}
	}
	err = m.context.emitEvent("query_after", &CallbackParams{
		Action: "find",
		Model: m,
		Rows: rows,
		Map: ret,
	})
	m.collection.Fields = m.builder.resultFields
	m.collection.Data = ret
	return m.collection.SetExist(true).SetError(err)
}

// 获取数据集
func (m *Model) Select() *Collection {
	m.collection = NewCollection()

	m.collection.Type = "select"
	if m.schema.TableName == "" {
		return m.collection.SetError(errors.New("table is not set"))
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("prepare fail: %v", err))
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("query fail: %v", err))
	}
	defer rows.Close()

	maps := make([]map[string]interface{}, 0)
	for rows.Next() {
		ret, err := m.toMap(rows)
		if err != nil {
			return m.collection.SetError(fmt.Errorf("res to map fail: %v", err))
		}
		maps = append(maps, ret)

		m.schema.AddArrayItem(ret)
		m.collection.SetExist(true)
	}

	err = m.context.emitEvent("query_after", &CallbackParams{
		Action: "select",
		Model: m,
		Rows: rows,
		MapRows: maps,
	})

	m.collection.Fields = m.builder.resultFields
	m.collection.Data = maps
	return m.collection.SetError(err)
}

// 获取一列数据
func (m *Model) Value(col string, dst interface{}) *Collection {
	m.collection = NewCollection()

	m.collection.Type = "value"
	if m.schema.TableName == "" {
		return m.collection.SetError(fmt.Errorf("table is not set"))
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("prepare fail: %v", err))
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return m.collection.SetError(fmt.Errorf("query fail: %v", err))
	}
	defer rows.Close()

	if rows.Next() {
		ret, err := m.toMap(rows)
		if err != nil {
			return m.collection.SetExist(true).SetError(fmt.Errorf("res to map fail: %v", err))
		}
		p := reflect.TypeOf(dst)
		if p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		value := reflect.ValueOf(dst)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		utils.CallScan(ret[col], value)

		err = m.context.emitEvent("query_after", &CallbackParams{
			Action: "value",
			Model: m,
			Rows: rows,
			Map: ret,
		})
		m.collection.Data = m.model
		return m.collection.SetExist(true).SetError(fmt.Errorf("res to map fail: %v", err))
	}

	return m.collection.SetExist(false)
}

// 是否存在记录
func (m *Model) Exist() *Collection {
	m.builder.fields = []string{}
	return m.Field(m.schema.PrimaryKey).Find()
}

// 统计
func (m *Model) Count() (int64, error) {
	var dst int64
	m.builder.fields = []string{"COUNT(*) AS __COUNT__"}
	c := m.Value("__COUNT__", &dst)
	return dst, c.Error
}

// 求和
func (m *Model) Sum(col string, dst interface{}) error {
	p := utils.ParseField(m.db.dbConf.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("SUM(%s) AS __SUM__", p)}
	c := m.Value("__SUM__", dst)
	return c.Error
}

// 最大值
func (m *Model) Max(col string, dst interface{}) error {
	p := utils.ParseField(m.db.dbConf.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("MAX(%s) AS __VALUE__", p)}
	c := m.Value("__VALUE__", dst)
	return c.Error
}

// 最小值
func (m *Model) Min(col string, dst interface{}) error {
	p := utils.ParseField(m.db.dbConf.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("MIN(%s) AS __VALUE__", p)}
	c := m.Value("__VALUE__", dst)
	return c.Error
}

// 平均值
func (m *Model) Avg(col string, dst *float64) error {
	p := utils.ParseField(m.db.dbConf.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("AVG(%s) AS __VALUE__", p)}
	c := m.Value("__VALUE__", dst)
	return c.Error
}

// query
func (m *Model) Query(sqlStr string, params ...interface{}) (*sql.Rows, error) {
	var (
		err error
		stmt *sql.Stmt
	)
	db := m.db
	stmt, err = db.Prepare(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, fmt.Errorf("query fail: %v", err)
	}
	return rows, nil
}

// 创建
func (m *Model) Create() error {
	db := m.db
	m.builder.p = "insert"
	fieldNum := len(m.schema.Fields)
	m.builder.data = make(map[string]interface{}, fieldNum)
	for i := 0; i < fieldNum; i++ {
		field := m.schema.Fields[i]
		m.builder.data[field.Name] = m.schema.GetFieldValue(field.Name)
	}
	sqlStr, bindParams := m.builder.ToString()

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()

	var lastId int64

	if m.db.dbConf.Driver == "mssql" {
		result, err := stmt.Query(bindParams...)
		if err != nil {
			return fmt.Errorf("insert exec fail: %v", err)
		}

		if result.Next() {
			_ = result.Scan(&lastId)
		}
	} else {
		result, err := stmt.Exec(bindParams...)
		if err != nil {
			return fmt.Errorf("insert exec fail: %v", err)
		}

		lastId, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("insert getLastInsertId fail: %v", err)
		}
	}

	err = m.schema.SetFieldValue(m.schema.PrimaryKey, lastId)
	if err != nil {
		return err
	}
	err = m.context.emitEvent("insert_after", &CallbackParams{
		Action: "insert",
		Model: m,
	})
	return err
}

// 修改
func (m *Model) Update() error {
	db := m.db
	m.builder.p = "update"
	m.builder.data = utils.StructToMap(m.model)
	if m.builder.where == nil {
		if pkValue := m.schema.GetFieldValue(m.schema.PrimaryKey); pkValue != nil {
			m.Where(m.schema.PrimaryKey, pkValue)
		}
	}
	sqlStr, bindParams := m.builder.ToString()

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(bindParams...)
	if err != nil {
		return fmt.Errorf("query fail: %v", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update fail: %v", err)
	}
	err = m.context.emitEvent("update_after", &CallbackParams{
		Action: "update",
		Model: m,
	})

	return err
}

// 删除
func (m *Model) Delete() error {
	db := m.db
	m.builder.p = "delete"

	if m.builder.where == nil {
		if pkValue := m.schema.GetFieldValue(m.schema.PrimaryKey); pkValue != nil {
			m.Where(m.schema.PrimaryKey, pkValue)
		}
	}
	sqlStr, bindParams := m.builder.ToString()

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(bindParams...)
	if err != nil {
		return fmt.Errorf("query fail: %v", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete fail: %v", err)
	}
	err = m.context.emitEvent("delete_after", &CallbackParams{
		Action: "delete",
		Model: m,
	})

	return err
}

