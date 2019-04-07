TRAVIS_TAG ?= latest
demo := demo

.PHONY: build-all
build-all:
	docker build -f Dockerfile . -t 0xfff/eureka_exporter:$(TRAVIS_TAG)
	for d in $(demo); \
	do \
		$(MAKE) build-all --directory=$$d; \
	done

.PHONY: publish-all
publish-all:
	docker push 0xfff/eureka_exporter:$(TRAVIS_TAG)
	for d in $(demo); \
	do \
		$(MAKE) publish-all --directory=$$d; \
	done
