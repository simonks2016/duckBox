package DataModel

const (
	VideoStatusTranscoding = 4
	VideoStatusUploading   = 3
	VideoStatusRestricted  = 2
	VideoStatusNormal      = 1
	VideoStatusDeleted     = 0
	VideoStatusError       = 5
)

type Video struct {
	Id          string `orm:"pk"`
	Title       string `orm:"size(15)"`
	Description string `orm:"type(text);size(1000)"`
	Thumb       string `orm:"null"`
	GIF         string `orm:"null;column(gif)"`
	Published   int64
	Duration    float64 `orm:"null"`
	Size        float64 `orm:"null"`

	Click  int   `orm:"default(0)"` //点击人数
	Viewer int64 `orm:"default(0)"` //浏览人数
	Like   int64 `orm:"default(0);column(like)"`
	UnLike int64 `orm:"default(0);column(unlike)"`

	State     int  `orm:"null"`
	IsPrivate bool //是否私有

	Link      string //播放链接
	Container string //播放容器
	Remarks   string
	UploadIP  string    `orm:"column(upload_ip)"`
	Applicant *Customer `orm:"rel(fk)"`
	Tags      []*Tag    `orm:"rel(m2m)"`
	Program   *Program  `orm:"rel(fk)"`
}
