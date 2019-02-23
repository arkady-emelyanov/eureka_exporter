.PHONY: minikube
minikube: fake-build fake-apply mini-build mini-apply

.PHONY: mini-build
mini-build:
	@eval $$(minikube docker-env --shell bash) ;\
	docker build -f Dockerfile -t localhost:5000/eureka_exporter:latest .

.PHONY: mini-apply
mini-apply:
	kubectl apply -f ./deployment.yml

.PHONY: mini-delete
mini-delete:
	kubectl delete ns monitoring
	kubectl delete clusterrolebinding eureka-exporter-rolebinding
	kubectl delete clusterrole eureka-exporter-role

.PHONY: fake-build
fake-build:
	@eval $$(minikube docker-env --shell bash) ;\
	docker build -f examples/fake_exporter/Dockerfile -t localhost:5000/fake_exporter:latest .

	@eval $$(minikube docker-env --shell bash) ;\
	docker build -f examples/fake_eureka/Dockerfile -t localhost:5000/fake_eureka:latest .

.PHONY: fake-apply
fake-apply:
	kubectl apply -f ./examples/fake_exporter/deployment.yml
	kubectl apply -f ./examples/fake_eureka/deployment.yml

.PHONY: fake-delete
fake-delete:
	kubectl delete ns cluster-one
	kubectl delete ns cluster-two
