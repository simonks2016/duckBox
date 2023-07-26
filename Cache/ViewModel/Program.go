package ViewModel

import (
	"github.com/simonks2016/Subway/ViewModel"
)

type Program struct {
	Id           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ShowSubtitle string  `json:"show_subtitle"`
	Poster       string  `json:"poster"`
	Score        float64 `json:"score"`
	Creator      string  `json:"creator"`
	IsAdult      bool    `json:"is_adult"`
	IsPayProgram bool    `json:"is_pay_program"`
	Price        float64 `json:"price"`
	CreateTime   int64   `json:"create_time"`
	State        int     `json:"state"`
	Evaluation   int64   `json:"evaluation"`

	Episodes string `json:"episodes"`

	ViewModel.BasicModelOperation[Program] `json:"-"`
}

func NewProgram() *Program {

	var p Program
	p.BasicModelOperation = ViewModel.NewBasicModelOperation[Program](Pool, &p)
	return &p
}
