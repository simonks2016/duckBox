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
}
