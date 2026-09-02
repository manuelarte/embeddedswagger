.PHONY: test fmt lint check-swagger-ui

test:
	@go test -v ./...

fmt:
	@golangci-lint fmt

lint:
	@golangci-lint custom -vv
	@./custom-gcl run --fix ./...

check-swagger-ui:
	@./scripts/check-swagger-ui.sh
