package controllers

import (
	"DuckBox/Cache/ViewModel"
	"DuckBox/Define"
	"DuckBox/models"
	"encoding/json"
	"errors"
	"github.com/nsqio/go-nsq"
)

type VideoCacheControllers struct {
	nsq.Handler
}

func (this *VideoCacheControllers) HandleMessage(message *nsq.Message) error {

	var p Define.ICP[models.Video]
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
			return err
		}
	case Define.ActionEdit:
		if err := this.updateVideo(&p.ExtraData, false); err != nil {
			return err
		}
	case Define.ActionDelete:
		if err := this.removeVideo(p.ItemId); err != nil {
			return err
		}
	default:
		return errors.New("we do not support this action")
	}

	message.Finish()
	return nil
}

func (this *VideoCacheControllers) switchStatus(s int, video *models.Video) error {

	switch s {
	case Define.StatusCompleteTranscoding:
		return this.updateVideo(video, true)
	}
	return nil
}

func (this *VideoCacheControllers) updateVideo(v *models.Video, needInsertLine bool) error {

	var video = ViewModel.NewVideo()
	video.Title = v.Title
	video.Id = v.Id
	video.Description = v.Description
	video.CreateTime = v.Published
	video.State = v.State
	video.Thumb = v.Thumb
	video.GIF = v.GIF
	video.Creator = v.Applicant.Id

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
			//add document in line
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
	return video.Remove()
}
