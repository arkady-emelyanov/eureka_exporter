package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/rs/zerolog/log"

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
		log.Info().Msg("Running outside of Kubernetes cluster, make sure `kubectl proxy` is running...")
	} else {
		log.Info().Msg("Kubernetes cluster detected.")
	}
}

func main() {
	// TODO: flags
	log.Info().
		Str("port", "8080").
		Msg("Listening on :8080")

	if err := http.ListenAndServe(":8080", http.HandlerFunc(promHandler)); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start http server")
	}
}

//
func promHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("uri", r.RequestURI).
		Msg("New collect request")

	svcList, err := utils.DiscoverServices(labelSelector, inCluster)
	if err != nil {
		log.Error().Str("selector", labelSelector).Err(err).Msg("Failed to discover")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte{}); err != nil {
			log.Error().Str("uri", r.RequestURI).Err(err).Msg("Failed to write response")
		}
		return
	}

	log.Info().
		Int("found", len(svcList)).
		Msg("Eureka discovery finished")

	appList := getApps(svcList)
	metrics := getMetrics(appList)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := utils.WriteMetrics(w, metrics); err != nil {
		log.Error().
			Str("uri", r.RequestURI).
			Err(err).
			Msg("Failed to write response")
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
