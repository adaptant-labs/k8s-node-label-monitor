package notifiers

import (
	"encoding/json"
	"github.com/go-logr/logr"
)

// Default log-based notifier
type LogNotifier struct{}

func (l LogNotifier) Notify(log logr.Logger, notification NodeUpdateNotification) error {
	msg, err := json.Marshal(notification)
	if err != nil {
		log.Error(err, "failed to marshal JSON")
		return err
	}

	log.Info(string(msg))
	return nil
}
