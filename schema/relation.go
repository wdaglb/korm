package schema

import "reflect"

type Relation struct {
	HasModel interface{}
	HasType reflect.Type
	Field *Field
}
