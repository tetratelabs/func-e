deps:
	go mod download

build: deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/getenvoy ./cmd/getenvoy/main.go

docker: build
	docker build -t liamwhite/getenvoy:latest .
	# docker push liamwhite/getenvoy:latest
