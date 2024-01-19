package DataModel

type Goods struct {
	Id               string `orm:"pk"`
	Inventory        int
	Images           string
	Videos           string
	Poster           string
	SalePrice        float64 `orm:"column(sale_price)"`
	LowestPrice      float64 `orm:"column(lowest_price)"`
	IsVirtualProduct bool    `orm:"column(is_virtual_product)"`
	From             int
	State            int
	Published        int64
	Description      string
	SKU              []*SKU `orm:"reverse(many)"`
}

type SKU struct {
	Id          string `orm:"pk"`
	State       int
	GoodsSpec   string
	ImageURL    string `orm:"column(image_url)"`
	Description string
	Remark      string
	SalePrice   float64
	Inventory   int
	CreateTime  int64
	UpdateTime  int64
	Goods       *Goods `orm:"rel(fk)"`
}

func (this *SKU) TableName() string { return "sku" }
