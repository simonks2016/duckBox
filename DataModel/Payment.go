package DataModel

type PaymentRecord struct {
	ReceiveAmount  float64 `orm:"default(0);column(receive_amount)"`
	PayTime        int64
	CreateTime     int64
	State          int64
	Id             string `orm:"pk"`
	PaymentChannel string
	Remark         string
	Order          *Order `orm:"rel(fk)"`
}
