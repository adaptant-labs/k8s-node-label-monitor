package main

import (
	"flag"
	"fmt"
	"github.com/adaptant-labs/k8s-node-label-monitor/notifiers"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

var (
	monitorName     = "k8s-node-label-monitor"
	nodeLocal       = false
	logging         = true
	cronjob         = ""
	log             = logf.Log.WithName(monitorName)
	nodeLabels      = map[string]map[string]string{}
	nodeAnnotations = map[string]map[string]string{}
)

type Controller struct {
	indexer   cache.Indexer
	queue     workqueue.RateLimitingInterface
	informer  cache.Controller
	notifiers []notifiers.NodeUpdateNotifier
}

// Compare two maps and determine which key/value pairs have been added, deleted, or updated.
func compareMaps(oldMap map[string]string, newMap map[string]string) (added map[string]string, deleted []string, updated map[string]string) {
	added = map[string]string{}
	deleted = []string{}
	updated = map[string]string{}

	// Compare the old map to the new
	for oldKey, oldValue := range oldMap {
		if val, ok := newMap[oldKey]; ok {
			// The same key exists, but the values differ - record it as updated
			if val != oldValue {
				updated[oldKey] = val
			}
		} else {
			// Key has been removed
			deleted = append(deleted, oldKey)
		}
	}

	// Compare the new map to the old
	for newKey, newValue := range newMap {
		if _, ok := oldMap[newKey]; !ok {
			// If the key does not exist in the old map, record it as added
			added[newKey] = newValue
		}
	}

	return added, deleted, updated
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		indexer:   indexer,
		informer:  informer,
		queue:     queue,
		notifiers: make([]notifiers.NodeUpdateNotifier, 0),
	}
}

func (c Controller) notify(log logr.Logger, notification notifiers.NodeUpdateNotification) error {
	for _, notifier := range c.notifiers {
		err := notifier.Notify(log, notification)
		if err != nil {
			return err
		}
	}

	return nil
}

// Calculate label and annotation changes across each node update
func (c *Controller) nodeUpdateHandler(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Error(err, "Failed to get key")
		return err
	}

	if exists {
		node := obj.(*v1.Node)
		nodeName := node.GetName()
		notification := notifiers.NodeUpdateNotification{
			Node: node.GetName(),
		}

		// Compare the cached label state to the incoming one
		ladded, ldeleted, lupdated := compareMaps(nodeLabels[nodeName], node.Labels)

		notification.LabelsAdded = ladded
		notification.LabelsDeleted = ldeleted
		notification.LabelsUpdated = lupdated

		// Compare the cached annotation state to the incoming one
		aadded, adeleted, aupdated := compareMaps(nodeAnnotations[nodeName], node.Annotations)

		notification.AnnotationsAdded = aadded
		notification.AnnotationsDeleted = adeleted
		notification.AnnotationsUpdated = aupdated

		// Log any label/annotation updates
		if len(aadded) > 0 || len(adeleted) > 0 || len(aupdated) > 0 ||
			len(ladded) > 0 || len(ldeleted) > 0 || len(lupdated) > 0 {
			err := c.notify(log, notification)
			if err != nil {
				log.Error(err, "Failed to dispatch notification")
				return err
			}
		}

		// Remove any previously cached labels
		if nodeLabels[nodeName] != nil {
			for k := range nodeLabels[nodeName] {
				delete(nodeLabels[nodeName], k)
			}
		} else {
			// Ensure the label cache is allocated for this node
			nodeLabels[nodeName] = make(map[string]string)
		}

		// Cache the updated label state
		for k, v := range node.Labels {
			nodeLabels[nodeName][k] = v
		}

		// Remove any previously cached annotations
		if nodeAnnotations[nodeName] != nil {
			for k := range nodeAnnotations[nodeName] {
				delete(nodeAnnotations[nodeName], k)
			}
		} else {
			// Ensure the annotation cache is allocated for this node
			nodeAnnotations[nodeName] = make(map[string]string)
		}

		// Cache the updated annotation state
		for k, v := range node.Annotations {
			nodeAnnotations[nodeName][k] = v
		}
	}

	return nil
}

func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.nodeUpdateHandler(key.(string))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	// If we have failed, requeue the work for later
	c.queue.AddRateLimited(key)
	return true
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	// Let workers stop when we are done
	defer c.queue.ShutDown()

	log.Info("Starting node controller")
	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		log.Info("Timed out waiting for cache sync")
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Info("Stopping node controller")
}

func getNodeName() (string, error) {
	// Within the Kubernetes Pod, the hostname provides the Pod name, rather than the node name, so we pass in the
	// node name via the NODE_NAME environment variable instead.
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) > 0 {
		return nodeName, nil
	}

	// If the NODE_NAME environment variable is unset, fall back on hostname matching (e.g. when running outside of
	// a Kubernetes deployment).
	return os.Hostname()
}

func enqueueNodeUpdate(nodeName string, queue workqueue.RateLimitingInterface) {
	if nodeLocal {
		hostname, err := getNodeName()
		if err != nil {
			log.Error(err, "unable to determine local hostname for node-local monitoring")
			return
		}

		if hostname != nodeName {
			return
		}
	}

	queue.Add(nodeName)
}

func main() {
	var endpoint string

	flag.BoolVar(&nodeLocal, "local", false, "Only track changes to the local node")
	flag.BoolVar(&logging, "logging", true, "Enable/disable logging")
	flag.StringVar(&cronjob, "cronjob", "", "Manually trigger named CronJob on label changes")
	flag.StringVar(&endpoint, "endpoint", "", "Notification endpoint to POST updates to")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Node Update Monitor for Kubernetes\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", monitorName)
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg := config.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(cfg)

	logf.SetLogger(zap.New(zap.UseDevMode(false)))

	if nodeLocal {
		hostname, err := getNodeName()
		if err != nil {
			log.Error(err, "unable to determine local hostname for node-local monitoring")
			return
		}
		msg := fmt.Sprintf("configured for node-local monitoring on %s", hostname)
		log.Info(msg)
	} else {
		log.Info("configured for cluster-wide monitoring")
	}

	// Create the node watcher
	nodeListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "nodes", metav1.NamespaceAll, fields.Everything())

	// Create the workqueue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), monitorName)

	// Monitor when nodes are added to, updated, or deleted from the Cluster
	indexer, informer := cache.NewIndexerInformer(nodeListWatcher, &v1.Node{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				// Add node
				enqueueNodeUpdate(key, queue)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				// Update node
				enqueueNodeUpdate(key, queue)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				// Remove node
				delete(nodeAnnotations, key)
				delete(nodeLabels, key)
				enqueueNodeUpdate(key, queue)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer)

	// Set up the notifiers for this controller
	if logging {
		log.Info("Enabling Logging notifier")
		controller.notifiers = append(controller.notifiers, notifiers.LogNotifier{})
	}

	if len(endpoint) > 0 {
		notifier, err := notifiers.NewEndpointNotifier(log, endpoint)
		if err != nil {
			log.Error(err, "failed to instantiate endpoint notifier")
			return
		}
		controller.notifiers = append(controller.notifiers, notifier)
	}

	if len(cronjob) > 0 {
		notifier, err := notifiers.NewCronJobNotifier(log, clientset, cronjob)
		if err != nil {
			log.Error(err, "failed to instantiate cronjob notifier")
			return
		}
		controller.notifiers = append(controller.notifiers, notifier)
	}

	// Start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)
	select {}
}
