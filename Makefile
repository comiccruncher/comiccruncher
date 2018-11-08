# Docker app container name for the Go backend.
DOCKER_APP_CONTAINER_NAME = comiccruncher
# Docker container name for Redis.
DOCKER_REDIS_CONTAINER_NAME = comiccruncher_redis
# Docker container name for Postgres.
DOCKER_PG_CONTAINER_NAME = comiccruncher_postgres
# Command to run Docker.
DOCKER_RUN = docker-compose run ${DOCKER_APP_CONTAINER_NAME}
DOCKER_EXEC = docker-compose exec ${DOCKER_APP_CONTAINER_NAME}
# Command to run docker with the exposed port.
DOCKER_RUN_WITH_PORTS = docker-compose run --service-ports ${DOCKER_APP_CONTAINER_NAME}
# Settings to cross-compile go binary so that it works on Linux amd64 systems.
DOCKER_RUN_XCOMPILE = docker-compose run -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 ${DOCKER_APP_CONTAINER_NAME}
# The container for tests.
DOCKER_RUN_TEST = docker-compose -f docker-compose.yml -f docker-compose.test.yml run ${DOCKER_APP_CONTAINER_NAME}

# Command to go run locally.
GO_RUN_LOCAL = GORACE="log_path=./" go run -race

# The path to the migrations bin.
MIGRATIONS_BIN = bin/migrations
# The path to the webapp bin.
WEBAPP_BIN = bin/webapp
# The path to the temp webapp bin.
WEBAPP_TMP_BIN = bin/webapp1
# The path to the cerebro bin.
CEREBRO_BIN = bin/cerebro

# Locaton of the migrations cmd.
MIGRATIONS_CMD = ./cmd/migrations/migrations.go
# Location of the cerebro cmd.
CEREBRO_CMD = ./cmd/cerebro/cerebro.go
# Location of the web cmd.
WEB_CMD = ./cmd/web/web.go

# The username and location to the api server (that's also the tasks server for now).
LB_SERVER = root@142.93.52.234
API_SERVER1 = aimee@68.183.132.127
API_SERVER2 = aimee@142.93.121.60

# Creates a .netrc file for access to private Github repository for cerebro.
.PHONY: netrc
netrc:
	rm -rf .netrc && echo "machine github.com\nlogin $(GITHUB_ACCESS_TOKEN)" > .netrc && chmod 600 .netrc

# Build the docker container.
.PHONY: up
docker-up:
	docker-compose up -d --build --force-recreate --remove-orphans

# stop the docker containers.
.PHONY: docker-stop
docker-stop:
	docker-compose stop

# Show the docker logs from the services.
.PHONY: docker-logs
docker-logs:
	docker-compose -f docker-compose.yml logs --tail="100" -f postgres redis comiccruncher

# Run the migrations for the test db.
.PHONY: docker-migrations-test
docker-migrations-test:
	${DOCKER_RUN_TEST} go run ${MIGRATIONS_CMD}

# Create the containers for testing.
.PHONY: docker-up-test
docker-up-test:
	docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d --build

# Remove the test containers.
.PHONY: docker-rm-test
docker-rm-test:
	docker-compose -f docker-compose.test.yml rm

# Stop the test containers.
.PHONY: docker-stop-test
docker-stop-test:
	docker-compose -f docker-compose.test.yml stop

# Run the go tests in the docker container.
.PHONY: docker-test
docker-test:
	${DOCKER_RUN_TEST} go test -v $(shell ${DOCKER_RUN_TEST} go list ./... | grep -v ./cmd) -coverprofile=coverage.txt

# Just run the database tests.
.PHONY: docker-test-db
docker-test-db:
	${DOCKER_RUN_TEST} go test -v github.com/aimeelaplant/comiccruncher/comic/ github.com/aimeelaplant/comiccruncher/search/

# Install the docker images and Go dependencies.
.PHONY: docker-install
docker-install: docker-up docker-dep-ensure

# Install the Go dependencies.
.PHONY: dep-ensure
dep-ensure:
	dep ensure -update

