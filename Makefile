ci: lint report

###########################################
# Bootstrapping
##########################################
BOOTSTRAP= \
	   github.com/alecthomas/gometalinter

$(BOOTSTRAP):
	go get -u $@

bootstrap: $(BOOTSTRAP)
	gometalinter --install

###########################################
# Linting and testing
##########################################
LINTERS=\
		gofmt \
		golint \
		gosimple \
		vet \
		misspell \
		ineffassign \
		deadcode

$(LINTERS):
	gometalinter --tests --disable-all --vendor --deadline=5m -s data ./... --enable $@

lint: $(LINTERS)

test:
	go test -v ./...

report:
	@echo "" > coverage.txt
	@for d in $$(go list ./...); do \
		go test -race -coverprofile=profile.out -covermode=atomic $$d; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done

cover:
	go test -cover ./...
