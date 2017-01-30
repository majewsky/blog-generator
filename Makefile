all: blog-generator

# force people to use golangvend
GOCC := env GOPATH=$(CURDIR)/.gopath go
GOFLAGS := -ldflags '-s -w'

blog-generator: *.go
	$(GOCC) build $(GOFLAGS) -o $@ github.com/majewsky/blog-generator

vendor:
	@golangvend

.PHONY: vendor
