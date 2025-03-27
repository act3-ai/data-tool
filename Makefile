.PHONY: clean
clean:
	- rm -rf bin/ace-dt*

.PHONY: template
template:
	- rm -rf internal/mirror/testing/testdata/large/oci
	go run ./cmd/ace-dt run-recipe internal/mirror/testing/testdata/large/recipe.jsonl

	- rm -rf internal/mirror/testing/testdata/small/oci
	go run ./cmd/ace-dt run-recipe internal/mirror/testing/testdata/small/recipe.jsonl

.PHONY: cover
cover:
	go clean -testcache
	- rm coverage.txt
	go test -count=1 ./... -coverprofile coverage.txt -coverpkg=$(shell go list )/...
	./filter-coverage.sh < coverage.txt > coverage.txt.filtered
	go tool cover -func coverage.txt.filtered

# bench is the only test suite duplicated with dagger, as running within a container may
# not be as effective
.PHONY: bench
bench:
	go test -benchmem -run=^$$ -bench=. ./...
