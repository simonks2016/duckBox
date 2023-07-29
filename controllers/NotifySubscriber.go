package controllers

import (
	"DuckBox/Define"
	"DuckBox/models"
	"encoding/json"
	"github.com/nsqio/go-nsq"
)

type NotifySubscriber struct {
	nsq.Handler
}

// HandleMessage Notify subscribers when episodes are added
func (this *NotifySubscriber) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*models.Episodes]
	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	message.Finish()
	return nil
}
