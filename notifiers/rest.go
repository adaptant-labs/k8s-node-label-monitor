package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"net"
	"net/http"
	"net/url"
	"time"
)

// REST Endpoint Notifier
type EndpointNotifier struct {
	endpoint string
}

func NewEndpointNotifier(log logr.Logger, endpoint string) (*EndpointNotifier, error) {
	log.Info("Enabling REST endpoint notifier", "endpoint", endpoint)

	uri, err := url.ParseRequestURI(endpoint)
	if err != nil || uri.Host == "" {
		// POST requires a valid endpoint scheme, add a safe default if none is provided
		uri, err = url.ParseRequestURI("http://" + endpoint)
		if err != nil {
			return nil, err
		}
	}

	// Validate endpoint connectivity
	conn, err := net.DialTimeout("tcp", uri.Host, 5*time.Second)
	if err != nil {
		return nil, err
	} else {
		conn.Close()
	}

	return &EndpointNotifier{
		endpoint: uri.String(),
	}, nil
}

func (e EndpointNotifier) Notify(log logr.Logger, notification NodeUpdateNotification) error {
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
