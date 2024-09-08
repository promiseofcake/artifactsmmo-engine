build:
	docker build --platform linux/amd64 -t $(HOME_DOCKER_REGISTRY)/promiseofcake/artifactsmmo-engine .


deploy: build
	docker push $(HOME_DOCKER_REGISTRY)/promiseofcake/artifactsmmo-engine
	curl -XPOST $(ARTIFACTS_WEBHOOK)

test:
	go test ./...
