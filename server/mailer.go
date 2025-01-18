package server

import (
	"context"
	"log"
)

type Mailer interface {
	Send(c context.Context, addr, subj, msg string) error
}

type LogMailer struct{}

func (m LogMailer) Send(_ context.Context, addr, subj, msg string) error {
	log.Printf("To: %s\nSubject: %s\nMesage: %s\n", addr, subj, msg)
	return nil
}

func NewMailer(config Config) Mailer {
	if config.Debug {
		return LogMailer{}
	}
	panic("no production emailer implemented")
}
