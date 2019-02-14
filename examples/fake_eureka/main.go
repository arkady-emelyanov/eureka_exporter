package main

import (
	"encoding/xml"
	"log"
	"net/http"

	"github.com/arkady-emelyanov/eureka_exporter/pkg/kube"
	"github.com/arkady-emelyanov/eureka_exporter/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	eurekaDefaultAddr     = ":8761"
	eurekaFakePodSelector = "app=fake-exporter"
)

type (
	appEntryList []models.Instance
)

func (a appEntryList) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	appListTag := xml.StartElement{Name: xml.Name{Local: "applications"}}
	if err := e.EncodeToken(appListTag); err != nil {
		return err
	}
	for _, app := range a {
		appTag := xml.StartElement{Name: xml.Name{Local: "application"}}
		if err := e.EncodeToken(appTag); err != nil {
			return err
		}
		insTag := xml.StartElement{Name: xml.Name{Local: models.InstanceTag}}
		if err := e.EncodeElement(app, insTag); err != nil {
			return err
		}
		if err := e.EncodeToken(xml.EndElement{Name: appTag.Name}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: appListTag.Name})
}

func main() {
	log.Printf("Listening on: %s\n", eurekaDefaultAddr)
	log.Fatal(http.ListenAndServe(eurekaDefaultAddr, handleRequest()))
}

func handleRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// construct api client
		client, err := kube.GetClient()
		if err != nil {
			log.Println("GetClient error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte{}); err != nil {
				log.Println("Error while serving response:", err)
			}
			return
		}

		// detect current namespace
		namespace := kube.GetNamespace()
		log.Printf("Searching for label=%s in namespace=%s\n", eurekaFakePodSelector, namespace)
		pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: eurekaFakePodSelector,
		})
		if err != nil {
			log.Println("GetClient error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte{}); err != nil {
				log.Println("Error while serving response:", err)
			}
			return
		}

		// generate response structure
		log.Printf("Found %d pods matching label=%s criteria\n", len(pods.Items), eurekaFakePodSelector)
		var root appEntryList
		for _, p := range pods.Items {
			root = append(root, models.Instance{
				InstanceId: p.Name,
				Name:       p.Name,
				Metadata:   []models.Metadata{{PrometheusURI: "/metrics"}},
				ActionType: "ADDED",
				IpAddress:  p.Status.PodIP,
				Port: models.Tag{
					Value:   "8080",
					Enabled: true,
				},
				SecurePort: models.Tag{
					Value:   "8443",
					Enabled: false,
				},
			})
		}

		// encode xml
		log.Println("Generating response...")
		b, err := xml.Marshal(root)
		if err != nil {
			log.Println("Error while marshalling response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte{}); err != nil {
				log.Println("Error while serving response:", err)
			}
			return
		}

		// write final response
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(b); err != nil {
			log.Println("Error while serving response:", err)
		}

		log.Println("Done!")
	})
}
