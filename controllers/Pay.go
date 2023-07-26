package controllers

import "github.com/nsqio/go-nsq"

type PayControllers struct {
	nsq.Handler
}

func (this PayControllers) HandleMessage(message *nsq.Message) error {

	return nil
}

func (this PayControllers) Notify() {}
