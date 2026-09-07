EXECUTABLE = gdiff
BUILD_DIR = bin
GORELEASER = go run github.com/goreleaser/goreleaser/v2@latest

.PHONY: build build-mos run dev check tui-studio release release-confirm clean

build:
	go build -o $(BUILD_DIR)/$(EXECUTABLE)

build-mos: build
	cp bin/gdiff ~/.local/bin/gdiff

run: build
	./$(BUILD_DIR)/$(EXECUTABLE)

dev: build
	./$(BUILD_DIR)/$(EXECUTABLE) --dev

check:
	go build ./...
	go vet ./...
	gofmt -l .

tui-studio:
	docker run --name tui-studio -p 8080:80 javieralonso716/tui-studio-web:latest

release:
	$(GORELEASER) release --snapshot --clean --skip=publish
	@echo
	@echo "Dry run complete. Inspect dist/homebrew/Casks/$(EXECUTABLE).rb and the archives,"
	@echo "then run: make release-confirm VERSION=vX.Y.Z"

release-confirm:
	@if [ -z "$(VERSION)" ]; then echo "usage: make release-confirm VERSION=v0.1.0"; exit 1; fi
	@case "$(VERSION)" in v*) ;; *) echo "VERSION must start with 'v' — the workflow only triggers on v* tags"; exit 1;; esac
	@if [ -n "$$(git status --porcelain)" ]; then echo "working tree is dirty; commit or stash first"; exit 1; fi
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then echo "not on main"; exit 1; fi
	@if git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null; then echo "tag $(VERSION) already exists"; exit 1; fi
	@git fetch --quiet origin main
	@if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse origin/main)" ]; then echo "main and origin/main differ; push or pull first"; exit 1; fi
	@printf 'Tag and push %s? This publishes a public release. [y/N] ' "$(VERSION)"; \
		read -r reply; case "$$reply" in y|Y|yes|Yes|YES) ;; *) echo "aborted"; exit 1;; esac
	git tag $(VERSION)
	git push origin $(VERSION)
	@echo "Pushed $(VERSION). Watch: https://github.com/JNSAPH/gdiff/actions"

clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist
