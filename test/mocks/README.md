# Mocks Directory

This directory contains mock implementations for testing purposes.

## Mocks Available

- `MockPlatform` - Mock implementation of `internal/platform.Platform`
- `MockJDKProvider` - Mock implementation of `internal/provider.JDKProvider`
- `MockGradleProvider` - Mock implementation of `internal/provider.GradleProvider`
- `MockConfigRepository` - Mock implementation of `internal/config.ConfigRepository`
- `MockPrompter` - Mock implementation of `cmd.Prompter`

## Usage

Each mock struct has optional function fields that allow you to customize behavior:

```go
platform := &mocks.MockPlatform{
    HomeDirFunc: func() string {
        return "/tmp/test-home"
    },
    JDKDetectionPathsFunc: func() []string {
        return []string{"/tmp/test-jdks"}
    },
}
```

If a function field is not set, the mock returns sensible defaults:

```go
platform := &mocks.MockPlatform{}
platform.HomeDir() // Returns "/tmp" (default)
```

## Generating Mocks (Optional)

If you prefer to use `mockgen` for generating mocks, install it first:

```bash
go install go.uber.org/mock/mockgen@latest
```

Then generate mocks:

```bash
mockgen -source=internal/platform/interface.go -destination=test/mocks/platform_mock.go
mockgen -source=internal/provider/provider.go -destination=test/mocks/provider_mock.go
mockgen -source=internal/config/repository.go -destination=test/mocks/config_mock.go
```

Note: The current mocks are manually written inline in this file for simplicity and to avoid external dependencies.
