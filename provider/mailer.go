package provider

import (
	"context"
	"gin-gorm-api/config"
	"log"
)

// A Mailer can send messages with a subject to and address.
type Mailer interface {
	// Send a message msg with a subject sbj to and address addr.
	Send(c context.Context, addr, subj, msg string) error
}

// LogMailer is a basic Mailer that logs messages to the standard output.
type LogMailer struct{}

func (m LogMailer) Send(_ context.Context, addr, subj, msg string) error {
	log.Printf("To: %s\nSubject: %s\nMesage: %s\n", addr, subj, msg)
	return nil
}

// NewMailer returns a Mailer as specified by config.
func NewMailer(config config.Config) Mailer {
	if config.Debug || config.Testing {
		return LogMailer{}
	}
	panic("no production emailer implemented")
}
