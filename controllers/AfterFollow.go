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

type AfterFollow struct {
	nsq.Handler
}

func (this *AfterFollow) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Follow]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	var o = orm.NewOrm()
	if err := o.Read(p.ExtraData); err != nil {
		return err
	}

	var subscriber []string
	var followers []*DataModel.Follow

	if _, err := o.QueryTable(&DataModel.Follow{}).Filter("leader_id", p.ExtraData.Leader.Id).All(&followers); err != nil {
		return err
	}

	for _, follower := range followers {
		//if the follower
		if follower.Followers == nil {
			if _, err := o.LoadRelated(follower, "Followers"); err != nil {
				return err
			}
		}

		subscriber = append(subscriber,
			Define.MakeItemId("customer", follower.Followers.Id))
	}

	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey,
	)
	var ctx = context.TODO()

	_, err := cli.UpdateUser(ctx, Define.MakeItemId("customer", p.ExtraData.Leader.Id), client.UserPatch{
		Subscribe: subscriber,
	})
	if err != nil {
		return err
	}

	message.Finish()
	return nil
}
