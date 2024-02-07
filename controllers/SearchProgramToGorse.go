package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/nsqio/go-nsq"
	"github.com/zhenghaoz/gorse/client"
	"strings"
	"time"
)

type SendProgramToGorse struct {
	nsq.Handler
}

func (this *SendProgramToGorse) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Program]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("notify subscriber", err.Error(), LogError)
		//return error message
		return err
	}

	type HandleFunc func(*DataModel.Program) error

	//make the map of handle function
	var h = map[string]HandleFunc{
		Define.ActionAdd:    this.HandleProgramAdd,
		Define.ActionEdit:   this.HandleProgramEdit,
		Define.ActionDelete: this.HandleProgramDelete,
		Define.ActionReview: this.HandleProgramReview,
	}

	if fun, exist := h[p.Action]; !exist {
		//log
		Log("handle program", fmt.Sprintf("the action code is not recognized(%s)", p.Action), LogError)
		//return error message
		return errors.New("the action code is not recognized")
	} else {
		if strings.Compare(p.ItemId, p.ExtraData.Id) != 0 {
			//return error message
			return errors.New("the item ID and data ID in the agreement are not consistent")
		}

		if err := fun(p.ExtraData); err != nil {
			//log the error message
			Log("handle program", err.Error(), LogError)
			//return error message
			return err
		}
	}
	//finish the message
	message.Finish()
	//return not error
	return nil
}

func (this *SendProgramToGorse) HandleProgramAdd(dataModel *DataModel.Program) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var o = orm.NewOrm()
	var p DataModel.Program
	var labels []string

	if err := o.QueryTable(&DataModel.Program{}).Filter("Id", dataModel.Id).One(&p); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&p, "Tags"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}

	}

	if p.Tags != nil && len(p.Tags) > 0 {
		for _, tag := range p.Tags {
			labels = append(labels, tag.Name)
		}
	}

	//insert into gorse client
	_, err := cli.InsertItem(ctx, client.Item{
		ItemId:     Define.MakeItemId("program", p.Id),
		IsHidden:   p.State != DataModel.VideoStatusNormal,
		Labels:     labels,
		Categories: []string{Define.GorseCategoryProgram},
		Timestamp:  time.Unix(p.CreateTime, 0).Format("2006-01-02"),
		Comment:    p.Title,
	})
	if err != nil {
		return err
	}

	return nil
}

func (this *SendProgramToGorse) HandleProgramEdit(dataModel *DataModel.Program) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var o = orm.NewOrm()
	var p DataModel.Program
	var labels []string

	if err := o.QueryTable(&DataModel.Program{}).Filter("Id", dataModel.Id).One(&p); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&p, "Tags"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if p.Tags != nil && len(p.Tags) > 0 {
		for _, tag := range p.Tags {
			labels = append(labels, tag.Name)
		}
	}
	IsHidden := p.State != DataModel.VideoStatusNormal
	//update item
	_, err := cli.UpdateItem(ctx, Define.MakeItemId("program", p.Id),
		client.ItemPatch{
			IsHidden: &IsHidden,
			Labels:   labels,
			Comment:  &p.Title,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (this *SendProgramToGorse) HandleProgramDelete(dataModel *DataModel.Program) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()

	_, err := cli.DeleteItem(ctx, Define.MakeItemId("program", dataModel.Id))
	if err != nil {
		return err
	}
	return nil
}

func (this *SendProgramToGorse) HandleProgramReview(dataModel *DataModel.Program) error {
	return this.HandleProgramDelete(dataModel)
}
