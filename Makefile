.PHONY: test fmt lint check-swagger-ui update-swagger-ui

test:
	@go test -v ./...

fmt:
	@golangci-lint fmt

lint:
	@golangci-lint custom -vv
	@./custom-gcl run --fix ./...

check-swagger-ui:
	@python3 ./scripts/main.py "$(CURDIR)"

update-swagger-ui:
	@python3 ./scripts/update_swagger_ui.py "$(CURDIR)"
