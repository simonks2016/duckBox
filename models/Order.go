package models

const (
	OrderStatusDeleted  = 0
	OrderStatusNormal   = 1
	OrderStatusCreating = 2

	OrderTradingStatusCreated     = 1
	OrderTradingStatusAlreadyPaid = 2
	OrderTradingStatusShipped     = 3
	OrderTradingStatusComplete    = 4
	OrderTradingStatusCancel      = 5
)

type Order struct {
	Id string `orm:"pk"`

	CreateTime    int64 `orm:"default(0)"` //创建时间
	UpdateTime    int64 `orm:"default(0)"` //更新时间
	PaymentTime   int64 `orm:"default(0)"` //付款时间
	Status        int   `orm:"default(1)"` //状态
	TradingStatus int   `orm:"default(0)"` //交易状态
	Tax           float64
	TaxRate       float64
	Total         float64 //合计
	ActuallyPaid  float64 //实付
	Due           float64 //应付
	Reimburse     float64 //退款金额
	DiscountedFee float64 //优惠金额
	ShippingFee   float64 //运费
	IsMemberOrder bool    `orm:"column(is_member_order)"`
	Remarks       string  `orm:"null"`

	PaymentRecord []*PaymentRecord `orm:"reverse(many)"`
	IsPaid        bool

	Customer        *Customer          `orm:"rel(fk);null;column(customer_id)"`
	SalePerson      *Employee          `orm:"rel(fk);null;column(sale_person_id)"`
	OrderItems      []*OrderItem       `orm:"reverse(many)"` //订单详细
	MemberOrderItem []*MemberOrderItem `orm:"reverse(many)"`
	//Refund     []*Refund    `orm:"reverse(many)"` //退款项目
	//Delivery *Delivery `orm:"rel(fk);null;column(delivery_id);size(500)"`
}

type Delivery struct {
	Id               string `orm:"pk;size(500)"`
	Province         string
	City             string
	Area             string
	Street           string
	Address          string
	Contact          string
	Recipient        string
	LogisticsCode    string
	LogisticsCompany string
}
type OrderItem struct {
	Id            int64 `orm:"pk;auto;column(id)"`
	ProductTitle  string
	Count         int64
	OriginalPrice float64 //原来单价
	SalePrice     float64 //现在单价
	Total         float64
	Remarks       string
	SKU           *SKU     `orm:"rel(fk);null"`
	OrderID       *Order   `orm:"rel(fk);null"`
	Product       *Goods   `orm:"rel(fk);null"`
	Topic         *Program `orm:"rel(fk);null"`
}

type MemberOrderItem struct {
	Id            string `orm:"pk;column(id)"`
	Remarks       string `orm:"null"`
	Price         float64
	OriginalPrice float64
	Deadline      int64
	CreateTime    int64
	Member        *MemberPackage `orm:"rel(fk)"`
	Order         *Order         `orm:"rel(fk)"`
}

type Refund struct {
	Id               string `orm:"pk"`
	CreateTime       int64
	UpdateTime       int64
	ArrivalTime      int64
	PaymentChannel   string
	Total            float64
	Status           int
	TradeStatus      int
	LogisticsCompany string
	LogisticsNumber  string
	Order            *Order         `orm:"rel(fk);column(order_id)"`
	SalePerson       *Employee      `orm:"rel(fk);null;column(sale_person_id)"`
	Customer         *Customer      `orm:"rel(fk);null;column(customer_id)"`
	Items            []*RefundItems `orm:"reverse(many)"` //退货单详情
}

type RefundItems struct {
	ID           int64  `orm:"pk;auto"`
	ProductTitle string `orm:"null"`
	Count        float64
	Price        float64
	Total        float64
	Remarks      string
	//Status       int      `orm:"default(0)"`
	RefundOrder *Refund `orm:"rel(fk);null"`
	SKU         *SKU    `orm:"rel(fk);null"`
}
