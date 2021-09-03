package korm

import (
	"fmt"
	"github.com/wdaglb/korm/schema"
	"reflect"
)

type relation struct {
	Type string
	PrimaryKey string
	ForeignKey string
	Value interface{}
	Field *schema.Field
}

func (m *Model) loadRelationData(params interface{}) error {
	if len(m.withList) == 0 {
		return nil
	}
	if m.schema.IsArray() {
		list := params.([]map[string]interface{})
		for i, item := range list {
			if err := m.loadRelationDataItem(i, item); err != nil {
				return err
			}
		}
	} else {
		return m.loadRelationDataItem(-1, params.(map[string]interface{}))
	}
	return nil
}

// 加载关联的数据
func (m *Model) loadRelationDataItem(index int, item map[string]interface{}) error {
	if m.relationData == nil {
		m.relationData = make(map[string][]*relation)
		m.relationMap = make(map[string]*relation)
	}
	for i := range m.withList {
		name := m.withList[i]
		if mod := m.schema.Relations[name]; mod != nil {
			field := m.schema.GetFieldName(name)
			if field == nil {
				continue
			}

			pk := field.GetPrimaryKey()
			fk := field.GetForeignKey()
			if m.relationData[field.Name] == nil {
				m.relationData[field.Name] = make([]*relation, 0)
			}
			r := &relation{
				Type: mod.Type,
				PrimaryKey: pk,
				ForeignKey: fk,
				Field: field,
				Value: item[pk],
			}
			m.relationData[field.Name] = append(m.relationData[field.Name], r)
			m.relationMap[fmt.Sprintf("%v", item[pk])] = r
		}
	}

	return nil
}

// 获取数据库数据
func (m *Model) fetchRelationDbData() error {
	// 读取数据库
	if len(m.relationData) > 0 {
		for _, v := range m.relationData {
			pks := make([]interface{}, 0)
			var (
				relation *relation
			)
			for _, data := range v {
				relation = data
				pks = append(pks, data.Value)
			}
			if relation == nil {
				continue
			}

			sliceOf := reflect.SliceOf(relation.Field.DeepType)
			ptr := reflect.New(sliceOf)

			ptr.Elem().Set(reflect.MakeSlice(sliceOf, 0, 0))

			if err := m.context.Model(ptr.Interface()).Where(relation.ForeignKey, "in", pks).Select(); err != nil {
				return err
			}

			for i := 0; i < ptr.Elem().Len(); i++ {
				f := ptr.Elem().Index(i)
				id := f.FieldByName(relation.Field.GetForeignName())

				if re := m.relationMap[fmt.Sprintf("%v", id.Interface())]; re != nil {
					row := m.schema.Data.Index(i)
					fieldValue := row.FieldByName(relation.Field.Name)

					if err := m.schema.SetStructValue(f.Interface(), fieldValue); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
