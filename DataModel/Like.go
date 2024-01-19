package DataModel

const (
	LikeCode   = 1
	UnLikeCode = 2
)

type Like struct {
	Id         string `orm:"pk"`
	Type       int
	State      int
	CreateTime int64
	UpdateTime int64
	Applicant  *Customer `orm:"rel(fk)"`
	Video      *Video    `orm:"rel(fk);null"`
	Topic      *Program  `orm:"rel(fk);null"`
}
