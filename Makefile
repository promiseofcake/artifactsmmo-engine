build:
	docker build --build-arg HOME_DIRECTORY=$$HOME -t promiseofcake/artifactsmmo-engine .

test:
	go test ./...