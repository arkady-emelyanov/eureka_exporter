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


## Running on minikube

```
$ minikube start

## deploy system components
$ kubectl create ns prod
$ kubectl apply -f ./demo/cloud-config-service.yml
$ kubectl apply -f ./demo/eureka-service.yml

## wait a little and deploy services
$ kubectl apply -f ./demo/auto-service.yml
$ kubectl apply -f ./demo/moto-service.yml

## check eureka
$ kubectl proxy
$ curl http://localhost:8001/api/v1/namespaces/prod/services/eureka-service:8761/proxy/eureka/apps

## deploy exporter
$ kubectl apply -f ./deployment.yml

## get metrics
$ curl http://$(minikube ip):31000/metrics
```

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
