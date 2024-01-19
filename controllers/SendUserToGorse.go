package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"context"
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"github.com/zhenghaoz/gorse/client"
)

type SendUserToGorse struct {
	nsq.Handler
}

func (this *SendUserToGorse) HandleMessage(message *nsq.Message) error {
	var body = message.Body
	var p Define.ICP[*DataModel.Customer]
	var ctx = context.TODO()
	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	if p.Status == Define.StatusCreated {
		_, err := cli.InsertUser(ctx, client.User{
			UserId:    Define.MakeItemId("customer", p.ExtraData.Id),
			Labels:    nil,
			Subscribe: nil,
			Comment:   p.ExtraData.Username,
		})
		if err != nil {
			return err
		}
	}

	message.Finish()
	return nil
}
