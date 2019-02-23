package utils

import (
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePromResponse(t *testing.T) {
	s := `
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 1967
`
	r := strings.NewReader(s)
	e := models.Endpoint{
		Context: models.Context{
			Namespace: "default",
			Name: "example",
			InstanceId: "123-321-123",
		},
	}

	m, err := parsePromResponse(r, e)
	require.NoError(t, err)
	require.Len(t, m, 1)
}

func TestParseEurekaResponse(t *testing.T) {
	s := `<applications>
<application>
<instance>
	<app>fake-exporter</app>
	<ipAddr>172.17.0.8</ipAddr>
	<port enabled="true">8080</port>
	<securePort enabled="false">8443</securePort>
	<metadata>
		<prometheusURI>/metrics</prometheusURI>
	</metadata>
	<actionType>ADDED</actionType>
	<instanceId>fake-exporter-5554b8f746-g6b7s:34583714-2576-4238-a4a9-9bb95e568033</instanceId>
</instance>
</application>
</applications>
`
	r := strings.NewReader(s)
	e := models.Endpoint{
		Context: models.Context{
			Namespace: "default",
		},
	}

	m, err := parseEurekaResponse(r, e)
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, "fake-exporter-5554b8f746-g6b7s:34583714-2576-4238-a4a9-9bb95e568033", m[0].InstanceId)
	require.Equal(t, "fake-exporter", m[0].Name)
	require.Equal(t, "default", m[0].Namespace)
	require.Equal(t, "172.17.0.8", m[0].IpAddress)
	require.Equal(t, true, m[0].Port.Enabled)
	require.Equal(t, "8080", m[0].Port.Value)
	require.Equal(t, "/metrics", m[0].Metadata[0].PrometheusURI)
}
