package relation

type HasOne struct {
	Model interface{}
	PrimaryKey string
	ForeignKey string
}
