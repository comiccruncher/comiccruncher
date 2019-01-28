# Docker app container name for the Go backend.
DOCKER_APP_CONTAINER_NAME = comiccruncher
# Docker container name for Redis.
DOCKER_REDIS_CONTAINER_NAME = comiccruncher_redis
# Docker container name for Postgres.
DOCKER_PG_CONTAINER_NAME = comiccruncher_postgres
# Command to run Docker.
DOCKER_RUN = docker-compose run --rm ${DOCKER_APP_CONTAINER_NAME}
DOCKER_EXEC = docker-compose exec ${DOCKER_APP_CONTAINER_NAME}
# Command to run docker with the exposed port.
DOCKER_RUN_WITH_PORTS = docker-compose run --service-ports --rm  ${DOCKER_APP_CONTAINER_NAME}
# Settings to cross-compile go binary so that it works on Linux amd64 systems.
DOCKER_RUN_XCOMPILE = docker-compose run -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 --rm ${DOCKER_APP_CONTAINER_NAME}
# The container for tests.
DOCKER_RUN_TEST = docker-compose -f docker-compose.yml -f docker-compose.test.yml run --rm ${DOCKER_APP_CONTAINER_NAME}

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
# The path to the comic bin.
COMIC_BIN = bin/comic

# Locaton of the migrations cmd.
MIGRATIONS_CMD = ./cmd/migrations/migrations.go
# Location of the cerebro cmd.
CEREBRO_CMD = ./cmd/cerebro/cerebro.go
# Location of the web cmd.
WEB_CMD = ./cmd/web/web.go
# Location of comic cmd
COMIC_CMD = ./cmd/comic/comic.go

# The username and location to the api server (that's also the tasks server for now).
LB_SERVER = aimee@142.93.52.234
API_SERVER1 = aimee@68.183.132.127
API_SERVER2 = aimee@198.199.91.173

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
	dep ensure

.PHONY: dep-ensure-update
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

# Test the files with any race conditions (unfortunately Alpine-based images don't work w/ race command...so
# use this command locally :(
.PHONY: test
test:
	go test -race -v $(shell go list ./... | grep -v ./cmd) -coverprofile=coverage.txt

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
	 go build -o ./build/deploy/api/bin ${WEB_CMD}

docker-docker-build-webapp:
	docker build ./build/deploy/api -t comiccruncher/api:latest

docker-docker-push-webapp:
	docker push comiccruncher/api:latest

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

# Runs the program for generating thumbnails for characters.
.PHONY: import-characters
generate-thumbs:
	 ${GO_RUN_LOCAL} ${COMIC_CMD} generate thumbs ${EXTRA_FLAGS}

.PHONY: docker-import-characters
docker-import-characters:
	${DOCKER_RUN} go run ${CEREBRO_CMD} import characters

.PHONY: docker-generate-thumbs
docker-generate-thumbs:
	${DOCKER_RUN} go run ${COMIC_CMD} generate thumbs ${EXTRA_FLAGS}

.PHONY: queue-characters
queue-characters:
	go run -race cmd/queuecharacters.go

# Generate mocks for testing.
.PHONY: mockgen
mockgen:
	mockgen -destination=internal/mocks/comic/sync.go -source=comic/sync.go
	mockgen -destination=internal/mocks/comic/repositories.go -source=comic/repositories.go
	mockgen -destination=internal/mocks/comic/services.go -source=comic/services.go
	mockgen -destination=internal/mocks/comic/cache.go -source=comic/cache.go
	mockgen -destination=internal/mocks/cerebro/characterissue.go -source=cerebro/characterissue.go
	mockgen -destination=internal/mocks/search/service.go -source=search/service.go
	mockgen -destination=internal/mocks/storage/s3.go -source=storage/s3.go
	mockgen -destination=internal/mocks/cerebro/utils.go -source=cerebro/utils.go
	mockgen -destination=internal/mocks/imaging/thumbnail.go -source=imaging/thumbnail.go
	mockgen -destination=internal/mocks/auth/auth.go -source=auth/auth.go

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

# Builds the comic commands.
.PHONY: build-comic
build-comic:
	go build -o ./bin/comic -v ${COMIC_CMD}

# Builds the comic commands inside the Docker container.
docker-build-comic-xcompile:
	${DOCKER_RUN_XCOMPILE} make build-comic

# Builds all the app binaries in the Docker contaner.
.PHONY: docker-build-xcompile
docker-build-xcompile: docker-build-migrations-xcompile docker-build-cerebro-xcompile docker-build-webapp-xcompile docker-build-comic-xcompile

# Uploads the cerebro binary to the remote server. Used for CircleCI.
.PHONY: remote-upload-cerebro
remote-upload-cerebro:
	scp ./${CEREBRO_BIN} ${LB_SERVER}:~/bin

# Uploads the cerebro binary to the remote server. Used for CircleCI.
.PHONY: remote-deploy-cerebro
remote-deploy-cerebro: remote-upload-cerebro

# Uploads the comic binary to the remote server.
.PHONY: remote-deploy-comic
remote-deploy-comic:
	scp ./${COMIC_BIN} ${LB_SERVER}:~/bin

# Uploads the migrations binary to the remote server. Used for CircleCI.
.PHONY: remote-upload-migrations
remote-upload-migrations:
	scp ./${MIGRATIONS_BIN} ${LB_SERVER}:~/bin

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
	scp -r ./build/deploy/nginx/* ${LB_SERVER}:~/

# Uploads the reload.sh script.
.PHONY: remote-upload-reload-script
remote-upload-reload-script:
	scp -r ./build/deploy/nginx/reload.sh ${LB_SERVER}:~/reload.sh

.PHONY: remote-deploy-nginx-initial
remote-deploy-nginx-initial: remote-upload-nginx
	ssh ${LB_SERVER} "sh deploy.sh"

# Reloads nginx.
.PHONY: docker-reload-nginx
remote-reload-nginx:
	ssh ${LB_SERVER} "sh reload.sh"

# Uploads script and restarts nginx on server.
.PHONY: remote-deploy-nginx
remote-deploy-nginx: remote-upload-reload-script remote-reload-nginx

remote-deploy-api:
	scp ./build/deploy/api/* ${API_SERVER}:~/
	ssh ${API_SERVER} "sh deploy.sh"

remote-deploy-api1:
	API_SERVER=${API_SERVER1} make remote-deploy-api

remote-deploy-api2:
	API_SERVER=${API_SERVER2} make remote-deploy-api

remote-deploy-lb: remote-upload-nginx remote-reload-nginx

remote-deploy-webapps: remote-deploy-api1 remote-deploy-api2 remote-deploy-lb
