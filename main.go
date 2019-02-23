package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

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

	namespace *string
	selector  *string
)

func main() {
	// global
	namespace = flag.StringP("namespace", "n", "", "Namespace to search, default: search all")
	selector = flag.StringP("selector", "s", labelSelector, "Eureka service selector")

	// local
	verbose := flag.BoolP("debug", "d", false, "Display debug output")
	port := flag.IntP("port", "p", 8080, "Server listen port")
	help := flag.BoolP("help", "h", false, "Display help")
	test := flag.BoolP("test", "t", false, "Test, do not run webserver, discover and exit (requires 'kubectl proxy')")
	flag.Parse()

	// help requested?
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// detecting k8s cluster
	inCluster = kube.InCluster()
	if inCluster == false {
		log.Info().Msg("Running outside of Kubernetes cluster, make sure `kubectl proxy` is running...")
	} else {
		log.Info().Msg("Kubernetes cluster detected.")
	}

	// adjusting level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *test {
		metrics, err := collectMetrics()
		if err != nil {
			log.Error().Err(err).Msg("Failed to collect metrics")
			os.Exit(1)
		}
		if _, err := utils.WriteMetrics(os.Stdout, metrics); err != nil {
			log.Error().Err(err).Msg("Failed to writing metrics")
			os.Exit(1)
		}
	} else {
		// formatting listen address
		addr := fmt.Sprintf(":%d", *port)
		log.Info().
			Str("addr", addr).
			Msg("Listening")

		if err := http.ListenAndServe(addr, http.HandlerFunc(promHandler)); err != nil {
			log.Fatal().
				Err(err).
				Str("addr", addr).
				Msg("Failed to start http server")
		}
	}
}

func promHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("uri", r.RequestURI).
		Msg("New collect request")

	metrics, err := collectMetrics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte{}); err != nil {
			log.Error().Str("uri", r.RequestURI).Err(err).Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := utils.WriteMetrics(w, metrics); err != nil {
		log.Error().
			Str("uri", r.RequestURI).
			Err(err).
			Msg("Failed to write response")
	}
}

func collectMetrics() ([]map[string]*io_prometheus_client.MetricFamily, error) {
	svcList, err := utils.DiscoverServices(*namespace, *selector, inCluster)
	if err != nil {
		log.Error().Str("selector", *selector).Err(err).Msg("Failed to discover")
	}

	log.Info().
		Int("found", len(svcList)).
		Msg("Eureka discovery finished")

	appList := getApps(svcList)
	metrics := getMetrics(appList)
	return metrics, nil
}

// for a given list of service endpoints call eureka and parse response
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

// for a given list of endpoints, call and parse prometheus metrics
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
