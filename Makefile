install-tools:
	go get github.com/golang/mock/mockgen

generate-mocks:
	ensure generate mocks

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=/tmp/lambgo.coverage && go tool cover -html=/tmp/lambgo.coverage -o=./tests/coverage.html

lint:
	golangci-lint run

generate-toc:
	doctoc README.md
