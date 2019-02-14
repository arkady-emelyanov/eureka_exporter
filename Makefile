.PHONY: fake-build fake-push fake-local test

fake-build:
	@eval $$(minikube docker-env --shell bash) ;\
	docker build -f examples/fake_exporter/Dockerfile -t localhost:5000/fake_exporter:latest .

	@eval $$(minikube docker-env --shell bash) ;\
	docker build -f examples/fake_eureka/Dockerfile -t localhost:5000/fake_eureka:latest .

fake-apply:
	kubectl apply -f ./examples/fake_exporter/deployment.yml
	kubectl apply -f ./examples/fake_eureka/deployment.yml

fake-delete:
	kubectl delete ns cluster-one
	kubectl delete ns cluster-two
