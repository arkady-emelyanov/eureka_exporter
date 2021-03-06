package models

type Context struct {
	InstanceId string
	Name       string
	Namespace  string
}

type Endpoint struct {
	Context
	URL string
}

func (e Endpoint) IsEmpty() bool {
	return e.URL == ""
}
