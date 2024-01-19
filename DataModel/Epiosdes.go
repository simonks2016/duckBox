package DataModel

type Episodes struct {
	Id         string `orm:"pk"`
	SortNumber int
	Stage      string
	Program    *Program `orm:"rel(fk);null"`
	Video      *Video   `orm:"rel(fk);null"`
}
