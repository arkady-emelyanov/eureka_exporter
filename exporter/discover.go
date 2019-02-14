package exporter

import (
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Enable GCP auth

	"github.com/arkady-emelyanov/eureka_exporter/pkg/kube"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

const (
	// eureka
	inClusterEndpointFmt  = "http://%s.%s:%d"
	outClusterEndpointFmt = "http://localhost:8001/api/v1/namespaces/%s/services/%s:%d/proxy"

	// application
	inClusterAppUrlFmt  = "http://%s:%s%s"
	outClusterAppUrlFmt = "http://localhost:8001/api/v1/namespaces/%s/pods/%s:%s/proxy%s"
)

type Endpoint struct {
	URL       string
	Name      string
	Namespace string
}

// GetEurekaUrlList returns list of Eureka services found across all namespaces
func GetEurekaUrlList(selector string, inCluster bool) ([]Endpoint, error) {
	client, err := kube.GetClient()
	if err != nil {
		return nil, err
	}

	svcList, err := client.CoreV1().Services("").List(metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	endpointList := make([]Endpoint, len(svcList.Items))
	for i, s := range svcList.Items {
		// take only the first port: eureka rest port
		for _, p := range s.Spec.Ports {
			if inCluster {
				// generate in-cluster endpoint
				endpointList[i] = Endpoint{
					Namespace: s.Namespace,
					URL: fmt.Sprintf(
						inClusterEndpointFmt,
						s.Name,
						s.Namespace,
						p.Port,
					),
				}
			} else {
				// local, make sure command `kubectl proxy` is running
				endpointList[i] = Endpoint{
					Namespace: s.Namespace,
					URL: fmt.Sprintf(
						outClusterEndpointFmt,
						s.Namespace,
						s.Name,
						p.Port,
					),
				}
			}
			break
		}
	}

	return endpointList, nil
}

// GetApplicationUrlList returns list of final endpoints for scrape
func GetApplicationUrlList(appList []models.Instance, inCluster bool) []Endpoint {
	endpointList := make([]Endpoint, len(appList))
	for i, app := range appList {
		if app.Port.Enabled == false {
			log.Printf("Insecure port for %s/%s is disabled, skipping..", app.Namespace, app.Name)
			continue
		}

		metricsUri := ""
		for _, m := range app.Metadata {
			if m.PrometheusURI != "" {
				metricsUri = m.PrometheusURI
				break
			}
		}

		if metricsUri == "" {
			log.Printf("No PrometheusURI for %s/%s, skipping..", app.Namespace, app.Name)
			continue
		}

		if inCluster {
			endpointList[i] = Endpoint{
				Name:      app.Name,
				Namespace: app.Namespace,
				URL: fmt.Sprintf(
					inClusterAppUrlFmt,
					app.IpAddress,
					app.Port.Value,
					metricsUri,
				),
			}
		} else {
			endpointList[i] = Endpoint{
				Name:      app.Name,
				Namespace: app.Namespace,
				URL: fmt.Sprintf(
					outClusterAppUrlFmt,
					app.Namespace,
					app.InstanceId,
					app.Port.Value,
					metricsUri,
				),
			}
		}
	}

	return endpointList
}
