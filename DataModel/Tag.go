package DataModel

type Tag struct {
	Name       string
	Id         string     `orm:"pk"`
	AboutVideo []*Video   `orm:"reverse(many)"`
	AboutTopic []*Program `orm:"reverse(many)"`
}
