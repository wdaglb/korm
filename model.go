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
	config Config
	model interface{}
	context *Context
	builder *SqlBuilder
	schema *schema.Schema
	withList []string

	relationData map[string][]*relation
	relationMap map[string]*relation
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
	for _, v := range m.withList {
		for _, v2 := range list {
			if v2 != v {
				temps = append(temps, v2)
			}
		}
	}
	m.withList = append(m.withList, temps...)
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
func (m *Model) Find() (bool, error) {
	if m.schema.TableName == "" {
		return false, errors.New("table is not set")
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return false, fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return false, fmt.Errorf("query fail: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	ret, err := m.toMap(rows)
	if err != nil {
		return true, fmt.Errorf("res to map fail: %v", err)
	}

	for k, v := range m.schema.FieldNames {
		if err := m.schema.SetFieldValue(k, ret[v.ColumnName]); err != nil {
			return false, err
		}
	}
	err = m.context.callQueryAfterCallbacks(&CallbackParams{
		Action: "find",
		Model: m,
		Rows: rows,
		Map: ret,
	})
	return true, err
}

// 获取数据集
func (m *Model) Select() error {
	if m.schema.TableName == "" {
		return errors.New("table is not set")
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return fmt.Errorf("query fail: %v", err)
	}
	defer rows.Close()

	maps := make([]map[string]interface{}, 0)
	for rows.Next() {
		ret, err := m.toMap(rows)
		if err != nil {
			return fmt.Errorf("res to map fail: %v", err)
		}
		maps = append(maps, ret)

		m.schema.AddArrayItem(ret)
	}

	return m.context.callQueryAfterCallbacks(&CallbackParams{
		Action: "select",
		Model: m,
		Rows: rows,
		MapRows: maps,
	})
}

// 获取一列数据
func (m *Model) Value(col string, dst interface{}) (bool, error) {
	if m.schema.TableName == "" {
		return false, fmt.Errorf("table is not set")
	}
	db := m.db
	m.builder.p = "select"
	sqlStr, bindParams := m.builder.ToString()
	// sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE id=?", m.table)

	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return false, fmt.Errorf("prepare fail: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(bindParams...)
	if err != nil {
		return false, fmt.Errorf("query fail: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		ret, err := m.toMap(rows)
		if err != nil {
			return true, fmt.Errorf("res to map fail: %v", err)
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

		err = m.context.callQueryAfterCallbacks(&CallbackParams{
			Action: "value",
			Model: m,
			Rows: rows,
			Map: ret,
		})
		return true, err
	}

	return false, nil
}

// 是否存在记录
func (m *Model) Exist() bool {
	m.builder.fields = []string{}
	ok, err := m.Field(m.schema.PrimaryKey).Find()
	return ok && err == nil
}

// 统计
func (m *Model) Count() (int64, error) {
	var dst int64
	m.builder.fields = []string{"COUNT(*) AS __COUNT__"}
	_, err := m.Value("__COUNT__", &dst)
	return dst, err
}

// 求和
func (m *Model) Sum(col string, dst interface{}) error {
	p := utils.ParseField(m.config.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("SUM(%s) AS __SUM__", p)}
	_, err := m.Value("__SUM__", dst)
	return err
}

// 最大值
func (m *Model) Max(col string, dst interface{}) error {
	p := utils.ParseField(m.config.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("MAX(%s) AS __VALUE__", p)}
	_, err := m.Value("__VALUE__", dst)
	return err
}

// 最小值
func (m *Model) Min(col string, dst interface{}) error {
	p := utils.ParseField(m.config.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("MIN(%s) AS __VALUE__", p)}
	_, err := m.Value("__VALUE__", dst)
	return err
}

// 平均值
func (m *Model) Avg(col string, dst *float64) error {
	p := utils.ParseField(m.config.Driver, m.schema.Type, col)
	m.builder.fields = []string{fmt.Sprintf("AVG(%s) AS __VALUE__", p)}
	_, err := m.Value("__VALUE__", dst)
	return err
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

	if m.config.Driver == "mssql" {
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

	_ = m.schema.SetFieldValue(m.schema.PrimaryKey, lastId)

	return nil
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

	return nil
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

	return nil
}

