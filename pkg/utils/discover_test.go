package utils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

func TestFormatEndpoint(t *testing.T) {
	app := models.Instance{
		Namespace:  "example-ns",
		Name:       "example-microservice",
		IpAddress:  "10.152.11.25",
		InstanceId: "example-microservice-1231-321-223",
		Port:       models.Tag{Value: "9091", Enabled: true},
		SecurePort: models.Tag{Value: "8491", Enabled: false},
		Metadata: []models.Metadata{
			{
				PrometheusURI: "/actuator/prometheus",
			},
		},
	}

	expInCluster := models.Endpoint{
		Context: models.Context{
			InstanceId: "example-microservice-1231-321-223",
			Name:       "example-microservice",
			Namespace:  "example-ns",
		},
		URL: "http://10.152.11.25:9091/actuator/prometheus",
	}
	resInCluster := FormatEndpoint(app, true)
	require.Equal(t, expInCluster, *resInCluster)

	// out of cluster a little bit tricky, it assumes that
	// instanceId == pod, behaviour of fake_eureka.
	// anyway, outOfCluster mode is for development/testing purposes only
	// and not for production usage.
	expOutCluster := models.Endpoint{
		Context: models.Context{
			InstanceId: "example-microservice-1231-321-223",
			Name:       "example-microservice",
			Namespace:  "example-ns",
		},
		URL: "http://localhost:8001/api/v1/namespaces/example-ns/pods/example-microservice-1231-321-223:9091/proxy/actuator/prometheus",
	}
	resOutCluster := FormatEndpoint(app, false)
	require.Equal(t, expOutCluster, *resOutCluster)
}
