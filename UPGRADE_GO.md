# Upgrading Go Version

## Current Status

- **Project Go version**: 1.19 (compatible with system)
- **System Go version**: 1.19.2 (needs upgrade to 1.21+ for VS Code tools)

## Issue

VS Code Go extension requires Go 1.21.0 or newer for tool installation, but the system has Go 1.19.2.

## Solutions

### Option 1: Upgrade System Go (Recommended)

#### Using Homebrew (macOS):

```bash
# Remove old Go
sudo rm -rf /usr/local/go

# Install latest Go
brew install go

# Or install specific version
brew install go@1.21

# Verify installation
go version
```

#### Manual Installation:

1. Download Go 1.21+ from https://golang.org/dl/
2. Remove old installation:
   ```bash
   sudo rm -rf /usr/local/go
   ```
3. Extract new version:
   ```bash
   sudo tar -C /usr/local -xzf go1.21.x.darwin-amd64.tar.gz
   ```
4. Verify:
   ```bash
   go version
   ```

### Option 2: Use Go Version Manager (g)

```bash
# Install g
curl -sSL https://git.io/g-install | sh -s

# Install Go 1.21
g install 1.21.0

# Use it
g 1.21.0
```

### Option 3: Configure VS Code to Use Different Go

If you have multiple Go versions installed, configure VS Code:

1. Open VS Code settings (`.vscode/settings.json`)
2. Set `go.toolsManagement.go` to point to Go 1.21+:
   ```json
   {
     "go.toolsManagement.go": "/path/to/go1.21/bin/go"
   }
   ```

### Option 4: Disable Tool Auto-installation

If you want to use Go 1.19 for now:

1. Open VS Code settings
2. Set:
   ```json
   {
     "go.toolsManagement.checkForUpdates": "off"
   }
   ```
3. Manually install tools:
   ```bash
   go install golang.org/x/tools/gopls@latest
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

## Verify Installation

After upgrading:

```bash
go version
# Should show: go version go1.21.x darwin/amd64

go env GOROOT
go env GOPATH
```

## Project Compatibility

The project currently uses Go 1.19 in `go.mod`, which is compatible with your current system Go version (1.19.2).

After upgrading your system Go to 1.21+, you can optionally update `go.mod` to `go 1.21` to take advantage of newer features, but it's not required - Go 1.21 is backward compatible with Go 1.19 code.

## Notes

- Go 1.21 is backward compatible with Go 1.19 code
- All existing code should work without changes
- Some new features from Go 1.21 are available but not required

