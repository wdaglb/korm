package korm

import "fmt"

type Collection struct {
	Type string
	Data interface{}
	Fields map[string]string
	Error error
	Exist bool
}

func NewCollection() *Collection {
	c := &Collection{}
	return c
}

func (c *Collection) SetError(err error) *Collection {
	c.Error = err
	return c
}

func (c *Collection) SetExist(v bool) *Collection {
	c.Exist = v
	return c
}

func (c *Collection) Row() map[string]interface{} {
	row := make(map[string]interface{})
	if c.Type == "find" {
		fmt.Printf("row: %v\n", c.Fields)
	}

	return row
}

func (c *Collection) Rows() []map[string]interface{} {
	rows := make([]map[string]interface{}, 0)
	if c.Type == "select" {
		src := c.Data.([]map[string]interface{})
		for i := range src {
			row := make(map[string]interface{})
			for k := range c.Fields {
				kb := c.Fields[k]
				row[k] = src[i][kb]
			}
			rows = append(rows, row)
		}
	}

	return rows
}
