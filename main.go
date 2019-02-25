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
	promContentType = "text/plain; version=0.0.4; charset=utf-8"
	labelSelector   = "app=eureka-service"
	httpTimeoutMs   = 5000
)

var (
	inCluster = false       // running inside of kubernetes cluster?
	namespace *string       // namespace for search in, default is search everywhere
	selector  *string       // selector for search, default is labelSelector
	timeoutMs *int          // timeout for every REST operation
	timeout   time.Duration // timeout for every REST operation, duration
)

func main() {
	verbose := flag.BoolP("debug", "d", false, "Display debug output")
	selector = flag.StringP("selector", "s", labelSelector, "Eureka service selector")
	namespace = flag.StringP("namespace", "n", "", "Namespace to search, default: search all")
	timeoutMs = flag.IntP("timeout", "o", httpTimeoutMs, "HTTP call timeout, ms")
	port := flag.IntP("listen-port", "l", 8080, "Server listen port")
	help := flag.BoolP("help", "h", false, "Display help")
	test := flag.BoolP("test", "t", false, "Run metric collection write to stdout and exit (requires 'kubectl proxy')")
	flag.Parse()

	// help requested?
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// setting up global timeout
	timeout = time.Duration(time.Duration(*timeoutMs) * time.Millisecond)

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
		// test run
		// collect metrics and dump all metrics to stdout.
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
		// normal run
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

// web-server handler
func promHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("uri", r.RequestURI).
		Msg("Collect request received")

	metrics, err := collectMetrics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte{}); err != nil {
			log.Error().Str("uri", r.RequestURI).Err(err).Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", promContentType)
	w.WriteHeader(http.StatusOK)

	if _, err := utils.WriteMetrics(w, metrics); err != nil {
		log.Error().
			Err(err).
			Str("uri", r.RequestURI).
			Msg("Failed to write response")
	}

	log.Info().
		Str("uri", r.RequestURI).
		Msg("Collect request done")
}

// discover, scrape and return metrics
func collectMetrics() ([]map[string]*io_prometheus_client.MetricFamily, error) {
	svcList, err := utils.DiscoverServices(*namespace, *selector, timeout, inCluster)
	if err != nil {
		log.Error().
			Err(err).
			Str("namespace", *namespace).
			Str("selector", *selector).
			Msg("Discover error")
		return nil, err
	}

	log.Info().
		Str("namespace", *namespace).
		Str("selector", *selector).
		Msgf("Found %d services", len(svcList))

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
			apps := utils.FetchApps(eurekaEndpoint, timeout)
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
			if m := utils.FetchMetrics(appEndpoint, timeout); m != nil {
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
