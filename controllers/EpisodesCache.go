package controllers

import (
	"DuckBox/Cache/ViewModel"
	"DuckBox/Define"
	"DuckBox/models"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"github.com/nsqio/go-nsq"
	"math/rand"
	"time"
)

type EpisodesCacheControllers struct {
	nsq.Handler
}

func (this *EpisodesCacheControllers) HandleMessage(message *nsq.Message) error {

	var body = message.Body
	var p Define.ICP[*models.Episodes]
	if err := json.Unmarshal(body, &p); err != nil {
		//return error message
		return err
	}

	if p.ExtraData == nil {
		//finish
		message.Finish()
		//return
		return nil
	}

	switch p.Action {
	case Define.ActionAdd:
		if err := this.UpdateCache(p.ExtraData); err != nil {
			//log
			Log("更新缓存时候,发生错误", err.Error(), LogError)
			//return error
			return err
		}
	case Define.ActionEdit:
		if err := this.UpdateCache(p.ExtraData); err != nil {
			//log
			Log("更新缓存时候,发生错误", err.Error(), LogError)
			//return error
			return err
		}
	case Define.ActionDelete:
		if err := this.removeCache(p.ExtraData); err != nil {
			//log
			Log("删除缓存时候,发生错误", err.Error(), LogError)
			//return
			return err
		}
	}

	message.Finish()
	return nil
}

func (this *EpisodesCacheControllers) UpdateCache(data *models.Episodes) error {

	if data.Program == nil {
		var e models.Episodes
		var o = orm.NewOrm()
		//reload episode data
		if err := o.QueryTable(&models.Episodes{}).Filter("Id", data.Id).One(&e); err != nil {
			//log
			Log("查阅数据时候发生错误", err.Error(), LogError)
			return err
		}
		//load program
		if _, err := o.LoadRelated(&e, "Program"); err != nil {
			//log
			Log("查阅节目信息时候发生错误", err.Error(), LogError)
			return err
		}
		data = &e
	}

	var episodes = ViewModel.NewEpisodes()

	//update cache for redis
	episodes.Video = &ViewModel.EpisodeVideo{
		Id:        data.Video.Id,
		Thumb:     data.Video.Thumb,
		Title:     data.Video.Title,
		Click:     int64(data.Video.Click),
		Published: data.Video.Published,
		State:     data.Video.State,
	}
	episodes.Stage = data.Stage
	episodes.SortNumber = data.SortNumber
	episodes.Id = data.Id
	//update episodes
	err := episodes.Update()
	if err != nil {
		//log
		Log("更新缓存时候发生错误", err.Error(), LogError)
		return err
	}

	rand.Seed(time.Now().UnixNano())
	err = episodes.Expire(int64(86400 * rand.Intn(15)))
	if err != nil {
		//log
		Log("添加过期时间发生错误", err.Error(), LogError)
		return err
	}

	var program = ViewModel.NewProgram()
	//copy program id
	program.Id = data.Program.Id
	// Create episodes relationship controller
	controller := program.CreateRelationship("Episodes")
	//Check if the episode exists
	member, err := controller.IsMember(episodes.GetDataId())
	if err != nil {
		//log
		Log("检查是否存有该分集时候发生错误", err.Error(), LogError)
		return err
	}
	if !member {
		//if is not exist
		err = controller.Add(episodes.GetDataId())
		if err != nil {
			//log
			Log("添加分集时候发生错误", err.Error(), LogError)
			return err
		}
	}
	return nil
}

func (this *EpisodesCacheControllers) removeCache(data *models.Episodes) error {

	var episodes = ViewModel.NewEpisodes()
	episodes.Id = data.Id

	exist, err := episodes.Exist(data.Id)
	if err != nil {
		return err
	}

	if exist {
		err = episodes.Remove()
		if err != nil {
			return err
		}
	}
	//new program view model
	var program = ViewModel.NewProgram()
	program.Id = data.Program.Id
	//create relationship controller for epsiodes
	controller := program.CreateRelationship("Episodes")
	//check cache exist
	member, err := controller.IsMember(episodes.GetDataId())
	if err != nil {
		return err
	}
	//if exist
	if member {
		err = controller.Remove(episodes.GetDataId())
		if err != nil {
			return err
		}
	}
	return nil
}
