# Kubernetes Node Label/Annotation Monitor

[![Docker Pulls](https://img.shields.io/docker/pulls/adaptant/k8s-node-label-monitor.svg)](https://hub.docker.com/repository/docker/adaptant/k8s-node-label-monitor)

This tool provides a custom Kubernetes controller for monitoring and notifying changes in the label and annotation
states of Kubernetes nodes (labels/annotations added, deleted, or updated), and can be run either node-local or
cluster-wide. Notifications can be dispatched to a number of different targets, and can be easily extended or
customized througha simple notification interface.

## Installation

If planning to use the CLI App directly, e.g. for node-local testing, this can be compiled and installed as other Go
projects:

```
$ go get github.com/adaptant-labs/k8s-node-label-monitor
```

For Docker containers and Kubernetes deployment instructions, see below.

## Usage

General usage is as follows:

```
$ k8s-node-label-monitor --help
Node Update Monitor for Kubernetes
Usage: k8s-node-label-monitor [flags]

  -cronjob string
    	Manually trigger named CronJob on label changes
  -endpoint string
    	Notification endpoint to POST updates to
  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -local
    	Only track changes to the local node
  -logging
    	Enable/disable logging (default true)
  -master --kubeconfig
    	(Deprecated: switch to --kubeconfig) The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
```

### Running Node-Local via Docker

Multi-arch Docker images are available on Docker Hub at [adaptant/k8s-node-label-monitor]. These may be run as-is
in-cluster, or out of cluster with an appropriate `KUBECONFIG` passed through.

### Running as a Kubernetes Deployment (Cluster-wide Monitoring)

An example Deployment configuration for cluster-wide monitoring and notification is provided in
`k8s-node-label-monitor-cluster-deployment.yaml`, which can be, as the name implies, directly applied to the running
cluster:

```
$ kubectl apply -f https://raw.githubusercontent.com/adaptant-labs/k8s-node-label-monitor/k8s-node-label-monitor-cluster-deployment.yaml
```

This will create a single Deployment constrained to running on the Kubernetes master. It will further create a special
`node-label-monitor` service account, cluster role, and binding with the permission to list and watch nodes.

### Running as a Kubernetes DaemonSet (Node-Local Monitoring)

An example DaemonSet configuration for node-local monitoring and notification is provided in
`k8s-node-label-monitor-cluster-deployment.yaml`, which can be applied directly:

```
$ kubectl apply -f https://raw.githubusercontent.com/adaptant-labs/k8s-node-label-monitor/k8s-node-label-monitor-node-local-daemonset.yaml
```

This will create a DaemonSet that will run on each node, with each node monitoring and notifying changes to its own
label state directly. It will further create a special `node-label-monitor` service account, cluster role, and binding
with the permission to list and watch nodes.

## Notification Plugins

Notification targets are provided through easily-extensible plugins. At present, the following notification mechanisms
are supported:

| Notifier | Description |
|----------|-------------|
| Logger   | Log-based notification, piggybacking on the default logger instance |
| REST API Endpoint | POSTs the JSON-encoded payload to a defined REST API endpoint  |
| Kubernetes CronJob | Manually trigger a Kubernetes CronJob (such as [descheduler]) |

## Payload Example

```json
{
  "node": "jetson-nano",
  "labelsAdded": {
    "beta.devicetree.org/nvidia-jetson-nano": "1",
    "beta.devicetree.org/nvidia-tegra210": "1",
    "beta.kubernetes.io/arch": "arm64",
    "beta.kubernetes.io/instance-type": "k3s",
    "beta.kubernetes.io/os": "linux",
    "feature.node.kubernetes.io/cpu-cpuid.AES": "true",
    "feature.node.kubernetes.io/cpu-cpuid.ASIMD": "true",
    "feature.node.kubernetes.io/cpu-cpuid.CRC32": "true",
    "feature.node.kubernetes.io/cpu-cpuid.EVTSTRM": "true",
    "feature.node.kubernetes.io/cpu-cpuid.FP": "true",
    "feature.node.kubernetes.io/cpu-cpuid.PMULL": "true",
    "feature.node.kubernetes.io/cpu-cpuid.SHA1": "true",
    "feature.node.kubernetes.io/cpu-cpuid.SHA2": "true",
    "feature.node.kubernetes.io/kernel-config.NO_HZ": "true",
    "feature.node.kubernetes.io/kernel-config.NO_HZ_IDLE": "true",
    "feature.node.kubernetes.io/kernel-config.PREEMPT": "true",
    "feature.node.kubernetes.io/kernel-version.full": "4.9.140-tegra",
    "feature.node.kubernetes.io/kernel-version.major": "4",
    "feature.node.kubernetes.io/kernel-version.minor": "9",
    "feature.node.kubernetes.io/kernel-version.revision": "140",
    "feature.node.kubernetes.io/storage-nonrotationaldisk": "true",
    "feature.node.kubernetes.io/system-os_release.ID": "ubuntu",
    "feature.node.kubernetes.io/system-os_release.VERSION_ID": "18.04",
    "feature.node.kubernetes.io/system-os_release.VERSION_ID.major": "18",
    "feature.node.kubernetes.io/system-os_release.VERSION_ID.minor": "04",
    "k3s.io/hostname": "jetson-nano",
    "k3s.io/internal-ip": "...",
    "kubernetes.io/arch": "arm64",
    "kubernetes.io/hostname": "jetson-nano",
    "kubernetes.io/os": "linux",
    "node.kubernetes.io/instance-type": "k3s"
  },
  "labelsDeleted": [],
  "labelsUpdated": {},
  "annotationsAdded": {
    "flannel.alpha.coreos.com/backend-data": "{\"VtepMAC\":\"42:b2:cc:71:52:e4\"}",
    "flannel.alpha.coreos.com/backend-type": "vxlan",
    "flannel.alpha.coreos.com/kube-subnet-manager": "true",
    "flannel.alpha.coreos.com/public-ip": "...",
    "k3s.io/node-args": "[\"agent\"]",
    "k3s.io/node-config-hash": "...",
    "k3s.io/node-env": "{\"K3S_DATA_DIR\":\"/var/lib/rancher/k3s/data/...\",\"K3S_TOKEN\":\"********\",\"K3S_URL\":\"https://<host>:6443\"}",
    "nfd.node.kubernetes.io/extended-resources": "",
    "nfd.node.kubernetes.io/feature-labels": "beta.devicetree.org/nvidia-jetson-nano,beta.devicetree.org/nvidia-tegra210,cpu-cpuid.AES,cpu-cpuid.ASIMD,cpu-cpuid.CRC32,cpu-cpuid.EVTSTRM,cpu-cpuid.FP,cpu-cpuid.PMULL,cpu-cpuid.SHA1,cpu-cpuid.SHA2,kernel-config.NO_HZ,kernel-config.NO_HZ_IDLE,kernel-config.PREEMPT,kernel-version.full,kernel-version.major,kernel-version.minor,kernel-version.revision,storage-nonrotationaldisk,system-os_release.ID,system-os_release.VERSION_ID,system-os_release.VERSION_ID.major,system-os_release.VERSION_ID.minor",
    "nfd.node.kubernetes.io/master.version": "",
    "nfd.node.kubernetes.io/worker.version": "",
    "node.alpha.kubernetes.io/ttl": "0",
    "volumes.kubernetes.io/controller-managed-attach-detach": "true"
  },
  "annotationsDeleted": [],
  "annotationsUpdated": {}
}

```

## Features and bugs

Please file feature requests and bugs in the [issue tracker][tracker].

## Acknowledgements

This project has received funding from the European Union’s Horizon 2020 research and innovation programme under grant
agreement No 825480 ([SODALITE]).

## License

`k8s-node-label-monitor` is licensed under the terms of the Apache 2.0 license, the full version of which can be found
in the LICENSE file included in the distribution.

[SODALITE]: https://www.sodalite.eu
[tracker]: https://github.com/adaptant-labs/k8s-node-label-monitor/issues
[adaptant/k8s-node-label-monitor]: https://hub.docker.com/repository/docker/adaptant/k8s-node-label-monitor
[descheduler]: https://github.com/kubernetes-sigs/descheduler
