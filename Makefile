.PHONY: run gen-logs fmt tidy test

run:
	go run ./cmd

gen-logs:
	go run task.go

fmt:
	gofmt -w .

tidy:
	go mod tidy

test:
	go test ./...