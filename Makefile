
.DEFAULT_GOAL := test

# Get version from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: init
init:
	go mod download
	go tool templ generate
	op inject -f -i envrc.template -o .envrc
	npm install
	npx @tailwindcss/cli -i ./ui/static/css/input.css -o ./ui/static/css/output.css --minify
	direnv allow .

.PHONY: secrets
secrets:
	op inject -f -i envrc.template -o .envrc
	direnv allow .

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run GO commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: build/web
build/web:
	go build $(LDFLAGS) -o bin/server ./cmd/web/

.PHONY: build/cli
build/cli:
	go build $(LDFLAGS) -o bin/gocost ./cmd/cli

.PHONY: templ
templ:
	go tool templ generate
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run development commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

# run templ generation in watch mode to detect all .templ files and
# re-create _templ.txt files on change, then send reload event to browser.
.PHONY: dev/templ
dev/templ:
	go tool templ generate --watch \
	--proxy="http://localhost:4000" \
	--proxybind="localhost" \
	--proxyport="4001" \
	--open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
.PHONY: dev/server
dev/server:
	go tool air \
	--build.cmd "go build $(LDFLAGS) -o ./tmp/bin/ ./cmd/web/" --build.bin "tmp/bin/web" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# watch for css or templ change in the ui, then reload the browser via templ proxy.
.PHONY: dev/sync_assets
dev/sync_assets:
	go run github.com/air-verse/air@latest \
	--build.cmd "templ generate --notify-proxy --proxybind='localhost' --proxyport='4001'" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.exclude_ext "templ" \
	--build.include_dir "ui" \
	--build.include_ext "css" \
	--build.follow_symlink true > /dev/null 2>&1 &

# run tailwindcss to generate the styles.css bundle in watch mode.
.PHONY: dev/tailwind
dev/tailwind:
	npx @tailwindcss/cli -i ./ui/static/css/input.css -o ./ui/static/css/output.css --minify --watch # > /dev/null 2>&1 &

# start all 4 watch processes in parallel.
.PHONY: dev
dev:
	make -j4 dev/templ dev/server dev/tailwind dev/sync_assets

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run test commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

# run go tests
.PHONY: test
test:
	go test ./internal... 

# check for data race conditions
.PHONY: test/race
test/race:
	go test -race ./internal...

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run code format and code style commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

# run go vet tool
.PHONY: vet
vet:
	go vet ./internal...

# run staticcheck tool
.PHONY: staticcheck
staticcheck:
	staticcheck ./internal...

# run all tools
.PHONY: check
check: vet staticcheck

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run docker commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: docker/deploy
docker/deploy:
	docker stack deploy -c compose.yml gocost --detach=true --with-registry-auth

.PHONY: docker/build
docker/build:
	docker build . -t gocost:latest --build-arg VERSION=$(VERSION)

.PHONY: docker/run
docker/run:
	docker run -it --rm --name gocost \
							-e ALLOWED_HOSTS=localhost \
							-e DOMAIN=localhost \
							-v gocost_data:/app/data \
              -p 4000:4000 \
              gocost:latest

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run release commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: release/patch
release/patch:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v0.0.1"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
    		PATCH=$$(echo $$LATEST_TAG | cut -d. -f3); \
    		NEW_PATCH=$$((PATCH + 1)); \
    		NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"


.PHONY: release/minor
release/minor:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v0.1.0"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
    		NEW_MINOR=$$((MINOR + 1)); \
    		NEW_TAG="v$$MAJOR.$$NEW_MINOR.0"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"

.PHONY: release/major
release/major:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v1.0.0"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		NEW_MAJOR=$$((MAJOR + 1)); \
    		NEW_TAG="v$$NEW_MAJOR.0.0"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"
