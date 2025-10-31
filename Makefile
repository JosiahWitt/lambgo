install-tools:
	go install github.com/JosiahWitt/ensure/cmd/ensure@v0.2.0

generate-mocks:
	ensure mocks generate

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=/tmp/lambgo.coverage && go tool cover -html=/tmp/lambgo.coverage -o=./tests/coverage.html

lint:
	staticcheck ./...
	golangci-lint run
