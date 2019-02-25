# Eureka Prometheus Aggregator and Exporter

Experimental project for collecting metrics from 
[Netflix Eureka](https://github.com/Netflix/eureka) attached application instances 
running inside of Kubernetes cluster.

The goal is to collect metrics outside of Kubernetes (external monitoring).

## Overview
```
                          +---------------------------------------------------------+
                          |Kubernetes cluster                                       |
                          |---------------------------------------------------------|
                          |                                                         |
                          | +---------------+  +---------------+ +----------------+ |
                          | |NS: monitoring |  |NS: staging    | |NS: live        | |
                          | |---------------|  |---------------| |----------------| |
   +------------------+   | |               |  |               | |                | |
   |                  |   | | +-----------+ |  |               | |                | |
   |    Prometheus    +------>|Exporter   +-------------+-----------------+       | |
   |                  |   | | +-----+-----+ |  |        |      | |        |       | |
   +------------------+   | |       |       |  |        |      | |        |       | |
                          | |       |       |  | +------v----+ | | +------v-----+ | |
                          | |       |       |  | |Eureka #1  | | | |Eureka #2   | | |
                          | |       |       |  | +-----------+ | | +------------+ | |
                          | |       |       |  |               | |                | |
                          | |       |       |  | +-----------+ | |                | |
                          | |       +----------->|Service #1 | | |                | |
                          | |               |  | +-----------+ | |                | |
                          | |               |  |               | |                | |
                          | +---------------+  +---------------+ +----------------+ |
                          |                                                         |
                          +---------------------------------------------------------+

```

* Expose `eureka-exporter` endpoint either via `NodePort` or `Ingress`
* Point Prometheus to `eureka-exporter` endpoint
* On each Prometheus collect request, eureka-exporter will:
    * Discover Eureka services across all namespaces or configured namespace
    * Call each found Eureka endpoint and collect attached instances
    * For each instance which exposes promethesURI metadata:
        * Collect metrics
    * Relabel all collected metrics (enrich with `app`, `namespace` and `instanceId` labels)
    * Return all collected and relabeled metrics back to Prometheus

## Requirements

* Minikube up'n'running
* Docker

## Minikube playground

Create minikube cluster:
```
$ minikube start
...

$ minikube status
minikube: Running
cluster: Running
kubectl: Correctly Configured: pointing to minikube-vm at 192.168.99.100
```

Deploy everything to minikube:
```
$ make minikube
```

Make sure pods are running:
```
> kubectl get pods --all-namespaces -l subject=eureka-exporter
NAMESPACE     NAME                               READY     STATUS    RESTARTS   AGE
cluster-one   fake-eureka-7fb76999cc-r5ftb       1/1       Running   0          31s
cluster-one   fake-exporter-5554b8f746-g6b7s     1/1       Running   0          32s
cluster-one   fake-exporter-5554b8f746-ssgtn     1/1       Running   0          32s
cluster-two   fake-eureka-7fb76999cc-582gv       1/1       Running   0          31s
cluster-two   fake-exporter-5554b8f746-mwn8q     1/1       Running   0          32s
cluster-two   fake-exporter-5554b8f746-s5xls     1/1       Running   0          32s
monitoring    eureka-exporter-5cb869d444-wlpkm   1/1       Running   0          24s
```

Make sure all services up and running:
```
> kubectl get svc --all-namespaces -l subject=eureka-exporter
NAMESPACE     NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
cluster-one   eureka            ClusterIP   10.107.49.151   <none>        8761/TCP         4h
cluster-two   eureka            ClusterIP   10.97.143.88    <none>        8761/TCP         4h
monitoring    eureka-exporter   NodePort    10.99.11.12     <none>        8080:31000/TCP   28s
```

### Checking metrics

> fake-exporter is just a Prometheus example app.

```
> minikube ip
192.168.99.100

> curl -s http://192.168.99.100:31000/
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls",quantile="0"} 0
go_gc_duration_seconds{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls",quantile="0.25"} 0
go_gc_duration_seconds{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls",quantile="0.5"} 0
go_gc_duration_seconds{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls",quantile="0.75"} 0
go_gc_duration_seconds{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls",quantile="1"} 0
go_gc_duration_seconds_sum{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls"} 0
go_gc_duration_seconds_count{namespace="cluster-two",app="fake-exporter-5554b8f746-s5xls"} 0
# HELP go_memstats_frees_total Total number of frees.
...
```

Create `prometheus.yml` with following contents
```
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: 'eureka_exporter'
    metrics_path: '/'
    static_configs:
    - targets: ['192.168.99.100:31000']
```

Run `./prometheus` and check Prometheus UI: `localhost:9090`.


## Options

```
> go build
> ./eureka_exporter -h
  -c, --config string      Kubernetes config file path (when running outside of cluster) (default "/Users/user/.kube/config")
  -d, --debug              Display debug output
  -h, --help               Display help
  -l, --listen-port int    Server listen port (default 8080)
  -n, --namespace string   Namespace to search, default: search all
  -s, --selector string    Eureka service selector (default "app=eureka-service")
  -t, --test               Run metric collection write to stdout and exit (requires 'kubectl proxy')
  -o, --timeout int        HTTP call timeout, ms (default 5000)
```
