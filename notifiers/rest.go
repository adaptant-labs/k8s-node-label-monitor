package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"net/url"
)

// REST Endpoint Notifier
type EndpointNotifier struct {
	endpoint string
}

func NewEndpointNotifier(log logr.Logger, endpoint string) (*EndpointNotifier, error) {
	_, err := url.ParseRequestURI(endpoint)
	if err != nil {
		log.Error(err, "failed to validate endpoint URL")
		return nil, err
	}

	return &EndpointNotifier{
		endpoint: endpoint,
	}, nil
}

func (e EndpointNotifier) Notify(log logr.Logger, notification LabelUpdateNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		log.Error(err, "failed to marshal JSON payload")
		return err
	}

	msg := fmt.Sprintf("notifying %s", e.endpoint)
	log.Info(msg)

	_, err = http.Post(e.endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Error(err, "failed to POST to endpoint")
		return err
	}

	return nil
}
