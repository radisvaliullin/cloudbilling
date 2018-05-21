include build/package/dev_config.env

PROOT=/go/src/github.com/pavlov-tony/xproject
# image name for docker
IMAGE_NAME=xproject
# build args for Dockerfile's
DOCKER_RUN_DEV=@docker run --rm -i -v `pwd`:${PROOT} -v `pwd`/build/.gometalinter.json:/go/src/.gometalinter.json -w ${PROOT}

.PHONY: build-base
build-base:
	@echo "::: building base image"
	@docker build -f build/package/Base.Dockerfile -t $(IMAGE_NAME):base .

#RUN actions
.PHONY: unit-test
unit-test:
	@echo "::: running unit tests"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "go test -v ./..."

.PHONY: lint
lint:
	@echo "::: running code lint"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "gometalinter --line-length=100 ./... --config=/go/src/.gometalinter.json"

.PHONY: cover
cover:
	@echo "::: generating go coverage report"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "go test ./... -coverprofile=../c.out && go tool cover -html=../c.out -o ../coverage.html && cat ../coverage.html" > coverage.html
	@xdg-open coverage.html

.PHONY: run
run:
	@echo "::: running command inside container"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "$(CMD)"

.PHONY: deps
deps:
	@echo "::: installing golang dependencies"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "dep ensure -v -update"

.PHONY: dep-add
dep-add:
	@echo "::: installing package $(PKG)"
	$(DOCKER_RUN_DEV) $(IMAGE_NAME):base sh -c "dep ensure -add $(PKG)"

#COMPOSE actions
.PHONY: up
up:
	@echo "::: building dev environment on port 8080"
	@docker-compose -f `pwd`/deployments/docker-compose.yml up --build

.PHONY: down
down:
	@echo " ::: tear down dev environment"
	@docker-compose -f `pwd`/deployments/docker-compose.yml down

.PHONY: debug
debug:
	@echo "::: debugging dev environment"
	@docker-compose -f `pwd`/deployments/docker-compose-debug.yml up --build

.PHONY: debug-down
debug-down:
	@echo "::: tear down debug dev environment"
	@docker-compose -f `pwd`/deployments/docker-compose-debug.yml down
