package notifiers

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"time"
)

type CronJobNotifier struct {
	clientset   *kubernetes.Clientset
	namespace   string
	cronjobName string
	cronjob		*v1beta1.CronJob
}

func (d CronJobNotifier) Notify(log logr.Logger, notification LabelUpdateNotification) error {
	return d.scheduleCronJob(log, d.cronjobName, d.cronjob)
}

func (d CronJobNotifier) findCronJob(name string) (*v1beta1.CronJob, error) {
	return d.clientset.BatchV1beta1().CronJobs(d.namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (d CronJobNotifier) scheduleCronJob(log logr.Logger, name string, cronjob *v1beta1.CronJob) error {
	client := d.clientset.BatchV1().Jobs(d.namespace)
	suffix := "-" + strconv.Itoa(int(time.Now().Unix()))

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + suffix,
			Namespace: d.namespace,
		},
		Spec: cronjob.Spec.JobTemplate.Spec,
	}

	result, err := client.Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		log.Error(err, "failed to create job")
		return err
	}

	log.Info("Created job %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func NewCronJobNotifier(clientset *kubernetes.Clientset, cronjobName string) (CronJobNotifier, error) {
	var err error

	notifier := CronJobNotifier{
		clientset:   clientset,
		cronjobName: cronjobName,
		namespace:   apiv1.NamespaceAll,
	}

	notifier.cronjob, err = notifier.findCronJob(cronjobName)
	if err != nil {
		return CronJobNotifier{}, errors.New("failed to find matching cronjob")
	}

	return notifier, nil
}
