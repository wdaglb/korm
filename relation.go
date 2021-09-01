package korm

import (
	"reflect"
)

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
	for i := range m.withList {
		name := m.withList[i]
		if mod := m.schema.Relations[name]; mod != nil {
			field := m.schema.GetFieldName(name)
			if field == nil {
				continue
			}
			var (
				val reflect.Value
				srcValue reflect.Value
				data interface{}
			)
			if index > -1 {
				srcValue = m.schema.Data.Index(index).FieldByName(field.Name)
			} else {
				srcValue = m.schema.Data.FieldByName(field.Name)
			}
			val = reflect.New(field.IndirectFieldType)

			data = val.Interface()

			pk := field.GetPrimaryKey()
			fk := field.GetForeignKey()
			if ok, err := m.context.Model(data).Where(fk, item[pk]).Find(); !ok || err != nil {
				return err
			}

			if srcValue.Kind() != reflect.Ptr && val.Kind() == reflect.Ptr {
				srcValue.Set(val.Elem())
			} else {
				srcValue.Set(val)
			}
		}
	}

	return nil
}
