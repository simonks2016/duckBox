package ViewModel

import (
	"github.com/simonks2016/Subway/ViewModel"
)

type Video struct {
	Title       string   `json:"title"`
	Thumb       string   `json:"thumb"`
	Id          string   `json:"id"`
	CreateTime  int64    `json:"create_time"`
	GIF         string   `json:"gif"`
	Description string   `json:"description"`
	State       int      `json:"state"`
	Creator     *Creator `json:"creator"`
	Tags        []*Tag   `json:"tags"`
	Program     *Program `json:"program"`
	Viewers     int64    `json:"viewers"`

	ViewModel.ModelOperation[Video] `json:"-"`
}

func NewVideo() *Video {

	var p Video
	p.ModelOperation = ViewModel.NewBasicModelOperation[Video](Pool, &p)
	return &p
}

type Tag struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}
