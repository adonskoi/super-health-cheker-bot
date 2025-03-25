build:
	go build -o bot ./app

lint:
	gofmt -w .
run:
	go run ./app 

test:
	go test -v ./...

build_x64:
	env GOOS=linux GOARCH=amd64 go build -o bot64 ./app
