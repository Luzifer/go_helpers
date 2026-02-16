default:

.PHONY: go.work
go.work:
	go work init || true
	find . -name 'go.mod' | cut -d / -f 2 | xargs go work use
