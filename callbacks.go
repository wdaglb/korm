package korm

import "database/sql"

type CallbackParams struct {
	Action string
	Model *Model
	Rows *sql.Rows
}

type QueryAfterCallback func(params *CallbackParams) error
