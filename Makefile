build:
	go build -o bot ./app

run:
	go run ./app 

test:
	go test -v ./...