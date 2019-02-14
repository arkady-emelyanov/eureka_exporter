package models

const InstanceTag = "instance"

type Instance struct {
	Namespace string `xml:"-"`

	Name       string     `xml:"app"`
	IpAddress  string     `xml:"ipAddr"`
	Port       Tag        `xml:"port"`
	SecurePort Tag        `xml:"securePort"`
	Metadata   []Metadata `xml:"metadata"`
	ActionType string     `xml:"actionType"`
	InstanceId string     `xml:"instanceId"`
}

type Metadata struct {
	PrometheusURI string `xml:"prometheusURI"`
}

type Tag struct {
	Value   string `xml:",chardata"`
	Enabled bool   `xml:"enabled,attr"`
}
