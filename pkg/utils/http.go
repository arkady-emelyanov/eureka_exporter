package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

//
func WriteMetrics(w io.Writer, metrics []map[string]*io_prometheus_client.MetricFamily) (int, error) {
	var buf bytes.Buffer

	log.Debug().
		Msg("Encoding response...")

	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, m := range metrics {
		for _, v := range m {
			if err := enc.Encode(v); err != nil {
				panic(err)
			}
		}
	}

	log.Info().
		Int("length", buf.Len()).
		Msg("Writing response...")

	return w.Write(buf.Bytes())
}

//
func FetchApps(e models.Endpoint, t time.Duration) []models.Instance {
	b, err := getResponse(e.URL, t)
	if err != nil {
		log.Error().
			Str("url", e.URL).
			Str("namespace", e.Namespace).
			Str("name", e.Name).
			Err(err).
			Msg("Error calling URL")
		return nil
	}

	r := bytes.NewReader(b)
	l, err := parseEurekaResponse(r, e)
	if err != nil {
		log.Error().
			Str("url", e.URL).
			Str("namespace", e.Namespace).
			Str("name", e.Name).
			Err(err).
			Msg("Error parsing XML response")
		return nil
	}

	log.Debug().
		Str("url", e.URL).
		Str("namespace", e.Namespace).
		Str("name", e.Name).
		Int("count", len(l)).
		Err(err).
		Msg("Found Prometheus-enabled applications")

	return l
}

//
func FetchMetrics(e models.Endpoint, t time.Duration) map[string]*io_prometheus_client.MetricFamily {
	b, err := getResponse(e.URL, t)
	if err != nil {
		log.Error().
			Str("url", e.URL).
			Str("namespace", e.Namespace).
			Str("name", e.Name).
			Err(err).
			Msg("Error calling URL")

		return nil
	}

	r := bytes.NewReader(b)
	m, err := parsePromResponse(r, e)
	if err != nil {
		log.Error().
			Str("url", e.URL).
			Str("namespace", e.Namespace).
			Str("name", e.Name).
			Err(err).
			Msg("Error parsing Prometheus response")

		return nil
	}

	return m
}

func getResponse(url string, timeout time.Duration) ([]byte, error) {
	log.Debug().
		Str("url", url).
		Msg("Dialing")

	c := http.Client{Timeout: timeout}
	r, err := c.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().
				Str("url", url).
				Err(err).
				Msg("Error closing response body")
		}
	}()

	return ioutil.ReadAll(r.Body)
}
