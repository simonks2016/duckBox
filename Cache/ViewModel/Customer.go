package ViewModel

import (
	"github.com/simonks2016/Subway/ViewModel"
	"time"
)

type Customer struct {
	Id             string `json:"id"`
	UserName       string `json:"user_name"`
	UserIcon       string `json:"user_icon"`
	UserBackground string `json:"user_background"`
	UserPoster     string `json:"user_poster"`
	Description    string `json:"description"`
	BrandName      string `json:"brand_name"`

	SubscribeMember string `json:"subscribe_member"`
	Program         string `json:"program"`
	Video           string `json:"video"`

	ViewModel.ModelOperation[Customer] `json:"-"`
}

func NewCustomer() *Customer {
	var c Customer
	c.ModelOperation = ViewModel.NewBasicModelOperation[Customer](Pool, &c)
	return &c
}

func (this Customer) HasSubscribeMember() (bool, error) {

	//new controllers
	controllers := this.LoadRelationship("SubscribeMember")
	//
	members, err := controllers.Members()
	if err != nil {
		return false, err
	}

	m := NewMember()
	err, i := m.BatchRead(members...)
	if err != nil {
		return false, err
	}

	for _, m2 := range i {
		if m2.State == 1 && m2.Deadline > time.Now().Unix() {
			return true, nil
		}
	}
	return false, nil
}

func (this Customer) AddSubscribeMember(m *Member) error {

	controllers := this.LoadRelationship("SubscribeMember")
	//add
	err := controllers.Add(m.GetDataId())
	if err != nil {
		return err
	}
	return err
}

func (this Customer) AddProgram(p *Program) error {

	controllers := this.LoadRelationship("Program")
	//add
	return controllers.Add(p.GetDataId())
}

func (this Customer) AddVideo(v *Video) error {

	controllers := this.LoadRelationship("Video")
	//add
	return controllers.Add(v.GetDataId())
}

func (this Customer) RemoveProgram(p *Program) error {

	controllers := this.LoadRelationship("Program")
	//remove
	return controllers.Remove(p.GetDataId())
}

func (this Customer) RemoveVideo(v *Video) error {
	controllers := this.LoadRelationship("Video")
	//remove
	return controllers.Remove(v.GetDataId())
}
