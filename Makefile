FLAGS?=-v
SERVICES?=services
IMAGES?=images
SERVICE?=auth

services := $(notdir $(shell find ./$(SERVICES)/ -mindepth 1 -maxdepth 1 -type d))

default: build

.SILENT:

build:
	if [ "$(SERVICE)" = "" ]; then \
		go build -o ./bin/$(SERVICES) $(FLAGS) ./$(SERVICES) ;\
	else \
  		go build -o ./bin/$(SERVICES)/$(SERVICE) $(FLAGS) ./$(SERVICES)/$(SERVICE) ;\
	fi

build_all:
	$(foreach dir,$(wildcard SERVICES/*), go build $(FLAGS) ./$(dir);)

build_image:
	if [ "$(SERVICE)" = "" ]; then \
		docker build -f ./$(SERVICES)/Dockerfile -t async_arch_course/$(IMAGES) . ;\
	else \
	  	docker build -f ./$(SERVICES)/$(SERVICE)/Dockerfile -t async_arch_course/$(IMAGES)/$(SERVICE) . ;\
	fi

build_image_all:
	for service in $(services) ; do \
  		docker build -f ./$(SERVICES)/$$service/Dockerfile -t async_arch_course/$(IMAGES)/$$service . ;\
  	done

run:
	docker-compose build && docker-compose up -d

stop:
	docker-compose down

lint:
	golangci-lint run -v ./...

tidy:
	go mod tidy

.PHONY: build build_all build_image build_image_all run stop lint tidy
