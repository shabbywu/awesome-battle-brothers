##@ Scripts
.PHONY: build
build:  ## build image
	@docker build --force-rm -f rootfs\Dockerfile -t online-championships .


.PHONY: deploy
deploy:  ## deploy with docker compose
	@docker compose -f deploy/docker-compose.yaml up -d

##@ Misc
.PHONY: help
help:  ## Show this message
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
