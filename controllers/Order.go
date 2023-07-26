package controllers

import (
	"DuckBox/Cache/ViewModel"
	"DuckBox/Define"
	"DuckBox/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/nsqio/go-nsq"
	"github.com/segmentio/ksuid"
	"time"
)

type OrderController struct {
	nsq.Handler
}

func (this *OrderController) HandleMessage(message *nsq.Message) error {

	var body = message.Body

	var p Define.ICP[models.Order]
	var p1 Define.ICP[Define.CancelOrder]
	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	if p.Action == Define.ActionReview {
		//new p1
		if err := json.Unmarshal(body, &p1); err != nil {
			return err
		}
	}
	switch p.Action {
	//case Define ActionAdd
	case Define.ActionAdd:
		//order is created ,notify to seller
		if err := this.HandleOrderAdd(p.ItemId); err != nil {
			return err
		}
	//case Define.ActionEdit
	case Define.ActionEdit:
		if err := this.switchStatus(p.Status, p.ExtraData); err != nil {
			return err
		}
	//case Define.ActionDelete:
	case Define.ActionReview:
		if err := this.Review(p1.ItemId, p1.ExtraData); err != nil {
			return err
		}

	}

	message.Finish()
	return nil
}

func (this *OrderController) switchStatus(status int, order models.Order) error {

	switch status {
	case Define.StatusCompletePayment:
		return this.HandleOrderCompletePayment(order)

	}

	return nil

}

func (this *OrderController) HandleOrderAdd(orderId string) error {

	return nil
}

func (this *OrderController) HandleOrderCompletePayment(order models.Order) error {

	//If it is a member order
	if order.IsMemberOrder == true {
		return this.AutoDeliveryMember(&order)
	}
	//notify seller
	return this.NotifySeller()
}

func (this *OrderController) AutoDeliveryMember(order *models.Order) error {

	//log
	Log("new-subscribe-member",
		fmt.Sprintln(order),
		LogInfo)
	//orm
	var o = orm.NewOrm()
	//read order
	if err := o.Read(order); err != nil {
		//
		Log("new-subscribe-member", err.Error(), LogError)
		//return error message
		return err
	}

	if order.TradingStatus != models.OrderTradingStatusAlreadyPaid {
		return errors.New("the order status is incorrect")
	}

	if order.Due != order.ActuallyPaid {
		return errors.New("the order was not paid in full")
	}

	if _, err := o.LoadRelated(order, "MemberOrderItem"); err != nil {
		//log
		Log("webhook", err.Error(), LogError)
		//send
		return errors.New("load order item failed")
	}

	var benefitsMap = models.DisplayBenefits()

	for _, item := range order.MemberOrderItem {
		//load
		if _, err := o.LoadRelated(item, "Member"); err != nil {
			//log
			Log("webhook", err.Error(), LogError)
			//return error message
			return errors.New("failed to load member package ")
		}
		if o.QueryTable(&models.SubscribeMember{}).Filter("Order", order).Exist() == true {
			//log
			Log("new-subscribe-member", "The order has generated a member license", LogError)
			//return
			return errors.New("the order has generated a member authorization code")
		}

		//make license
		var newMember = models.SubscribeMember{
			Id:         ksuid.New().String(),
			DeadLine:   item.Member.ExpireTime + time.Now().Unix(),
			CreateTime: time.Now().Unix(),
			Benefits:   item.Member.Benefits,
			Customer:   order.Customer,
			Order:      order,
		}
		//make signature
		newMember.Signature = newMember.MakeSignature()

		// transaction start
		if err := o.Begin(); err != nil {
			return err
		}
		//insert into mysql
		if _, err := o.Insert(&newMember); err != nil {
			//log
			Log("new-subscribe-member", err.Error(), LogError)
			//rollback
			_ = o.Rollback()
			//return err
			return err
		}

		err, benefits := item.Member.ToBenefitsContent()
		if err != nil {
			//log
			Log("new-subscribe-member", err.Error(), LogError)
			_ = o.Rollback()
			//return
			return err
		}

		var benefitsDisplay []string
		for _, benefit := range benefits {
			//
			if val, exist := benefitsMap[benefit]; exist == false {
				benefitsDisplay = append(benefitsDisplay, benefit)
			} else {
				benefitsDisplay = append(benefitsDisplay, val)
			}
		}

		m := ViewModel.NewMember()
		m.Id = newMember.Id
		m.Deadline = newMember.DeadLine
		m.CreateTime = newMember.CreateTime
		m.Benefits = benefitsDisplay
		m.Name = item.Member.Name
		m.State = 1
		m.EnterpriseLevel = item.Member.EnterpriseLevel
		m.WhetherShareMember = item.Member.WhetherShareMember
		m.ShareMemberMaxNumber = item.Member.ShareMemberMaxNumber

		if err = m.Update(); err != nil {
			//log
			Log("new-subscribe-member", err.Error(), LogError)
			//roll back data
			_ = o.Rollback()
			return err
		}

		var c = ViewModel.NewCustomer()
		//copy id
		c.Id = order.Customer.Id
		//load Relationship
		s := c.LoadRelationship("SubscribeMember")
		//add
		err = s.Add(m.GetDataId())
		if err != nil {
			//log
			Log("new-subscribe-member", err.Error(), LogError)
			//log
			_ = o.Rollback()
			return err
		}
	}

	order.TradingStatus = models.OrderTradingStatusShipped
	order.Status = 1
	order.UpdateTime = time.Now().Unix()

	//update order
	if _, err := o.Update(order); err != nil {
		//rollback
		_ = o.Rollback()
		//return
		return err
	}

	//commit
	if err := o.Commit(); err != nil {
		return err
	}

	return nil
}

func (this *OrderController) NotifySeller() error {

	return nil
}

func (this *OrderController) Review(orderId string, cancelOrder Define.CancelOrder) error {

	var o = orm.NewOrm()
	var order models.Order

	if err := o.QueryTable(&models.Order{}).Filter("Id", orderId).One(&order); err != nil {
		return err
	}

	if cancelOrder.IsSelfCancel == false {

		switch cancelOrder.MachineClosingConditions {
		//if overdue payment
		case Define.MachineCancelConditionOverduePayment:
			if order.IsPaid == true && order.Due == order.ActuallyPaid {
				return nil
			}
			order.TradingStatus = models.OrderTradingStatusCancel
			break
		//if overtime not shipped
		case Define.MachineCancelConditionOvertimeNotShipped:
			if order.TradingStatus == models.OrderTradingStatusShipped {
				return nil
			}
			order.TradingStatus = models.OrderTradingStatusCancel
			order.Reimburse = order.ActuallyPaid
			break
		default:
			return nil
		}
		//update order
		if _, err := o.Update(&order); err != nil {
			return err
		}
		return nil
	}
	return nil
}
