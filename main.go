package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/arkady-emelyanov/eureka_exporter/exporter"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/kube"
)

const (
	eurekaTimeoutMs     = 5000
	eurekaLabelSelector = "app=eureka-service"
)

func getResponse(url string, timeout time.Duration) ([]byte, error) {
	log.Printf("Calling: %s\n", url)

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing body %s => %v", url, err)
		}
	}()

	if err != nil {
		log.Printf("Error calling URL: %s %v", url, err)
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

func main() {
	inCluster := kube.InCluster()
	if inCluster == false {
		log.Println("Running outside of Kubernetes cluster, make sure `kubectl proxy` is running...")
	} else {
		log.Println("Kubernetes cluster detected.")
	}

	log.Println("Searching for eureka services...")
	endpointList, err := exporter.GetEurekaUrlList(eurekaLabelSelector, inCluster)
	if err != nil {
		panic(err)
	}

	log.Printf("Found: %d endpoints in cluster\n", len(endpointList))
	timeout := time.Duration(time.Duration(eurekaTimeoutMs) * time.Millisecond)

	metricFamilyList := make([]map[string]*io_prometheus_client.MetricFamily, 0)

	// TODO: should be concurrent
	for _, endpoint := range endpointList {
		body, err := getResponse(endpoint.URL, timeout)
		if err != nil {
			log.Printf("Error calling URL: %s %v, skipping...", endpoint.URL, err)
			continue
		}

		bodyReader := bytes.NewReader(body)
		appList, err := exporter.ParseEurekaResponse(bodyReader, endpoint.Namespace)
		if err != nil {
			log.Printf("Error parsing response body, URL: %s %v, skipping...", endpoint.URL, err)
			continue
		}

		log.Printf("Found %d application(s) in Eureka response, namespace: %s\n", len(appList), endpoint.Namespace)
		appUrlList := exporter.GetApplicationUrlList(appList, inCluster)
		for _, appEndpoint := range appUrlList {
			body, err := getResponse(appEndpoint.URL, timeout)
			if err != nil {
				log.Printf("Error calling endpoint: %s response: %v, skipping...", appEndpoint.URL, err)
				continue
			}

			// parse and relabel
			metricReader := bytes.NewReader(body)
			metrics, err := exporter.ParsePromResponse(metricReader, appEndpoint.Namespace, appEndpoint.Name)
			if err != nil {
				log.Printf("Error parsing Prometheus response from: %s, %v, skipping...", appEndpoint.URL, err)
				continue
			}

			metricFamilyList = append(metricFamilyList, metrics)
		}
	}

	var buf bytes.Buffer
	metricEncoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range metricFamilyList {
		for _, v := range mf {
			if err := metricEncoder.Encode(v); err != nil {
				panic(err)
			}
		}
	}

	log.Printf("Final response length: %d bytes\n", buf.Len())
	log.Printf("%s...\n", buf.Bytes()[:200])
}
