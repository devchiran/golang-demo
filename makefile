.PHONY: all build

# Load the right config.env file
cnf = buildconfig.local.env

ifeq ($(BUILD_ENV),staging)
	cnf = buildconfig.staging.env
	DEPLOYABLE = true
endif
ifeq ($(BUILD_ENV),prod)
	cnf = buildconfig.prod.env
	DEPLOYABLE = true
endif

include build/$(cnf)
export $(shell sed 's/=.*//' build/$(cnf))

VERSION = $(shell cat build/VERSION)

install:
	@GO111MODULE=on GOPRIVATE=github.com/twitsprout go get -u all
	@GO111MODULE=on GOPRIVATE=github.com/twitsprout go mod tidy
	@GO111MODULE=on go mod vendor

run:
	@go run -mod=vendor -ldflags "-X main.version=$(VERSION)" cmd/golang-demo/main.go

clean:
	@rm -Rf dist

build:
	@go build -mod=vendor -o dist/golang-demo.bin -ldflags "-X main.version=$(VERSION)" cmdgolang-demo/main.go

test:
	@go test -mod=vendor -race -cover ./...

lint:
	@DOCKER_BUILDKIT=1 docker build -f build/Dockerfile.lint -t $(APP_NAME)-lint .
	@docker run -i --rm --name=$(APP_NAME)-lint $(APP_NAME)-lint

local-db:
	@echo --- Spinning down any existing Story Challenges PG db containers
	@make local-db-down
	@echo
	@echo --- Creating a new Story Challenges PG db container
	@make local-db-up
	@echo
	@echo --- Waiting for Story Challenges PG db container to be up
	@sleep 20
	@echo
	@echo --- Running migrations
	@make local-db-migrate
	@echo
	@echo --- Adding seed data
	@make local-db-seed
	@echo
	@echo --- Done

local-db-up:
	@docker compose up -d

local-db-migrate:
	@go run -mod=vendor db/migrate.go -host=localhost:2997 -database=catelog

local-db-seed:
	@docker compose run --rm db sh -c 'psql -h db -U postgres -d catelog_test -f /app/db/seed.sql'

local-db-down:
	@docker compose down

migration-down:
	@migrate \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable \
		-path db/migrations down $(N)

migration-up:
	@migrate \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASS}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable \
		-path db/migrations up $(N)

migration-create:
	@migrate create \
		-seq \
		-ext sql \
		-dir db/migrations \
		$(NAME)

# This is executed by the CI runner and the local-db targets
migrate:
	@go run -mod=vendor db/migrate.go $(filter-out --,$(MAKEFLAGS))

# This is executed by the CI runner
docker-build: check-deployable
	@DOCKER_BUILDKIT=1 docker build \
		-f build/Dockerfile.deploy \
		-t gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(CLOUD_RUN_NAME):$(VERSION) .

# This is executed by the CI runner
docker-push: check-deployable
	@docker push gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(CLOUD_RUN_NAME):$(VERSION)

# This is executed by the CI runner
deploy: check-deployable
	@gcloud run deploy $(CLOUD_RUN_NAME) \
		--quiet \
		--region $(GOOGLE_CLOUD_REGION) \
		--project $(GOOGLE_CLOUD_PROJECT) \
		--platform managed \
		--no-allow-unauthenticated \
		--memory $(CLOUD_RUN_MEMORY_LIMIT) \
		--concurrency $(CLOUD_RUN_CONCURRENCY) \
		--timeout $(CLOUD_RUN_TIMEOUT) \
		--max-instances $(CLOUD_RUN_MAX_INSTANCES) \
		--image gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(CLOUD_RUN_NAME):$(VERSION) \
		--service-account $(CLOUD_RUN_SERVICE_ACCOUNT) \
		--add-cloudsql-instances $(CLOUD_SQL_INSTANCE_NAME) \
		--set-env-vars="PUBSUB_APPROVE_STORY_TOPIC=$(PUBSUB_APPROVE_STORY_TOPIC)" \
		--set-env-vars="PUBSUB_PROJECT_ID=$(PUBSUB_PROJECT_ID)" \
		--set-env-vars ^::^POSTGRES_DB=$(POSTGRES_DB)::POSTGRES_USER=$(POSTGRES_USER)::POSTGRES_PASS=$(POSTGRES_PASS)::POSTGRES_HOST=$(POSTGRES_HOST)::STORY_NETWORK_GATEWAY_URL=$(STORY_NETWORK_GATEWAY_URL)::LOG_LEVEL=$(LOG_LEVEL)::GOOGLE_CLOUD_PROJECT=$(GOOGLE_CLOUD_PROJECT)
	@gcloud run services update-traffic $(CLOUD_RUN_NAME) \
        --to-latest \
        --region $(GOOGLE_CLOUD_REGION) \
        --project=$(GOOGLE_CLOUD_PROJECT) \
        --platform managed

# local pub/sub setup commands
local-pubsub:
	@sudo gcloud beta emulators pubsub start \
		--project=project-local \
		--host-port=localhost:9095

local-pubsub-list-topics:
	@PUBSUB_EMULATOR_HOST=localhost:9095 go run cmd/pubsub/main.go list-topics \
		--project-id=$(GOOGLE_CLOUD_PROJECT)

local-pubsub-create-approve-story-topic:
	@PUBSUB_EMULATOR_HOST=localhost:9095 go run cmd/pubsub/main.go create-topic \
		--project-id=$(GOOGLE_CLOUD_PROJECT) \
		--topic=$(PUBSUB_APPROVE_STORY_TOPIC)

local-pubsub-subscribe-approve-story-topic:
	@PUBSUB_EMULATOR_HOST=localhost:9095 go run cmd/pubsub/main.go subscribe-topic \
		--project-id=$(GOOGLE_CLOUD_PROJECT) \
		--topic=$(PUBSUB_APPROVE_STORY_TOPIC)

# only let commands tagged with this run when you're using a correct BUILD_ENV
check-deployable:
ifndef DEPLOYABLE
	$(error must run with BUILD_ENV = staging or prod)
endif

coverage:
	@go test `go list ./...` -coverprofile=cover.out ./... && \
		go tool cover -func cover.out && \
		rm ./cover.out

coverage-html:
	@go test ./... -coverprofile=cover.out && \
		go tool cover -html=cover.out -o coverage.html && \
		rm ./cover.out && \
		open coverage.html
		