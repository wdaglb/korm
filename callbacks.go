package korm

import (
	"database/sql"
)

type CallbackParams struct {
	Action string
	Model *Model
	MapRows []map[string]interface{}
	Map map[string]interface{}
	Rows *sql.Rows
}

type EventCallback func(params *CallbackParams) error

func RegisterCallback(ctx *Context)  {
	// region ----查询后事件----
	ctx.OnEventQueryAfter(func(params *CallbackParams) error {
		if params.Action == "select" {
			for i := 0; i < params.Model.schema.Data.Len(); i++ {
				_ = params.Model.schema.Data.Index(i).FieldByName(params.Model.schema.PrimaryKey)
				if err := params.Model.loadRelationDataItem(i, params.MapRows[i]); err != nil {
					return err
				}

				// fmt.Printf("id: %v\n", field)
			}
		} else {
			return params.Model.loadRelationDataItem(-1, params.Map)
		}
		// m.loadRelationData(maps)
		return nil
	})

	ctx.OnEventQueryAfter(func(params *CallbackParams) error {
		return params.Model.fetchRelationDbData()
	})
	// endregion

	// region ----插入后事件----
	ctx.OnInsertAfterCallback(func(params *CallbackParams) error {
		return params.Model.insertRelationData()
	})
	// endregion

	// region ----更新后事件----
	ctx.OnUpdateAfterCallback(func(params *CallbackParams) error {
		return params.Model.updateRelationData()
	})
	// endregion

	// region ----删除后事件----
	ctx.OnDeleteAfterCallback(func(params *CallbackParams) error {
		return params.Model.deleteRelationData()
	})
	// endregion
	

}
