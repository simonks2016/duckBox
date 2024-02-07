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

type SendVideoToGorse struct {
	nsq.Handler
}

func (this *SendVideoToGorse) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*DataModel.Video]
	//json unmarshal
	if err := json.Unmarshal(body, &p); err != nil {
		//log
		Log("send video to gorse", err.Error(), LogError)
		//return error message
		return err
	}

	type HandleFunc func(*DataModel.Video) error

	//make the map of handle function
	var h = map[string]HandleFunc{
		Define.ActionAdd:    this.HandleVideoAdd,
		Define.ActionEdit:   this.HandleVideoEdit,
		Define.ActionDelete: this.HandleVideoDelete,
		Define.ActionReview: this.HandleVideoReview,
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

func (this *SendVideoToGorse) HandleVideoAdd(dataModel *DataModel.Video) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var o = orm.NewOrm()
	var v DataModel.Video
	var labels []string
	var category []string

	if err := o.QueryTable(&DataModel.Video{}).
		Filter("Id", dataModel.Id).One(&v); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&v, "Tags"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if _, err := o.LoadRelated(&v, "Program"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if v.Program == nil || len(v.Program.Id) <= 0 {
		category = []string{Define.GorseCategoryEpisode}
	} else {
		category = []string{Define.GorseCategoryVideo}
	}

	fmt.Println(v.Tags)

	if v.Tags != nil && len(v.Tags) > 0 {
		for _, tag := range v.Tags {
			labels = append(labels, tag.Name)
		}
	}
	//insert into gorse client
	_, err := cli.InsertItem(ctx, client.Item{
		ItemId:     Define.MakeItemId("video", v.Id),
		IsHidden:   v.State != DataModel.VideoStatusNormal,
		Labels:     labels,
		Categories: category,
		Timestamp:  time.Unix(v.Published, 0).Format("2006-01-02"),
		Comment:    v.Title,
	})
	if err != nil {
		return err
	}

	return nil
}

func (this *SendVideoToGorse) HandleVideoEdit(dataModel *DataModel.Video) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()
	var o = orm.NewOrm()
	var v DataModel.Video
	var labels []string
	var category []string

	if err := o.QueryTable(&DataModel.Video{}).
		Filter("Id", dataModel.Id).One(&v); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&v, "Tags"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if _, err := o.LoadRelated(&v, "Program"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if v.Program == nil || len(v.Program.Id) <= 0 {
		category = []string{Define.GorseCategoryVideo}
	} else {
		category = []string{Define.GorseCategoryEpisode}
	}

	if v.Tags != nil && len(v.Tags) > 0 {
		for _, tag := range v.Tags {
			labels = append(labels, tag.Name)
		}
	}

	IsHidden := v.State != DataModel.VideoStatusNormal
	//insert into gorse client
	_, err := cli.UpdateItem(ctx, Define.MakeItemId("video", v.Id), client.ItemPatch{
		IsHidden:   &IsHidden,
		Comment:    &v.Title,
		Categories: category,
		Labels:     labels,
	})
	if err != nil {
		return err
	}

	return nil
}

func (this *SendVideoToGorse) HandleVideoDelete(dataModel *DataModel.Video) error {

	var cli = client.NewGorseClient(conf.AppConfig.Gorse.ToEndPoint(), conf.AppConfig.Gorse.ApiKey)
	var ctx = context.TODO()

	_, err := cli.DeleteItem(ctx, Define.MakeItemId("video", dataModel.Id))
	if err != nil {
		return err
	}
	return nil
}

func (this *SendVideoToGorse) HandleVideoReview(dataModel *DataModel.Video) error {

	return this.HandleVideoDelete(dataModel)
}
