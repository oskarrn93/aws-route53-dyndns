package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gregdel/pushover"
)

type PushoverNotication struct {
	config    *PushoverConfig
	app       *pushover.Pushover
	recipient *pushover.Recipient
	logger    *slog.Logger
}

func (pn *PushoverNotication) Send(recordName string) {
	pn.logger.DebugContext(context.Background(), "sendPushoverNotification", "app", pn.app, "recipient", pn.recipient, "recordName", recordName)

	message := pushover.NewMessageWithTitle(
		fmt.Sprintf("AWS Route53 DNS record updated for: %s", recordName),
		"DNS record updated",
	)

	response, err := pn.app.SendMessage(message, pn.recipient)
	if err != nil {
		pn.logger.ErrorContext(context.Background(), "Failed to send message using Pushover", "error", err)
		return
	}

	pn.logger.DebugContext(context.Background(), "Pushover send message response", "response", response)
}

func NewPushoverNotication(config *PushoverConfig, logger *slog.Logger) *PushoverNotication {
	app := pushover.New(config.ApiToken)
	recipient := pushover.NewRecipient(config.UserKey)

	return &PushoverNotication{
		config:    config,
		app:       app,
		recipient: recipient,
		logger:    logger,
	}
}
