package main

import (
	"github.com/nlopes/slack"
)

const slackToken = ""

// MultiNotification is the notification config supporting multiple types of
// notifications.  It also contains overall notifcation options
type MultiNotification struct {
	Enabled bool
	Slack   SlackNotification
	Email   EmailNotification
}

// Notify all configured notification subsystems
func (mn *MultiNotification) Notify(msgText string) error {
	var err error
	if len(mn.Slack.Channel) > 1 {
		e := mn.Slack.Notify(msgText)
		err = mergeErrors(err, e)
	}
	if len(mn.Email.Recipients) > 0 {
		e := mn.Notify(msgText)
		err = mergeErrors(err, e)
	}
	return err
}

// EmailNotification is used to send email notifications
type EmailNotification struct {
	Recipients []string
}

// Notify sends an email notification
func (en *EmailNotification) Notify(msgText string) error {
	//for _,r:=range en.Recipients{
	//}
	return nil
}

// SlackNotification contains configs to interact with slack.
type SlackNotification struct {
	// Uses global if not specified
	TeamDomain string
	Channel    string

	cli *slack.Client
}

// Notify sends a slack notification
func (sn *SlackNotification) Notify(msgText string) error {
	if sn.cli == nil {
		sn.cli = slack.New(slackToken)
	}

	params := slack.NewPostMessageParameters()
	// channel, timestamp, error
	_, _, err := sn.cli.PostMessage(sn.Channel, msgText, params)
	return err
}
