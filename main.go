package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
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

var (
	httpTimeout = time.Duration(time.Duration(eurekaTimeoutMs) * time.Millisecond)
	inCluster   = false
)

func init() {
	inCluster = kube.InCluster()
	if inCluster == false {
		log.Println("Running outside of Kubernetes cluster, make sure `kubectl proxy` is running...")
	} else {
		log.Println("Kubernetes cluster detected.")
	}
}

//
func main() {
	log.Printf("Listening on :8080")
	http.HandleFunc("/", prometheusHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func prometheusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Searching for eureka services...")
	eurekaList, err := exporter.GetEurekaUrlList(eurekaLabelSelector, inCluster)
	if err != nil {
		panic(err)
	}

	//
	log.Printf("Found: %d endpoints in cluster\n", len(eurekaList))
	appList := getAppList(eurekaList)
	metrics := getAppMetrics(appList)
	if err := writeMetrics(w, metrics); err != nil {
		panic(err)
	}
}

//
func writeMetrics(w io.Writer, metrics []map[string]*io_prometheus_client.MetricFamily) error {
	var buf bytes.Buffer
	metricEncoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range metrics {
		for _, v := range mf {
			if err := metricEncoder.Encode(v); err != nil {
				panic(err)
			}
		}
	}

	log.Printf("Final response length: %d bytes\n", buf.Len())
	if _, err := w.Write(buf.Bytes()); err != nil {
		panic(err)
	}
	return nil
}

//
func getAppList(list []exporter.Endpoint) []exporter.Endpoint {
	endpointChan := make(chan exporter.Endpoint)
	endpointList := make([]exporter.Endpoint, 0)
	done := make(chan bool)

	go func() {
		for {
			endpoint := <-endpointChan
			if endpoint.IsEmpty() {
				break
			}
			endpointList = append(endpointList, endpoint)
		}
		done <- true
	}()

	var wg sync.WaitGroup
	for _, endpoint := range list {
		wg.Add(1)
		go func(endpoint exporter.Endpoint) {
			found := discoverEndpoints(endpoint, httpTimeout, inCluster)
			for _, e := range found {
				endpointChan <- e
			}
			wg.Done()
		}(endpoint)
	}

	wg.Wait()
	close(endpointChan)

	<-done
	return endpointList
}

//
func getAppMetrics(list []exporter.Endpoint) []map[string]*io_prometheus_client.MetricFamily {
	metricChan := make(chan map[string]*io_prometheus_client.MetricFamily)
	metricList := make([]map[string]*io_prometheus_client.MetricFamily, 0)
	done := make(chan bool)

	go func() {
		for {
			metrics := <-metricChan
			if metrics == nil {
				break
			}
			metricList = append(metricList, metrics)
		}
		done <- true
	}()

	var wg sync.WaitGroup
	for _, endpoint := range list {
		wg.Add(1)
		go func(endpoint exporter.Endpoint) {
			metrics := fetchMetrics(endpoint, httpTimeout, inCluster)
			if metrics != nil {
				metricChan <- metrics
			}
			wg.Done()
		}(endpoint)
	}

	wg.Wait()
	close(metricChan)

	<-done
	return metricList
}

//
func discoverEndpoints(endpoint exporter.Endpoint, timeout time.Duration, inCluster bool) []exporter.Endpoint {
	body, err := getResponse(endpoint.URL, timeout)
	if err != nil {
		log.Printf("Error calling URL: %s %v, skipping...", endpoint.URL, err)
		return nil
	}

	bodyReader := bytes.NewReader(body)
	appList, err := exporter.ParseEurekaResponse(bodyReader, endpoint.Namespace)
	if err != nil {
		log.Printf("Error parsing response body, URL: %s %v, skipping...", endpoint.URL, err)
		return nil
	}

	log.Printf("Found %d application(s) in Eureka response, namespace: %s\n", len(appList), endpoint.Namespace)
	return exporter.FormatAppEndpoints(appList, inCluster)
}

//
func fetchMetrics(endpoint exporter.Endpoint, timeout time.Duration, inCluster bool) map[string]*io_prometheus_client.MetricFamily {
	body, err := getResponse(endpoint.URL, timeout)
	if err != nil {
		log.Printf("Error calling endpoint: %s response: %v, skipping...", endpoint.URL, err)
	}

	metricReader := bytes.NewReader(body)
	metrics, err := exporter.ParsePromResponse(metricReader, endpoint.Namespace, endpoint.Name)
	if err != nil {
		log.Printf("Error parsing Prometheus response from: %s, %v, skipping...", endpoint.URL, err)
	}

	return metrics
}

//
func getResponse(url string, timeout time.Duration) ([]byte, error) {
	log.Printf("Calling: %s\n", url)

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing body for URL: %s, err: %v", url, err)
		}
	}()

	return ioutil.ReadAll(resp.Body)
}
