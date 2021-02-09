package notifiers

import "github.com/go-logr/logr"

type NodeUpdateNotification struct {
	Node               string            `json:"node"`
	LabelsAdded        map[string]string `json:"labelsAdded"`
	LabelsDeleted      []string          `json:"labelsDeleted"`
	LabelsUpdated      map[string]string `json:"labelsUpdated"`
	AnnotationsAdded   map[string]string `json:"annotationsAdded"`
	AnnotationsDeleted []string          `json:"annotationsDeleted"`
	AnnotationsUpdated map[string]string `json:"annotationsUpdated"`
}

type NodeUpdateNotifier interface {
	Notify(log logr.Logger, notification NodeUpdateNotification) error
}
