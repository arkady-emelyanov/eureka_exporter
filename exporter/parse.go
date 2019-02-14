package exporter

import (
	"encoding/xml"
	"io"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

func ParseEurekaResponse(r io.Reader, namespace string) ([]models.Instance, error) {
	var appList []models.Instance

	decoder := xml.NewDecoder(r)
	for {
		tok, err := decoder.Token()
		if tok == nil {
			break
		}

		if err != nil {
			return nil, err
		}

		switch typ := tok.(type) {
		case xml.StartElement:
			if typ.Name.Local == models.InstanceTag {
				var app models.Instance
				if err := decoder.DecodeElement(&app, &typ); err != nil {
					log.Printf("Failed to decode xml element: %v, skipping...", err)
					break
				}

				app.Namespace = namespace
				app.Name = strings.ToLower(app.Name)
				appList = append(appList, app)
			}
		}
	}

	return appList, nil
}

func ParsePromResponse(r io.Reader, namespace, app string) (map[string]*io_prometheus_client.MetricFamily, error) {
	parser := expfmt.TextParser{}
	m, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return nil, err
	}

	for _, v := range m {
		for _, m := range v.Metric {
			// re-label response
			m.Label = append(m.Label, &io_prometheus_client.LabelPair{
				Name:  proto.String("namespace"),
				Value: proto.String(namespace),
			})
			m.Label = append(m.Label, &io_prometheus_client.LabelPair{
				Name:  proto.String("app"),
				Value: proto.String(app),
			})
		}
	}

	return m, nil
}
