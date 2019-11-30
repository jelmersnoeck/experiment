ci: lint report
###########################################
# Linting and testing
##########################################

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

test:
	go test -v ./...

report:
	@echo "" > coverage.txt
	@for d in $$(go list ./...); do \
		go test -v -race -coverprofile=profile.out -covermode=atomic $$d; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done

cover:
	go test -cover ./...
