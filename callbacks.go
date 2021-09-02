package korm

import (
	"database/sql"
	"fmt"
)

type CallbackParams struct {
	Action string
	Model *Model
	MapRows []map[string]interface{}
	Map map[string]interface{}
	Rows *sql.Rows
}

type QueryAfterCallback func(params *CallbackParams) error

func RegisterCallback(ctx *Context)  {
	ctx.AddQueryAfterCallback(func(params *CallbackParams) error {
		if params.Action == "select" {
			for i := 0; i < params.Model.schema.Data.Len(); i++ {
				field := params.Model.schema.Data.Index(i).FieldByName(params.Model.schema.PrimaryKey)
				if err := params.Model.loadRelationDataItem(i, params.MapRows[i]); err != nil {
					return err
				}

				fmt.Printf("id: %v\n", field)
			}
		} else {
			return params.Model.loadRelationDataItem(-1, params.Map)
		}
		// m.loadRelationData(maps)
		return nil
	})

	ctx.AddQueryAfterCallback(func(params *CallbackParams) error {
		return params.Model.fetchRelationDbData()
	})
}
