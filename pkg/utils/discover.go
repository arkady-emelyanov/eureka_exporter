package utils

import (
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/kube"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

const (
	inClusterEurekaUrlFmt  = "http://%s.%s:%d"
	outClusterEurekaUrlFmt = "http://localhost:8001/api/v1/namespaces/%s/services/%s:%d/proxy"

	inClusterServiceUrlFmt  = "http://%s:%s%s"
	outClusterServiceUrlFmt = "http://localhost:8001/api/v1/namespaces/%s/pods/%s:%s/proxy%s"
)

// DiscoverServices returns list of Eureka services found across all namespaces
func DiscoverServices(selector string, inCluster bool) ([]models.Endpoint, error) {
	svcLabelSelector := metav1.ListOptions{LabelSelector: selector}
	svcList, err := kube.GetClient().CoreV1().Services("").List(svcLabelSelector)
	if err != nil {
		return nil, err
	}

	res := make([]models.Endpoint, len(svcList.Items))

	// take only the first port: eureka rest port
	for i, s := range svcList.Items {
		for _, p := range s.Spec.Ports {
			context := models.Context{
				Namespace: s.Namespace,
				Name:      s.Name,
			}
			if inCluster {
				res[i] = models.Endpoint{
					Context: context,
					URL: fmt.Sprintf(
						inClusterEurekaUrlFmt,
						s.Name,
						s.Namespace,
						p.Port,
					),
				}
			} else {
				res[i] = models.Endpoint{
					Context: context,
					URL: fmt.Sprintf(
						outClusterEurekaUrlFmt,
						s.Namespace,
						s.Name,
						p.Port,
					),
				}
			}
			break
		}
	}

	return res, nil
}

//
func FormatEndpoint(app models.Instance, inCluster bool) *models.Endpoint {
	if app.Port.Enabled == false {
		log.Printf("Insecure port for %s/%s is disabled, skipping..", app.Namespace, app.Name)
		return nil
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
		return nil
	}

	ctx := models.Context{
		Namespace: app.Namespace,
		Name:      app.Name,
	}

	if inCluster {
		return &models.Endpoint{
			Context: ctx,
			URL: fmt.Sprintf(
				inClusterServiceUrlFmt,
				app.IpAddress,
				app.Port.Value,
				metricsUri,
			),
		}
	} else {
		return &models.Endpoint{
			Context: ctx,
			URL: fmt.Sprintf(
				outClusterServiceUrlFmt,
				app.Namespace,
				app.InstanceId,
				app.Port.Value,
				metricsUri,
			),
		}
	}
}
