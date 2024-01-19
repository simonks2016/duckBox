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

type SendFeedback2Gorse struct {
	nsq.Handler
}

func (this *SendFeedback2Gorse) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Record]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	if len(p.ExtraData.CustomerId) > 0 {

		if p.ItemType == "video" || p.ItemType == "program" {
			var cli = client.NewGorseClient(
				conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey,
			)
			var ctx = context.TODO()

			_, err := cli.InsertFeedback(ctx, []client.Feedback{
				{
					FeedbackType: "view",
					UserId:       Define.MakeItemId("customer", p.ExtraData.CustomerId),
					ItemId:       Define.MakeItemId(p.ExtraData.ItemType, p.ExtraData.ItemId),
					//Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
				},
			})
			if err != nil {
				return err
			}
		}
	}

	message.Finish()
	return nil
}
