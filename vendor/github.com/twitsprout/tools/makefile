.PHONY: all

NAME=twitsprout-tools

all:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  deps        - install and update dependencies"
	@echo "  lint        - run the gometalinter"
	@echo "  test        - standard go test"

deps:
	@go mod tidy

lint:
	@DOCKER_BUILDKIT=1 docker build -f build/Dockerfile.lint -t $(NAME)-lint .
	@docker run -i --rm --name=$(NAME)-lint $(NAME)-lint

test-prep: test-clean
	@DOCKER_BUILDKIT=1 docker build -f build/Dockerfile.postgres -t $(NAME)-postgres .
	@docker run -d -p 5432:5432 --name $(NAME)-postgres $(NAME)-postgres
	@sleep 5
	@docker run -d -p 6379:6379 --name $(NAME)-redis redis:5-alpine
	@sleep 5

test-clean:
	@docker stop $(NAME)-postgres > /dev/null || true 
	@docker rm $(NAME)-postgres > /dev/null || true
	@docker stop $(NAME)-redis > /dev/null || true
	@docker rm $(NAME)-redis > /dev/null || true

test:
	@go test -cover -race ./...

coverage-html:
	@go test ./... -coverprofile=profile.cov && \
		go tool cover -html=profile.cov -o coverage.html && \
		rm ./profile.cov && \
		open coverage.html

ci-prep: test-prep
	@echo Waiting for postgres
	@build/wait-for-it.sh 0.0.0.0:5432 -t 30
	@echo Waiting for redis
	@build/wait-for-it.sh 0.0.0.0:6379 -t 30
