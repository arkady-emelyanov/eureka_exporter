package utils

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParsePromResponse(t *testing.T) {
	s := `
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 1967
`
	r := strings.NewReader(s)
	m, err := parsePromResponse(r, "example", "default")

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
	<instanceId>fake-exporter-5554b8f746-g6b7s</instanceId>
</instance>
</application>
</applications>
`
	r := strings.NewReader(s)
	m, err := parseEurekaResponse(r, "default")

	require.NoError(t, err)
	require.Len(t, m, 1)

	require.Equal(t, "fake-exporter", m[0].Name)
	require.Equal(t, "default", m[0].Namespace)
	require.Equal(t, "172.17.0.8", m[0].IpAddress)
	require.Equal(t, true, m[0].Port.Enabled)
	require.Equal(t, "8080", m[0].Port.Value)
	require.Equal(t, "/metrics", m[0].Metadata[0].PrometheusURI)
}
