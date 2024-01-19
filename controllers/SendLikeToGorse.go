package controllers

import (
	"DuckBox/Define"
	"DuckBox/conf"
	"context"
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"github.com/zhenghaoz/gorse/client"
	"strings"
)

type SendLike2Gorse struct {
	nsq.Handler
}

type GiveLikeParams struct {
	ItemId     string `json:"item_id"`
	ItemType   string `json:"item_type"`
	Time       int64  `json:"time"`
	CustomerId string `json:"customer_id"`
}

func (this *SendLike2Gorse) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*GiveLikeParams]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	if strings.ToLower(p.ItemType) == "video" || strings.ToLower(p.ItemType) == "program" {
		//inert feedback
		err := this.sendFeedback(p.ExtraData)
		if err != nil {
			return err
		}
	}

	message.Finish()
	return nil
}

func (this *SendLike2Gorse) sendFeedback(data *GiveLikeParams) error {

	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey,
	)

	_, err := cli.InsertFeedback(ctx, []client.Feedback{
		{
			FeedbackType: "like",
			UserId:       Define.MakeItemId("customer", data.CustomerId),
			ItemId:       Define.MakeItemId(data.ItemType, data.ItemId),
			//Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	if err != nil {
		//logging
		logging.Error(err.Error())
		return err
	}

	return nil
}
