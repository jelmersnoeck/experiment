ci: lint test

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

cover:
	go test -cover ./...
