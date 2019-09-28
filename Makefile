NAME_APP = samsclub

PATH_APP = $(shell pwd)/cmd/$(NAME_APP)
PATH_BIN = $(shell pwd)/bin

build:
	CGO_ENABLED=0 go build -v -o $(PATH_BIN)/$(NAME_APP) $(PATH_APP)

run:
	$(PATH_BIN)/$(NAME_APP)
