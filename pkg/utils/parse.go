package utils

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"
)

func parseEurekaResponse(r io.Reader, e models.Endpoint) ([]models.Instance, error) {
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
					log.Error().
						Err(err).
						Msg("Error decoding element")
					break
				}

				app.Namespace = e.Namespace
				app.Name = strings.ToLower(app.Name)
				app.InstanceId = strings.ToLower(app.InstanceId)
				appList = append(appList, app)
			}
		}
	}

	return appList, nil
}

func parsePromResponse(r io.Reader, e models.Endpoint) (map[string]*io_prometheus_client.MetricFamily, error) {
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
				Value: proto.String(e.Namespace),
			})
			m.Label = append(m.Label, &io_prometheus_client.LabelPair{
				Name:  proto.String("app"),
				Value: proto.String(e.Name),
			})
			m.Label = append(m.Label, &io_prometheus_client.LabelPair{
				Name:  proto.String("instanceId"),
				Value: proto.String(e.InstanceId),
			})
		}
	}

	return m, nil
}