# Install the Go dependencies in the Docker container.
.PHONY: docker-dep-ensure
docker-dep-ensure:
	${DOCKER_RUN} make dep-ensure

# Format the files with `go fmt`.
.PHONY: docker-format
docker-format:
	${DOCKER_RUN} make format

# Format the files
.PHONY: format
format:
	go fmt $(shell go list ./...)

# Vet the files in the Docker container.
.PHONY: docker-vet
docker-vet:
	${DOCKER_RUN} make vet

# Vet the files.
.PHONY: vet
vet:
	go vet $(shell go list ./...)

# Lint the go files.
.PHONY: lint
lint:
	golint $(shell go list ./...)

# Lint the files in the Docker container.
# Not sure why I have to specify `/gocode/bin/golint` and not just `golint`?!?!
.PHONY: docker-lint
docker-lint:
	${DOCKER_RUN} /gocode/bin/golint $(shell go list ./...)

# Reports any cyclomatic complexilities over 15. For goreportcard.
.PHONY: cyclo
cyclo:
	gocyclo -over 15 $(shell ls -d */ | grep -v vendor | awk '{print $$$11}')

# Reports any ineffectual if assignments. For goreportcard.
.PHONY: ineffassign
ineffassign:
	ineffassign .

# Reports any misspellings. For goreportcard.
.PHONY: misspell
misspell:
	misspell $(shell go list ./...)

# Generate any errors for go report card.
.PHONY: reportcard
reportcard: ineffassign misspell lint vet cyclo

# Run the Docker redis-cli.
.PHONY: redis-cli
docker-redis-cli:
	docker exec -it ${DOCKER_REDIS_CONTAINER_NAME} redis-cli -p 6380 -a foo

# Flush the redis cache.
.PHONY: redis-flush
docker-redis-flush:
	docker exec -it ${DOCKER_REDIS_CONTAINER_NAME} redis-cli -p 6380 -a foo flushall

# Build the web application in the Docker container with cross compilation settings so it works on linux amd64 systems.
.PHONY: docker-build-webapp-xcompile
docker-build-webapp-xcompile:
	${DOCKER_RUN_XCOMPILE} make build-webapp

# Builds the webapp binary.
.PHONY: build-webapp
build-webapp:
	 go build -o ./bin/webapp ${WEB_CMD}

# Run the web application.
.PHONY: web
web:
	go run -race ${WEB_CMD} start -p 8001

# Run the web application in Docker container.
.PHONY: docker-web
docker-web:
	${DOCKER_RUN_WITH_PORTS} make web

# Docker run the migrations for the development database.
.PHONY: docker-migrations
docker-migrations:
	${DOCKER_RUN} go run ${MIGRATIONS_CMD}

# Run the migrations for the development database.
.PHONY: migrations
migrations:
	${GO_RUN_LOCAL} ${MIGRATIONS_CMD}

.PHONY: import-characterissues
import-characterissues:
	${GO_RUN_LOCAL} ${CEREBRO_CMD} import characterissues ${EXTRA_FLAGS}

.PHONY: import-charactersources
import-charactersources:
	${GO_RUN_LOCAL} ${CEREBRO_CMD} import charactersources ${EXTRA_FLAGS}

.PHONY: import-charactersources
start-characterissues:
	${GO_RUN_LOCAL} ${CEREBRO_CMD} start characterissues ${EXTRA_FLAGS}

# Runs the program for creating characters from the Marvel API.
.PHONY: import-characters
import-characters:
	 ${GO_RUN_LOCAL} ${CEREBRO_CMD} import characters ${EXTRA_FLAGS}

.PHONY: enqueue-characters
enqueue-characters:
	${GO_RUN_LOCAL} cmd/enqueue.go characters ${EXTRA_FLAGS}

.PHONY: docker-import-characters
docker-import-characters:
	${DOCKER_RUN} go run -race ${CEREBRO_CMD} import characters

# Builds the binary for sending characters to the sync queue.
.PHONY: build-queuecharacters
build-enqueue:
	go build -o bin/enqueue -v ./cmd/messaging/enqueue.go

.PHONY: build-enqueue-xcompile
build-enqueue-xcompile:
	${DOCKER_RUN_XCOMPILE} make build-enqueue

