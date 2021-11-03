NAME := JackAnalyser
SOURCES := $(shell find -name \*.go) go.mod

$(NAME): $(SOURCES)
	go build -o $@

.PHONY: clean
clean:
	go clean