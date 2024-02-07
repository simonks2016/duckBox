package Define

type VideoSearchModel struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Id          string `json:"id"`
	Thumb       string `json:"thumb"`
	CreateTime  int64  `json:"create_time"`
	Viewer      int64  `json:"viewer"`
	CreatorId   string `json:"creator_id"`
	CreatorName string `json:"creator_name"`
	State       int    `json:"state"`
}

type ProgramSearchModel struct {
	Title        string `json:"title"`
	ShowSubtitle string `json:"show_subtitle"`
	Description  string `json:"description"`
	Id           string `json:"id"`
	Poster       string `json:"poster"`
	CreateTime   int64  `json:"create_time"`
	Viewer       int64  `json:"viewer"`
	Subscriber   int64  `json:"subscriber"`
	CreatorId    string `json:"creator_id"`
	CreatorName  string `json:"creator_name"`
	State        int    `json:"state"`
}

type ChannelSearchModel struct {
	ChannelId          string `json:"channel_id"`
	ChannelName        string `json:"channel_name"`
	ChannelDescription string `json:"channel_description"`
	ChannelPoster      string `json:"channel_poster"`
	ChannelSubscriber  int64  `json:"channel_subscriber"`
	CreatorId          string `json:"creator_id"`
	CreatorName        string `json:"creator_name"`
	BrandName          string `json:"brand_name"`
	CreateTime         int64  `json:"create_time"`
	Viewer             int64  `json:"viewer"`
	State              int    `json:"state"`
}
