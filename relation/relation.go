package relation

type Relation struct {
	Root interface{}
	HasModel interface{}
	PrimaryKey string
	ForeignKey string
}
