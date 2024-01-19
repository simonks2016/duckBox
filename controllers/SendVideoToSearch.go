package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/orm"
	"github.com/meilisearch/meilisearch-go"
	"github.com/nsqio/go-nsq"
	"strings"
)

type HandlerVideoToSendSearch struct {
	nsq.Handler
}

func (this *HandlerVideoToSendSearch) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[DataModel.Program]

	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	switch p.Action {
	case Define.ActionAdd:
		//ToDO ActionAdd
		if p.Status == Define.StatusComplete {
			if err := this.HandleVideoAdd(&p); err != nil {
				//log
				Log("handle-video", err.Error(), LogError)
				//return error
				return err
			}
		}
	case Define.ActionEdit:
		//ToDO ActionEdit
		if err := this.HandleVideoEdit(&p); err != nil {
			//log
			Log("handle-video", err.Error(), LogError)
			//return
			return err
		}
	case Define.ActionDelete:
		//ToDO ActionDelete
		if err := this.HandleVideoDeleted(&p); err != nil {
			//log
			Log("handle-video", err.Error(), LogError)
			//return
			return err
		}
	}
	message.Finish()

	return nil
}

func (this *HandlerVideoToSendSearch) HandleVideoDeleted(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(strings.ToLower(p.ItemType), "video") != 0 {
		return errors.New("incorrect item type")
	}
	return this.removeDocument(p.ItemId)
}

func (this *HandlerVideoToSendSearch) HandleVideoEdit(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(p.ItemType, "video") != 0 {
		return errors.New("incorrect item type")
	}
	return this.updateSearchClient(p.ItemId)
}

func (this *HandlerVideoToSendSearch) HandleVideoAdd(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(p.ItemType, "program") != 0 {
		return errors.New("incorrect item type")
	}
	return this.updateSearchClient(p.ItemId)

}

func (this *HandlerVideoToSendSearch) updateSearchClient(videoId string) error {

	var o = orm.NewOrm()
	var video DataModel.Video

	if err := o.QueryTable(&DataModel.Video{}).Filter("Id", videoId).One(&video); err != nil {
		return err
	}

	if video.State != 1 {
		return nil
	}

	if _, err := o.LoadRelated(&video, "Applicant"); err != nil {
		return err
	}
	//
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    conf.AppConfig.MeiliSearch.ToHost(),
		APIKey:  conf.AppConfig.MeiliSearch.ApiKey,
		Timeout: 1000 * 60 * 5,
	})

	_, err := client.Index(MeiliSearchIndexVideo).AddDocuments(&Define.VideoSearchModel{
		Title:       video.Title,
		Description: video.Description,
		Id:          video.Id,
		CreateTime:  video.Published,
		Viewer:      video.Viewer,
		CreatorId:   video.Applicant.Id,
		CreatorName: video.Applicant.Username,
	}, "id")
	if err != nil {
		return err
	}
	return nil
}

func (this *HandlerVideoToSendSearch) removeDocument(videoId string) error {

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.AppConfig.MeiliSearch.ToHost(),
		APIKey: conf.AppConfig.MeiliSearch.ApiKey,
	})

	_, err := client.Index(MeiliSearchIndexVideo).DeleteDocument(videoId)
	if err != nil {
		return err
	}

	return nil
}
