# Replace with your output binary name
BIN = conex

GO = go
DIST_FLAGS = -ldflags "-s -w"
DIST_FOLDER = dist
GO_ENV = CGO_ENABLED=0

SQLC = github.com/sqlc-dev/sqlc/cmd/sqlc
TEMPL = github.com/a-h/templ/cmd/templ
AIR = github.com/air-verse/air@v1.63.0

NPM = npm
NPX = npx

LOGFILE = build.log
# LOG ?= >> $(LOGFILE) 2>&1
LIVELOGFILE = live.log
# LIVELOG ?= >> $(LIVELOGFILE) 2>&1
DISTLOGFILE = dist.log
# DISTLOG ?= >> $(DISTLOGFILE) 2>&1
GEN =

.PHONY: all
all: build

SQL=$(wildcard database/*.sql)
GEN += internal/db
internal/db: $(SQL)
	$(GO) tool $(SQLC) generate $(LOG)
	@touch $@

.PHONY: clean/sql
clean/sql:
	rm -rf internal/db

TEMPLATES := $(wildcard templates/*.templ)
TEMPLATES_GEN := $(TEMPLATES:.templ=_templ.go)
GEN += $(TEMPLATES_GEN)
$(TEMPLATES_GEN) &: $(TEMPLATES)
	$(GO) tool $(TEMPL) generate -v $(LOG)
	@touch $@

.PHONY: clean/templates
clean/templates:
	rm -rf $(TEMPLATES_GEN)

node_modules: package.json package-lock.json
	$(NPM) ci $(LOG)

.PHONY: clean/node_modules
clean/node_modules:
	rm -rf node_modules

ESBUILD_IN := $(wildcard resources/ts/*.ts) $(wildcard resources/ts/*.js)
ESBUILD_OUT := $(addsuffix .js,$(basename $(patsubst resources/ts/%,assets/js/%,$(ESBUILD_IN))))
GEN += $(ESBUILD_OUT)
$(ESBUILD_OUT) &: $(ESBUILD_IN) node_modules
	@mkdir -p assets/js
	$(NPX) --yes esbuild $(ESBUILD_IN) --bundle --outdir=assets/js $(LOG)
	@touch $@

.PHONY: clean/assets/js
clean/assets/js:
	rm -rf $(ESBUILD_OUT)

GEN += assets/css/styles.css
assets/css/styles.css: resources/css/tailwind.css $(TEMPLATES) node_modules
	$(NPX) --yes @tailwindcss/cli -i ./resources/css/tailwind.css -o ./$@ $(LOG)
	@touch $@

.PHONY: clean/assets/css
clean/assets/css:
	rm -rf assets/css/styles.css

.PHONY: clean/assets
clean/assets:
	rm -rf $(GEN)

.PHONY: gen
gen: $(GEN)

$(BIN): $(GEN)
	$(GO_ENV) $(GO) build -o $(BIN) $(LOG)

.PHONY: build
build: $(BIN)

.PHONY: clean/build
clean/build:
	rm -rf $(BIN) $(LOGFILE)

.PHONY: run
run: $(GEN)
	$(GO) run .

# dist: build scripts/release.sh
# 	./scripts/release.sh

live/sql:
	@$(GO) run $(AIR) \
	--build.cmd "$(GO) tool $(SQLC) generate" \
	--build.bin "/bin/true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "database" \
	--build.include_ext "sql" \
	--log.main_only "true" $(LIVELOG)

live/templ:
	@printf "\033[1mstarting templ proxy, url:\033[0m \033[32m\033[4m%s\033[0m\033[0m\n" "http://localhost:7331"
	@$(GO) tool $(TEMPL) generate --watch --proxy="http://localhost:8080" --open-browser=false --log-level="warn" $(LIVELOG)

live/server: $(GEN)
	@$(GO) run $(AIR) \
	--build.cmd "$(GO) build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go,js,css" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true \
	--log.main_only "true" $(LIVELOG)

live/tailwind: node_modules
	@$(GO) run $(AIR) \
	--build.cmd "$(NPX) --yes @tailwindcss/cli -i ./resources/css/tailwind.css -o ./assets/css/styles.css" \
	--build.bin "/bin/true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "resources,templates" \
	--build.include_ext "js,css,templ" \
	--log.main_only "true" $(LIVELOG)

live/esbuild: node_modules
	@$(NPX) --yes esbuild ./resources/ts/index.ts --bundle --outdir=assets/js --watch=forever $(LIVELOG)

live/sync_assets: assets
	@$(GO) run $(AIR) \
	--build.cmd "$(GO) tool github.com/a-h/templ/cmd/templ generate --notify-proxy" \
	--build.bin "/bin/true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css" \
	--log.main_only "true" $(LIVELOG)

.PHONY: live/log-notice
live/log-notice:
	@printf "\033[1m\033[34m%s\033[0m\033[0m\n" "logs are saved into $(LIVELOGFILE)"

# See:
# https://templ.guide/developer-tools/live-reload-with-other-tools
live: live/log-notice gen
	@$(MAKE) -j live/sql live/templ live/server live/esbuild live/sync_assets live/tailwind

.PHONY: clean/live
clean/live:
	rm -rf $(LIVELOGFILE)

$(DIST_FOLDER)/$(BIN)-amd64-linux: $(GEN)
	@mkdir -p $(DIST_FOLDER)
	$(GO_ENV) GOARCH=amd64 GOOS=linux   $(GO) build $(DIST_FLAGS) -o $@ $(DISTLOG)

$(DIST_FOLDER)/$(BIN)-amd64-windows.exe: $(GEN)
	@mkdir -p $(DIST_FOLDER)
	$(GO_ENV) GOARCH=amd64 GOOS=windows $(GO) build $(DIST_FLAGS) -o $@ $(DISTLOG)

.PHONY: dist/amd64
dist/amd64: $(DIST_FOLDER)/$(BIN)-amd64-linux $(DIST_FOLDER)/$(BIN)-amd64-windows.exe

$(DIST_FOLDER)/$(BIN)-arm64-linux: $(GEN)
	@mkdir -p $(DIST_FOLDER)
	$(GO_ENV) GOARCH=arm64 GOOS=linux   $(GO) build $(DIST_FLAGS) -o $@ $(DISTLOG)

$(DIST_FOLDER)/$(BIN)-arm64-windows.exe: $(GEN)
	@mkdir -p $(DIST_FOLDER)
	$(GO_ENV) GOARCH=arm64 GOOS=windows $(GO) build $(DIST_FLAGS) -o $@ $(DISTLOG)

.PHONY: dist/arm64
dist/arm64: $(DIST_FOLDER)/$(BIN)-arm64-linux $(DIST_FOLDER)/$(BIN)-arm64-windows.exe

.PHONY: dist
dist: dist/amd64 dist/arm64

.PHONY: clean/dist
clean/dist:
	rm -rf $(DIST_FOLDER) $(DISTLOGFILE)

.PHONY: clean
clean:
	$(MAKE) -j clean/assets/js clean/assets/css clean/node_modules clean/sql clean/templates clean/build clean/live clean/dist
