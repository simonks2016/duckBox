package controllers

import (
	"DuckBox/Define"
	"context"
	"fmt"
	"github.com/zhenghaoz/gorse/client"
	"testing"
	"time"
)

func TestSendLike2Gorse_HandleMessage(t *testing.T) {

	var data = &GiveLikeParams{
		ItemId:     "2YIkspWwOiLJVAcICfVFoSTgU0i",
		ItemType:   "video",
		Time:       time.Now().Unix(),
		CustomerId: "2aisLgyapFSCnLaArYwe1J11YYz",
	}
	var ctx = context.TODO()
	var cli = client.NewGorseClient(
		"http://127.0.0.1:8088", "YimiTV@Recommend@001",
	)

	r, err := cli.InsertFeedback(ctx, []client.Feedback{
		{
			FeedbackType: "like",
			UserId:       Define.MakeItemId("customer", data.CustomerId),
			ItemId:       Define.MakeItemId(data.ItemType, data.ItemId),
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	if err != nil {
		//logging
		fmt.Println(err.Error())
	}

	fmt.Println(r.RowAffected)

}
