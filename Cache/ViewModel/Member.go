package ViewModel

import (
	"github.com/simonks2016/Subway/ViewModel"
)

type Member struct {
	Id                   string   `json:"id"`
	Name                 string   `json:"name"`
	Deadline             int64    `json:"deadline"`
	CreateTime           int64    `json:"create_time"`
	Benefits             []string `json:"benefits"`
	State                int      `json:"state"`
	EnterpriseLevel      bool     `json:"enterprise_level"`
	WhetherShareMember   bool     `json:"whether_share_member"`
	ShareMemberMaxNumber int      `json:"share_member_max_number"`

	ViewModel.ModelOperation[Member] `json:"-"`
}

func NewMember() *Member {
	var m Member
	m.ModelOperation = ViewModel.NewBasicModelOperation[Member](Pool, &m)
	return &m
}

func (this *Member) Add(Id, Name string, Deadline, CreateTime int64, Benefits ...string) error {

	this.Id = Id
	this.Name = Name
	this.Deadline = Deadline
	this.CreateTime = CreateTime
	this.Benefits = Benefits
	this.State = 1

	err := this.Update()
	if err != nil {
		return err
	}
	return err
}
