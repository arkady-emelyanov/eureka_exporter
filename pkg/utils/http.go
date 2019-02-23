package utils

import (
	"bytes"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func GetResponse(url string, timeout time.Duration) ([]byte, error) {
	log.Printf("Calling: %s\n", url)

	c := http.Client{Timeout: timeout}
	r, err := c.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Error closing body for URL: %s, err: %v", url, err)
		}
	}()

	return ioutil.ReadAll(r.Body)
}

//
func WriteMetrics(w io.Writer, metrics []map[string]*io_prometheus_client.MetricFamily) (int, error) {
	var buf bytes.Buffer

	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, m := range metrics {
		for _, v := range m {
			if err := enc.Encode(v); err != nil {
				panic(err)
			}
		}
	}

	log.Printf("Final response length: %d bytes\n", buf.Len())
	return w.Write(buf.Bytes())
}

//
func FetchApps(e models.Endpoint, t time.Duration) []models.Instance {
	b, err := GetResponse(e.URL, t)
	if err != nil {
		log.Printf("Error calling URL: %s %v, skipping...", e.URL, err)
		return nil
	}

	r := bytes.NewReader(b)
	apps, err := ParseEurekaResponse(r, e.Namespace)
	if err != nil {
		log.Printf("Error parsing response: %s, %v, skipping...", e.URL, err)
		return nil
	}

	log.Printf("Found %d application(s), namespace: %s\n", len(apps), e.Namespace)
	return apps
}

//
func FetchMetrics(e models.Endpoint, t time.Duration) map[string]*io_prometheus_client.MetricFamily {
	b, err := GetResponse(e.URL, t)
	if err != nil {
		log.Printf("Error calling endpoint: %s, %v, skipping...", e.URL, err)
		return nil
	}

	r := bytes.NewReader(b)
	m, err := ParsePromResponse(r, e.Name, e.Namespace)
	if err != nil {
		log.Printf("Error parsing response: %s, %v, skipping...", e.URL, err)
		return nil
	}

	return m
}
