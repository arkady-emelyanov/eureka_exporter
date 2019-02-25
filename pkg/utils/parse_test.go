package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
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
  <versions__delta>1</versions__delta>
  <apps__hashcode>UP_10_</apps__hashcode>
  <application>
    <name>EXAMPLE-MICROSERVICE</name>
    <instance>
      <instanceId>example-microservice-8-rk9lp:example-microservice:8092</instanceId>
      <hostName>10.131.4.51</hostName>
      <app>EXAMPLE-MICROSERVICE</app>
      <ipAddr>10.131.4.51</ipAddr>
      <status>UP</status>
      <overriddenstatus>UNKNOWN</overriddenstatus>
      <port enabled="true">8092</port>
      <securePort enabled="false">443</securePort>
      <countryId>1</countryId>
      <dataCenterInfo class="com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo">
        <name>MyOwn</name>
      </dataCenterInfo>
      <leaseInfo>
        <renewalIntervalInSecs>30</renewalIntervalInSecs>
        <durationInSecs>90</durationInSecs>
        <registrationTimestamp>1551076318613</registrationTimestamp>
        <lastRenewalTimestamp>1551083249449</lastRenewalTimestamp>
        <evictionTimestamp>0</evictionTimestamp>
        <serviceUpTimestamp>1551076318097</serviceUpTimestamp>
      </leaseInfo>
      <metadata>
        <prometheusURI>/actuator/prometheus</prometheusURI>
      </metadata>
      <homePageUrl>http://10.131.4.51:8092/</homePageUrl>
      <statusPageUrl>http://10.131.4.51:8092/info</statusPageUrl>
      <healthCheckUrl>http://10.131.4.51:8092/health</healthCheckUrl>
      <vipAddress>example-microservice</vipAddress>
      <secureVipAddress>example-microservice</secureVipAddress>
      <isCoordinatingDiscoveryServer>false</isCoordinatingDiscoveryServer>
      <lastUpdatedTimestamp>1551076318613</lastUpdatedTimestamp>
      <lastDirtyTimestamp>1551076318042</lastDirtyTimestamp>
      <actionType>ADDED</actionType>
    </instance>
  </application>
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
	require.Equal(t, "example-microservice-8-rk9lp:example-microservice:8092", m[0].InstanceId)
	require.Equal(t, "example-microservice", m[0].Name)
	require.Equal(t, "default", m[0].Namespace)
	require.Equal(t, "10.131.4.51", m[0].IpAddress)
	require.Equal(t, true, m[0].Port.Enabled)
	require.Equal(t, "8092", m[0].Port.Value)
	require.Equal(t, "/actuator/prometheus", m[0].Metadata[0].PrometheusURI)
}
