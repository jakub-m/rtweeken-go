bin/main: $(shell find . -name \*.go)
	go build -o bin/main main/main.go

