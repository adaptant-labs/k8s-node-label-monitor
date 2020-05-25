# Kubernetes Node Label Monitor

This tool provides a custom Kubernetes controller for monitoring and notifying changes in the label states of Kubernetes
nodes (labels added, deleted, or updated), and can be run either node-local or cluster-wide. Notifications can be
dispatched to a number of different targets, and can be easily extended or customized through a simple notification
interface.

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
Node Label Monitor for Kubernetes
Usage: k8s-node-label-monitor [flags]

  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -l	Only track changes to the local node
  -n string
    	Notification endpoint to POST updates to
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
