package models

type Customer struct {
	Id             string `orm:"pk"`
	Account        string
	Username       string
	UserIcon       string
	UserBackground string
	UserPoster     string
	Description    string
	Password       string
	State          int    `orm:"null"`
	BrandName      string `orm:"null"`
	MobilePhone    string `orm:"null;column(mobile_phone)"`
	CountryCode    string `orm:"null;column(country_code)"`
	EmailAddress   string `orm:"column(email_address)"`
	StorageArea    string `orm:"column(storage_area)"`
	Level          int64  `orm:"default(1)"`
	Experience     int64  `orm:"default(0)"`
	Language       string `orm:"default(zh-cn)"`
	RegisterIP     string `orm:"column(register_ip)"`
	OffLine        bool
	Administrators bool
	VideoAmount    int64 `orm:"default(0);column(video_amount)"`
	ProgramAmount  int64 `orm:"default(0);column(program_amount)"`
	FansAmount     int64 `orm:"default(0);column(fans_amount)"`
	FollowAmount   int64 `orm:"default(0);column(follow_amount)"`
	LikesAmount    int64 `orm:"default(0);column(likes_amount)"`
}
