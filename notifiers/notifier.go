package notifiers

import "github.com/go-logr/logr"

type LabelUpdateNotification struct {
	Node    string            `json:"node"`
	Added   map[string]string `json:"labelsAdded"`
	Deleted []string          `json:"labelsDeleted"`
	Updated map[string]string `json:"labelsUpdated"`
}

type LabelNotifier interface {
	Notify(log logr.Logger, notification LabelUpdateNotification) error
}
