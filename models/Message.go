package models

const (
	InteractiveMessageAt      = 1
	InteractiveMessageReply   = 2
	InteractiveMessageLike    = 3
	InteractiveMessageComment = 4
)

type SystemNotification struct {
	Id          string `orm:"pk"`
	Action      string
	State       int
	Title       string
	Content     string
	PublishTime int64
	FinishTime  int64
	Recipient   *Customer `orm:"rel(fk);null"`
}

type InteractiveMessage struct {
	Id         string `orm:"pk"`
	SourceId   string
	SourceType string
	Action     int
	State      int
	CreateTime int64
	FinishTime int64
	Sender     *Customer `orm:"rel(fk)"`
	Recipient  *Customer `orm:"rel(fk)"`
}
