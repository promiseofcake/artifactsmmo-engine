build:
	docker build --platform linux/amd64 --build-arg HOME_DIRECTORY=$$HOME -t registry.int.renwickpl.space/promiseofcake/artifactsmmo-engine .
	docker push registry.int.renwickpl.space/promiseofcake/artifactsmmo-engine

test:
	go test ./...
