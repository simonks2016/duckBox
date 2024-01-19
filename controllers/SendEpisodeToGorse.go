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
	"time"
)

type SendEpisodeToGorse struct {
	nsq.Handler
}

func (this *SendEpisodeToGorse) HandleMessage(message *nsq.Message) error {

	var p Define.ICP[*DataModel.Episodes]
	//json unmarshal
	if err := json.Unmarshal(message.Body, &p); err != nil {
		//log
		Log("send video to gorse", err.Error(), LogError)
		//return error message
		return err
	}

	type HandlerFunc func(*DataModel.Episodes) error

	var HandlerFuncMap = map[string]HandlerFunc{
		Define.ActionAdd:    this.HandleAdd,
		Define.ActionEdit:   this.HandleEdit,
		Define.ActionDelete: this.HandleDelete,
		Define.ActionReview: this.HandleReview,
	}

	if f, exist := HandlerFuncMap[p.Action]; !exist {
		//log
		Log("handle program", fmt.Sprintf("the action code is not recognized(%s)", p.Action), LogError)
		//return error message
		return errors.New("the action code is not recognized")
	} else {
		err := f(p.ExtraData)
		if err != nil {
			return err
		}
	}

	//finish the message
	message.Finish()
	//return error message
	return nil
}

func (this *SendEpisodeToGorse) HandleAdd(d *DataModel.Episodes) error {

	var o = orm.NewOrm()
	var e DataModel.Episodes
	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var Labels []string

	if err := o.QueryTable(&DataModel.Episodes{}).Filter("Id", d.Id).One(&e); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&e, "Video"); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&e, "Program"); err != nil {
		return err
	}

	if _, err := o.LoadRelated(e.Video, "Tags"); err != nil {
		return nil
	}

	for _, tag := range e.Video.Tags {

		Labels = append(Labels, tag.Name)
	}

	Labels = append(Labels, e.Program.Title)

	_, err := cli.InsertItem(ctx, client.Item{
		ItemId:     Define.MakeItemId("video", e.Video.Id),
		IsHidden:   e.Video.State != DataModel.VideoStatusNormal,
		Labels:     Labels,
		Categories: []string{Define.GorseCategoryEpisode},
		Timestamp:  time.Unix(e.Video.Published, 0).Format("2006-01-02"),
		Comment:    e.Video.Title,
	})
	if err != nil {
		return err
	}

	return nil
}
func (this *SendEpisodeToGorse) HandleEdit(d *DataModel.Episodes) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var o = orm.NewOrm()
	var ctx = context.TODO()
	var e DataModel.Episodes
	var Labels []string

	if err := o.QueryTable(&DataModel.Episodes{}).Filter("Id", d.Id).One(&e); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&e, "Video"); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&e, "Program"); err != nil {
		return err
	}

	if _, err := o.LoadRelated(e.Video, "Tags"); err != nil {
		return nil
	}

	for _, tag := range e.Video.Tags {

		Labels = append(Labels, tag.Name)
	}

	Labels = append(Labels, e.Program.Title)
	newTitle := fmt.Sprintf("%s - %s", e.Program.Title, e.Video.Title)

	_, err := cli.UpdateItem(ctx, Define.MakeItemId("Video", e.Video.Id), client.ItemPatch{
		Categories: []string{Define.GorseCategoryEpisode},
		Labels:     Labels,
		Comment:    &newTitle,
	})
	if err != nil {
		return err
	}
	return nil
}
func (this *SendEpisodeToGorse) HandleDelete(d *DataModel.Episodes) error {

	var o = orm.NewOrm()
	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var e DataModel.Episodes

	if err := o.QueryTable(&DataModel.Episodes{}).Filter("Id", d.Id).One(&e); err != nil {
		return err
	}
	//local related ship
	if _, err := o.LoadRelated(&e, "Video"); err != nil {
		return err
	}
	//update the item
	_, err := cli.UpdateItem(ctx, Define.MakeItemId("Video", e.Video.Id), client.ItemPatch{
		Categories: []string{Define.GorseCategoryVideo},
	})
	if err != nil {
		return err
	}

	return nil
}
func (this *SendEpisodeToGorse) HandleReview(d *DataModel.Episodes) error {

	return this.HandleDelete(d)
}
