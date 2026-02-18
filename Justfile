default:
    @just --list

test: unittest lint fmt-check gosec tidy build

unittest:
    go test ./...

lint:
    echo "Running linter..."
    @if command -v ~/tools/ext/bin/golangci-lint >/dev/null 2>&1; then \
        ~/tools/ext/bin/golangci-lint run; \
    elif command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run; \
    else \
        echo "golangci-lint not found, falling back to go vet"; \
        echo "To install golangci-lint locally, run: just install-golangci-lint"; \
        go vet ./...; \
    fi

gosec:
    @echo "Running security scanner..."
    @if command -v ~/go/bin/gosec >/dev/null 2>&1; then \
        ~/go/bin/gosec -quiet -fmt=text ./...; \
    elif command -v gosec >/dev/null 2>&1; then \
        gosec -quiet -fmt=text ./...; \
    else \
        echo "gosec not found, skipping security scan"; \
        echo "To install gosec, run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
    fi

# Install golangci-lint to ~/tools/ext/bin
install-golangci-lint:
    @mkdir -p ~/tools/ext/bin
    GOBIN=~/tools/ext/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "golangci-lint installed locally to this project in ~/tools/ext/bin/"
    @echo "Note that ~/tools/ext/bin is not assumed to be in your PATH"

fmt:
    go fmt ./cmd/... ./pkg/...

# Check formatting without modifying files
fmt-check:
    ./tools/bin/go-fmt-check

tidy:
    go mod tidy

build:
    mkdir -p bin
    go build -ldflags "-X github.com/sfkleach/pathman/pkg/commands.Source=https://github.com/sfkleach/pathman" -o bin/pathman ./cmd/pathman

clean:
    rm -rf bin
    rm -rf _build

install:
    go install ./cmd/pathman

clean:
    rm -rf bin
    rm -rf dist
    rm -rf _build


# Initialize decision records
init-decisions:
    python3 scripts/decisions.py --init

# Add a new decision record
add-decision TOPIC:
    python3 scripts/decisions.py --add "{{TOPIC}}"

# Create a demo package for showing how to do AI assisted coding, using pathman as an example.
build-demo-pack:
    @rm -rf _build/demo-pack
    @mkdir -p _build/demo-pack
    @cp -r Justfile README.md LICENSE CONTRIBUTING.md .gitignore .github scripts tools _build/demo-pack/
    @mkdir -p _build/demo-pack/cmd/pathman
    @touch _build/demo-pack/cmd/pathman/main.go
    @mkdir -p _build/demo-pack/docs/tasks
    @cp docs/tasks/README.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-23-initial-dev.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-24a-fix-script.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-24b-manage-dirs.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-24c-pathman-clean.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-24d-pathman-init-with-bubbletea.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-25-self-install.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2025-12-26-version-check-and-update.md _build/demo-pack/docs/tasks/
    @cp docs/tasks/2026-01-19-demo-initial-dev.md _build/demo-pack/docs/tasks/
    @(cd _build && zip -qr demo-pack.zip demo-pack)
    @echo "Demo package created at _build/demo-pack.zip"

git-init:
    git init
    git add .gitignore .github *
    git commit -m "Initial commit"

