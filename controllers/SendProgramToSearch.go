package controllers

import (
	"DuckBox/DataModel"
	"DuckBox/Define"
	"DuckBox/conf"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/meilisearch/meilisearch-go"
	"github.com/nsqio/go-nsq"
	"strings"
)

type HandlerProgramToSendSearch struct {
	nsq.Handler
}

func (this *HandlerProgramToSendSearch) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[DataModel.Program]

	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	switch p.Action {
	case Define.ActionAdd:
		//ToDO ActionAdd
		if err := this.HandleProgramAdd(&p); err != nil {
			//log
			Log("handle-program", err.Error(), LogError)
			//return error
			return err
		}
	case Define.ActionEdit:
		//ToDO ActionEdit
		if err := this.HandleProgramEdit(&p); err != nil {
			//log
			Log("handle-program", err.Error(), LogError)
			//return
			return err
		}
	case Define.ActionDelete:
		//ToDO ActionDelete
		if err := this.HandleProgramDeleted(&p); err != nil {
			//log
			Log("handle-program", err.Error(), LogError)
			//return
			return err
		}
	}
	message.Finish()

	return nil
}

func (this *HandlerProgramToSendSearch) HandleProgramDeleted(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(p.ItemType, "program") != 0 {
		return errors.New("incorrect item type")
	}
	return this.removeDocument(p.ItemId)
}

func (this *HandlerProgramToSendSearch) HandleProgramEdit(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(p.ItemType, "program") != 0 {
		return errors.New("incorrect item type")
	}
	return this.updateSearchClient(p.ItemId)
}

func (this *HandlerProgramToSendSearch) HandleProgramAdd(p *Define.ICP[DataModel.Program]) error {

	if strings.Compare(strings.ToLower(p.ItemType), "program") != 0 {
		return errors.New("incorrect item type")
	}
	return this.updateSearchClient(p.ItemId)

}

func (this *HandlerProgramToSendSearch) updateSearchClient(programId string) error {

	var o = orm.NewOrm()
	var program DataModel.Program

	if err := o.QueryTable(&DataModel.Program{}).Filter("Id", programId).One(&program); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&program, "Applicant"); err != nil {
		return err
	}

	fmt.Println(program.Id)
	//

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   conf.AppConfig.MeiliSearch.ToHost(),
		APIKey: conf.AppConfig.MeiliSearch.ApiKey,
	})

	_, err := client.Index(MeiliSearchIndexProgram).AddDocuments(&Define.ProgramSearchModel{
		Title:        program.Title,
		ShowSubtitle: program.ShowSubTitle,
		Description:  program.Description,
		Id:           program.Id,
		Poster:       program.Poster,
		CreateTime:   program.CreateTime,
		Viewer:       program.Viewer,
		Subscriber:   program.Subscriber,
		CreatorId:    program.Applicant.Id,
		CreatorName:  program.Applicant.Username,
	}, "id")
	if err != nil {
		return err
	}
	return nil
}

func (this *HandlerProgramToSendSearch) removeDocument(programId string) error {

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    conf.AppConfig.MeiliSearch.ToHost(),
		APIKey:  conf.AppConfig.MeiliSearch.ApiKey,
		Timeout: 100,
	})

	_, err := client.Index(MeiliSearchIndexProgram).DeleteDocument(programId)
	if err != nil {
		return err
	}

	return nil
}
