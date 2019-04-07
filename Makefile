TRAVIS_TAG ?= latest
demo := demo

.PHONY: build-all
build-all:
	docker build -f Dockerfile . -t 0xfff/eureka_exporter:$(TRAVIS_TAG)
	for d in $(demo); \
	do \
		$(MAKE) --directory=$$d; \
	done

.PHONY: publish-all
publish-all:
	docker push 0xfff/eureka_exporter:$(TRAVIS_TAG)
	for d in $(demo); \
	do \
		$(MAKE) --directory=$$d; \
	done

## Build and deploy exporter to Minikube cluster
.PHONY: minikube
minikube: mini-build mini-apply

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
