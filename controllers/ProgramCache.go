package controllers

import (
	"DuckBox/Cache/ViewModel"
	"DuckBox/Define"
	"DuckBox/models"
	"encoding/json"
	"github.com/nsqio/go-nsq"
)

type ProgramCacheControllers struct {
	nsq.Handler
}

func (this *ProgramCacheControllers) HandleMessage(message *nsq.Message) error {

	var p Define.ICP[models.Program]

	if err := json.Unmarshal(message.Body, &p); err != nil {
		return err
	}

	switch p.Action {
	case Define.ActionAdd:
		if err := this.updateCache(&p.ExtraData, true); err != nil {
			return err
		}
	case Define.ActionEdit:
		if err := this.updateCache(&p.ExtraData, false); err != nil {
			return err
		}
	case Define.ActionDelete:
		if err := this.removeProgram(p.ItemId); err != nil {
			return err
		}
	}
	//finish message
	message.Finish()
	//return
	return nil
}

func (this *ProgramCacheControllers) updateCache(p *models.Program, needInsertLine bool) error {

	var program = ViewModel.NewProgram()
	program.Id = p.Id
	program.Title = p.Title
	program.Description = p.Description
	program.IsPayProgram = p.Price > 0
	program.Price = p.Price
	program.CreateTime = p.CreateTime
	program.IsAdult = p.IsAdult
	program.ShowSubtitle = p.ShowSubTitle
	program.Poster = p.Poster
	program.Score = p.Score
	program.Creator = p.Applicant.Id
	program.Evaluation = p.Evaluation
	program.State = p.State

	//update document
	err := program.Update()
	if err != nil {
		return err
	}
	//set expire time
	err = program.Expire(86400)
	if err != nil {
		return err
	}
	if needInsertLine == true {
		//create a line for program
		controller := program.CreateLine("CreateTime")
		//if is exist
		member, err := controller.IsMember(program.GetDataId())
		if err != nil {
			return err
		}
		if !member {
			//add document to line
			err = controller.Add(program.GetDataId(), float64(program.CreateTime))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *ProgramCacheControllers) removeProgram(ItemId string) error {

	var p = ViewModel.NewProgram()
	p.Id = ItemId

	controller := p.CreateLine("CreateTime")
	//is in
	member, err := controller.IsMember(ItemId)
	if err != nil {
		return err
	}
	//假如存在该缓存
	if member {
		err = controller.Remove(ItemId)
		if err != nil {
			return err
		}
	}
	return p.Remove()
}
