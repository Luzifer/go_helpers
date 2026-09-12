default:

.PHONY: check-tidy
check-tidy:
	bash ci/check-tidy.sh

.PHONY: go.work
go.work:
	go work init || true
	find . -name 'go.mod' | xargs dirname | xargs go work use

test: go.work
	go test -v -cover $(shell go list -f '{{.Dir}}/...' -m)
