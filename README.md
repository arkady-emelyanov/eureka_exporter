# Eureka Prometheus Exporter

Experimental project for collecting metrics from 
Eureka attached services running inside of Kubernetes cluster

## Minikube

Create minikube cluster:
```
$ minikube status
minikube: Running
cluster: Running
kubectl: Correctly Configured: pointing to minikube-vm at 192.168.99.100
```

Deploy `fake-eureka` and `fake-exporter` pods to Kubernetes cluster:
```
$ make fake-build
$ make fake-apply
```

Build and deploy `eureka-exporter` pod to Kubernetes cluster:
```
$ make mini-build
$ make mini-apply
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

Check services:
```
> kubectl get svc --all-namespaces -l subject=eureka-exporter
NAMESPACE     NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
cluster-one   eureka            ClusterIP   10.107.49.151   <none>        8761/TCP   1m
cluster-two   eureka            ClusterIP   10.97.143.88    <none>        8761/TCP   1m
monitoring    eureka-exporter   ClusterIP   10.98.56.242    <none>        8080/TCP   1m
```

### Curl check

1) Run `kube proxy`
2) Run `curl -s http://localhost:8001/api/v1/namespaces/monitoring/services/eureka-exporter:prometheus/proxy/`

Observe results:
```
> curl -s http://localhost:8001/api/v1/namespaces/monitoring/services/eureka-exporter:prometheus/proxy/
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

### Prometheus

prometheus.yml
```
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

scrape_configs:
  - job_name: 'eureka_exporter'
    metrics_path: '/api/v1/namespaces/monitoring/services/eureka-exporter:prometheus/proxy/'
    static_configs:
    - targets: ['localhost:8001']
```


## Next steps

* Configuration options support
* Tests
