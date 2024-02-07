package DataModel

type Program struct {
	Id               string `orm:"pk"`
	Title            string
	ShowSubTitle     string  `orm:"null"`
	Description      string  `orm:"type(text);size(1000)"`
	Thumb            string  //截图
	Poster           string  `orm:"null"`
	Viewer           int64   //浏览人数
	Subscriber       int64   //订阅用户
	Like             int64   //喜欢人数
	UnLike           int64   //不喜欢人数
	Score            float64 `orm:"null"` //得分
	Evaluation       int64   `orm:"null"` //评价人数
	CreateTime       int64   `orm:"null"`
	UpdateTime       int64   `orm:"null"`
	State            int     `orm:"null"`
	TopLine          bool    `orm:"null"`
	IsAdult          bool    `orm:"null"`
	IsSubscribeTopic bool
	Price            float64 `orm:"null"`
	CopyrightHolder  string
	CopyrightLicense string
	Applicant        *Customer `orm:"rel(fk)"`
	Tags             []*Tag    `orm:"rel(m2m);rel_table(topic_tags)"`
}
