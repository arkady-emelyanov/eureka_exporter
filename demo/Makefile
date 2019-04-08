
.PHONY: build-all
build-all:
	cd auto-service && docker build . -t 0xfff/demo_auto_service:latest
	cd cloud-config-service && docker build . -t 0xfff/demo_cloud_config_service:latest
	cd eureka-service && docker build . -t 0xfff/demo_eureka_service:latest
	cd moto-service && docker build . -t 0xfff/demo_moto_service:latest

.PHONY: publish-all
publish-all:
	docker push 0xfff/demo_auto_service:latest
	docker push 0xfff/demo_cloud_config_service:latest
	docker push 0xfff/demo_eureka_service:latest
	docker push 0xfff/demo_moto_service:latest
