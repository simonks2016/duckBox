package models

import (
	"DuckBox/conf"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {

	//生成mysql 连接字符串
	dbConn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
		conf.AppConfig.Mysql.Account,
		conf.AppConfig.Mysql.Password,
		conf.AppConfig.Mysql.Host,
		conf.AppConfig.Mysql.Port,
		conf.AppConfig.Mysql.DB)

	//注册驱动
	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return
	}
	//注册数据库
	if err := orm.RegisterDataBase("default", "mysql", dbConn); err != nil {
		return
	}

	orm.RegisterModel(new(Program))
	orm.RegisterModel(new(Episodes))
	orm.RegisterModel(new(Customer))
	orm.RegisterModel(new(Order))
	orm.RegisterModel(new(OrderItem))
	orm.RegisterModel(new(MemberPackage))
	orm.RegisterModel(new(MemberOrderItem))
	orm.RegisterModel(new(Video))
	orm.RegisterModel(new(PaymentRecord))
	orm.RegisterModel(new(Employee))
	orm.RegisterModel(new(Goods))
	orm.RegisterModel(new(SKU))
	orm.RegisterModel(new(SubscribeMember))
	orm.RegisterModel(new(SystemNotification))
	orm.RegisterModel(new(InteractiveMessage))

	_ = orm.RunSyncdb("default", false, true)
}
