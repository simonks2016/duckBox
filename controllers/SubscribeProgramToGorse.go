package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"context"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"github.com/nsqio/go-nsq"
	"github.com/zhenghaoz/gorse/client"
)

type SubscribeProgramToGorse struct {
	nsq.Handler
}
type subscribeProgram struct {
	ProgramId  string `json:"program_id"`
	CustomerId string `json:"customer_id"`
	HappenTime int64  `json:"happen_time"`
}

func (this *SubscribeProgramToGorse) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*subscribeProgram]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	var o = orm.NewOrm()
	var pr DataModel.Program
	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey,
	)

	if len(p.ExtraData.ProgramId) > 0 && len(p.ExtraData.CustomerId) > 0 {

		if err := o.QueryTable(&DataModel.Program{}).Filter("Id", p.ExtraData.ProgramId).One(&pr); err != nil {
			return err
		}
		_, err := cli.InsertFeedback(ctx, []client.Feedback{
			{
				FeedbackType: "like",
				UserId:       Define.MakeItemId("customer", p.ExtraData.CustomerId),
				ItemId:       Define.MakeItemId("program", pr.Id),
			},
		})
		if err != nil {
			return err
		}
	}

	message.Finish()
	return nil
}
