package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_model/go"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/kube"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/utils"
)

const (
	httpTimeoutMs = 5000
	labelSelector = "app=eureka-service"
)

var (
	httpTimeout = time.Duration(time.Duration(httpTimeoutMs) * time.Millisecond)
	inCluster   = false
)

//
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
	http.HandleFunc("/", promHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//
func promHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Collect request: %s", r.RequestURI)
	svcList, err := utils.DiscoverServices(labelSelector, inCluster)
	if err != nil {
		panic(err)
	}

	log.Printf("Found: %d endpoints in cluster\n", len(svcList))
	appList := getApps(svcList)
	metrics := getMetrics(appList)

	if _, err := utils.WriteMetrics(w, metrics); err != nil {
		panic(err)
	}
}

//
func getApps(list []models.Endpoint) []models.Endpoint {
	var wg sync.WaitGroup

	resList := make([]models.Endpoint, 0)
	resChan := make(chan *models.Endpoint)
	done := make(chan bool)

	go func() {
		for {
			e := <-resChan
			if e == nil {
				break
			}
			resList = append(resList, *e)
		}
		done <- true
	}()

	wg.Add(len(list))
	for _, eurekaEndpoint := range list {
		go func(eurekaEndpoint models.Endpoint) {
			apps := utils.FetchApps(eurekaEndpoint, httpTimeout)
			for _, app := range apps {
				if appEndpoint := utils.FormatEndpoint(app, inCluster); appEndpoint != nil {
					resChan <- appEndpoint
				}
			}
			wg.Done()
		}(eurekaEndpoint)
	}

	wg.Wait()
	close(resChan)

	<-done
	return resList
}

//
func getMetrics(list []models.Endpoint) []map[string]*io_prometheus_client.MetricFamily {
	var wg sync.WaitGroup

	resList := make([]map[string]*io_prometheus_client.MetricFamily, 0)
	resChan := make(chan map[string]*io_prometheus_client.MetricFamily)
	done := make(chan bool)

	go func() {
		for {
			m := <-resChan
			if m == nil {
				break
			}
			resList = append(resList, m)
		}
		done <- true
	}()

	wg.Add(len(list))
	for _, appEndpoint := range list {
		go func(appEndpoint models.Endpoint) {
			if m := utils.FetchMetrics(appEndpoint, httpTimeout); m != nil {
				resChan <- m
			}
			wg.Done()
		}(appEndpoint)
	}

	wg.Wait()
	close(resChan)

	<-done
	return resList
}
