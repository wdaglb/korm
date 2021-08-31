package korm

type Scanner interface {
	Scan(src interface{}) error
}
