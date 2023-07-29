package ViewModel

import "github.com/simonks2016/Subway/ViewModel"

type Episodes struct {
	Id         string        `json:"id"`
	Stage      string        `json:"stage"`
	SortNumber int           `json:"sort_number"`
	Video      *EpisodeVideo `json:"video"`

	ViewModel.ModelOperation[Episodes] `json:"-"`
}

type EpisodeVideo struct {
	Id        string `json:"id"`
	Thumb     string `json:"thumb"`
	Title     string `json:"title"`
	Click     int64  `json:"click"`
	Published int64  `json:"published"`
	State     int    `json:"state"`
}

func NewEpisodes() *Episodes {
	var e Episodes
	e.ModelOperation = ViewModel.NewBasicModelOperation[Episodes](Pool, &e)
	return &e
}
