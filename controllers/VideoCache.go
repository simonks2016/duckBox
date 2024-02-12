package controllers

import (
	"DuckBox/Cache/ViewModel"
	"DuckBox/DataModel"
	"DuckBox/Define"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/orm"
	"github.com/nsqio/go-nsq"
)

type VideoCacheControllers struct {
	nsq.Handler
}

func (this *VideoCacheControllers) HandleMessage(message *nsq.Message) error {

	var p Define.ICP[DataModel.Video]
	//json unmarshal
	if err := json.Unmarshal(message.Body, &p); err != nil {
		//log
		Log("update cache", err.Error(), LogError)
		//return
		return err
	}

	switch p.Action {
	case Define.ActionAdd:
		if err := this.switchStatus(p.Status, &p.ExtraData); err != nil {
			//log
			Log("更新缓存时候", err.Error(), LogError)
			//return
			return err
		}
	case Define.ActionEdit:
		if err := this.updateVideo(&p.ExtraData, false); err != nil {
			//log
			Log("更新缓存时候", err.Error(), LogError)
			//return
			return err
		}
	case Define.ActionDelete:
		if err := this.removeVideo(p.ItemId); err != nil {
			//log
			Log("删除缓存时候", err.Error(), LogError)
			//return
			return err
		}
	default:
		return errors.New("we do not support this action")
	}

	message.Finish()
	return nil
}

func (this *VideoCacheControllers) switchStatus(s int, video *DataModel.Video) error {

	switch s {
	case Define.StatusCompleteTranscoding:
		return this.updateVideo(video, true)
	}
	return nil
}

func (this *VideoCacheControllers) updateVideo(v *DataModel.Video, needInsertLine bool) error {

	var o = orm.NewOrm()
	var vm DataModel.Video
	var tags []*ViewModel.Tag
	var program ViewModel.Program

	if err := o.QueryTable(&DataModel.Video{}).Filter("Id", v.Id).One(&vm); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&vm, "Applicant"); err != nil {
		return err
	}

	if _, err := o.LoadRelated(&vm, "Program"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if _, err := o.LoadRelated(&vm, "Tags"); err != nil {
		if !errors.Is(err, orm.ErrNoRows) {
			return err
		}
	}

	if vm.Tags != nil && len(vm.Tags) > 0 {

		for _, tag := range vm.Tags {
			tags = append(tags, &ViewModel.Tag{
				Name: tag.Name,
				Id:   tag.Id,
			})
		}
	}

	if vm.Program != nil {

		program = ViewModel.Program{
			Id:           vm.Program.Id,
			Title:        vm.Program.Title,
			Description:  vm.Program.Description,
			ShowSubtitle: vm.Program.ShowSubTitle,
			Poster:       vm.Program.Poster,
			Score:        vm.Program.Score,
			Creator: &ViewModel.Creator{
				Name:       vm.Applicant.Username,
				Id:         vm.Applicant.Id,
				Icon:       vm.Applicant.UserIcon,
				Background: vm.Applicant.UserBackground,
				BrandName:  vm.Applicant.BrandName,
			},
			IsAdult:      vm.Program.IsAdult,
			IsPayProgram: vm.Program.Price > 0,
			Price:        vm.Program.Price,
			CreateTime:   vm.Program.CreateTime,
			State:        vm.Program.State,
			Evaluation:   vm.Program.Evaluation,
			Viewers:      vm.Program.Viewer,
		}
	}

	var video = ViewModel.NewVideo()

	video.Title = v.Title
	video.Id = v.Id
	video.Description = v.Description
	video.CreateTime = v.Published
	video.State = v.State
	video.Thumb = v.Thumb
	video.GIF = v.GIF
	video.Viewers = v.Viewer
	video.Creator = &ViewModel.Creator{
		Name:       vm.Applicant.Username,
		Id:         vm.Applicant.Id,
		Icon:       vm.Applicant.UserIcon,
		Background: vm.Applicant.UserBackground,
		BrandName:  vm.Applicant.BrandName,
	}
	video.Tags = tags
	video.Program = &program

	if err := video.Update(); err != nil {
		//log
		Log("update cache", err.Error(), LogError)
		//return
		return err
	}

	if err := video.Expire(86400); err != nil {
		//log
		Log("update cache", err.Error(), LogError)
		//return
		return err
	}
	if needInsertLine == true {

		controller := video.CreateLine("CreateTime")
		//if in
		member, err := controller.IsMember(video.GetDataId())
		if err != nil {
			return err
		}
		if !member {
			//add a document in line
			err := controller.Add(video.GetDataId(), float64(video.CreateTime))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *VideoCacheControllers) removeVideo(itemId string) error {

	var video = ViewModel.NewVideo()
	video.Id = itemId

	if exist, err := video.Exist(itemId); err != nil {
		return err
	} else {
		if exist {
			return video.Remove()
		}
		return nil
	}
}
