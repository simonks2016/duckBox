package DataModel

const (
	ProgramItemType       = "program"
	AdvertisementItemType = "ad"
	VideoItemType         = "video"
	UserItemType          = "user"
	GoodsItemType         = "goods"
)

type Record struct {
	Id           string `orm:"pk"`
	ItemId       string
	ExtraData    string
	ItemType     string `orm:"column(item_type)"`
	Event        string
	HappenTime   int64
	IPAddress    string `orm:"column(ip_address)"`
	CustomerId   string `orm:"column(customer_id);null"`
	TrackSession string `orm:"column(track_session);null"`
}
