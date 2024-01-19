package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"encoding/json"
	"github.com/nsqio/go-nsq"
)

type Record struct {
	nsq.Handler
}

func (this *Record) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Record]
	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	message.Finish()
	return nil
}
