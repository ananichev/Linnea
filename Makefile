run:
	go build -o bin/main cmd/main.go
	./bin/main

debug:
	go build -o bin/main cmd/main.go
	./bin/main -debug

build:
	go build -o bin/main cmd/main.go

docker:
	docker build -t joachimflottorp/linnea .