package DataModel

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
}

type Follow struct {
	Id         string    `orm:"pk"`
	Followers  *Customer `orm:"rel(fk);column(followers_id)"`
	Leader     *Customer `orm:"rel(fk);column(leader_id)"`
	CreateTime int64
	State      int
}