# Runs the program to send characters to the sync queue.
.PHONY: queue-characters
docker-queue-characters:
	${DOCKER_RUN} go run cmd/queuecharacters.go

.PHONY: queue-characters
queue-characters:
	go run -race cmd/queuecharacters.go

# Generate mocks for testing.
.PHONY: mockgen
mockgen:
	mockgen -destination=internal/mocks/comic/repositories.go -source=comic/repositories.go
	mockgen -destination=internal/mocks/comic/services.go -source=comic/services.go
	mockgen -destination=internal/mocks/comic/cache.go -source=comic/cache.go
	mockgen -destination=internal/mocks/messaging/messenger.go -source=messaging/messenger.go
	mockgen -destination=internal/mocks/cerebro/characterissue.go -source=cerebro/characterissue.go
	mockgen -destination=internal/mocks/search/service.go -source=search/service.go

# Generate mocks for testing.
docker-mockgen:
	${DOCKER_RUN} make mockgen

# Builds the migrations binary.
.PHONY: build-migrations
build-migrations:
	go build -o ./bin/migrations -v ${MIGRATIONS_CMD}

# Builds the migrations binary inside the Docker container.
.PHONY: docker-build-migrations-xcompile
docker-build-migrations-xcompile:
	${DOCKER_RUN_XCOMPILE} make build-migrations

# Builds the cerebro binary.
.PHONY: build-cerebro
build-cerebro:
	go build -o ./bin/cerebro -v ${CEREBRO_CMD}

# Builds the cerebro binary inside the Docker container.
.PHONY: docker-build-cerebro-xcompile
docker-build-cerebro-xcompile:
	${DOCKER_RUN_XCOMPILE} make build-cerebro

# Builds all the app binaries in the Docker contaner.
.PHONY: docker-build-xcompile
docker-build-xcompile: docker-build-migrations-xcompile docker-build-cerebro-xcompile docker-build-webapp-xcompile

# Uploads the cerebro binary to the remote server. Used for CircleCI.
.PHONY: remote-upload-cerebro
remote-upload-cerebro:
	scp ./${CEREBRO_BIN} ${LB_SERVER}:/usr/local/bin

# Uploads the cerebro binary to the remote server. Used for CircleCI.
.PHONY: remote-deploy-cerebro
remote-deploy-cerebro: remote-upload-cerebro

# Uploads the migrations binary to the remote server. Used for CircleCI.
.PHONY: remote-upload-migrations
remote-upload-migrations:
	scp ./${MIGRATIONS_BIN} ${LB_SERVER}:/usr/local/bin

# Runs migrations over the remote server. Used for CircleCI.
.PHONY: remote-run-migrations
remote-run-migrations:
	ssh ${LB_SERVER} "bash -s" < ./build/migrations.sh

# Uploads and runs migrations over the server. Used for CircleCI.
.PHONY: remote-deploy-migrations
remote-deploy-migrations: remote-upload-migrations remote-run-migrations

# Uploads nginx config.
.PHONY: remote-upload-nginx
remote-upload-nginx:
	scp ./build/deploy/nginx/nginx.conf ${LB_SERVER}:/etc/nginx/nginx.conf

# Restarts nginx on server.
.PHONY: remote-deploy-nginx
remote-deploy-nginx: remote-upload-nginx
	ssh ${LB_SERVER} "systemctl restart nginx"

remote-deploy-api1:
	scp ./${WEBAPP_BIN} ${API_SERVER1}:~/${WEBAPP_TMP_BIN}
	ssh ${API_SERVER1} "bash -s" < ./build/webapp.sh

remote-deploy-api2:
	scp ./${WEBAPP_BIN} ${API_SERVER2}:~/${WEBAPP_TMP_BIN}
	ssh ${API_SERVER2} "bash -s" < ./build/webapp.sh

remote-deploy-lb: remote-upload-nginx
	ssh ${LB_SERVER} "nginx -s reload"

remote-deploy-webapps:
	make -j remote-deploy-api1 remote-deploy-api2; make remote-deploy-lb
