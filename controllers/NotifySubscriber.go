package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"encoding/json"
	"github.com/nsqio/go-nsq"
)

type NotifySubscriber struct {
	nsq.Handler
}

// HandleMessage Notify subscribers when episodes are added
func (this *NotifySubscriber) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Episodes]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	message.Finish()
	return nil
}
