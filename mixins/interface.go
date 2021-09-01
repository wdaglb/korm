package mixins

type Scanner interface {
	Scan(src interface{}) error
}

type ModelTable interface {
	Table() string
}

type ModelPk interface {
	Pk() string
}
