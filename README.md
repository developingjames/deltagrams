# Deltagram

A powerful command-line tool for applying structured file changes using the deltagram format - an advanced evolution of the mimeogram format that supports delta operations for efficient file manipulation.

## Overview

Deltagram enables you to apply complex file changes including:
- **Creating** new files with content
- **Modifying** existing files using unified diff format  
- **Copying** files to new locations
- **Moving/renaming** files
- **Deleting** files

All operations are bundled into a single deltagram that can be read from the clipboard and applied atomically to your project.

## Features

- 🔄 **Delta Operations**: CREATE, MODIFY, COPY, MOVE, DELETE
- 📋 **Clipboard Integration**: Read deltagrams directly from clipboard
- 🎯 **Unified Diff Support**: Industry-standard patch format for content changes
- 🔒 **Atomic Operations**: All changes applied in correct order or none at all
- 🧪 **Comprehensive Testing**: Unit and integration test coverage
- 🌍 **Cross-platform**: Linux, macOS, Windows support
- ⚡ **Fast & Reliable**: Written in Go for performance

## Installation

### Prerequisites

- Go 1.21 or later
- Clipboard utilities:
  - **Linux**: `xclip` or `xsel`
  - **macOS**: Built-in `pbpaste`
  - **Windows**: Built-in PowerShell clipboard

### Building from Source

```bash
# Clone the repository
git clone https://github.com/developingjames/deltagrams.git
cd deltagrams

# Build for your platform
make build

# Install to $GOPATH/bin
make install

# Build for all platforms
make build-all
```

### Binary Downloads

Download pre-built binaries from the [releases page](https://github.com/developingjames/deltagrams/releases).

## Usage

### Basic Usage

```bash
# Apply deltagram from clipboard to current directory
deltagram apply

# Show version information
deltagram version

# Show help
deltagram help
```

### Example Workflow

1. Copy a deltagram to your clipboard (from an LLM, text editor, etc.)
2. Navigate to your project directory
3. Run `deltagram apply`
4. All file operations are applied atomically

## Deltagram Format

Deltagrams use a structured format based on mimeograms with added delta operation support. See [deltagram_prompt.md](deltagram_prompt.md) for the complete specification.

### Using with LLMs

To enable deltagram generation in AI assistants:

1. Copy the contents of [deltagram_prompt.md](deltagram_prompt.md)
2. Paste it into your LLM conversation or system prompt
3. Ask the AI to generate deltagrams for your file changes
4. Copy the generated deltagram and use `deltagram apply`

This teaches the AI assistant the proper deltagram format and validation requirements.

### Example Deltagram

```
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Adding logging functionality and fixing bug in main function.
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/logger.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ src/logger.py
import logging

class Logger:
    def __init__(self, name="app"):
        self.logger = logging.getLogger(name)
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/main.py
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -1,5 +1,8 @@
+from logger import Logger
+
 def main():
+    logger = Logger()
+    logger.info("Starting application")
     print("Hello, world!")
-    return 0
+    return True
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--
```

## Development

### Project Structure

```
deltagram/
├── cmd/deltagram/           # Main CLI application
├── pkg/
│   ├── parser/             # Deltagram parsing logic
│   ├── operations/         # File operation handlers
│   └── clipboard/          # Clipboard interface
├── test/integration/       # Integration tests
├── internal/testutil/      # Test utilities
├── .github/workflows/      # CI/CD pipelines
└── bin/                    # Build output
```

### Building & Testing

```bash
# Run unit tests
make test

# Run integration tests  
make test-integration

# Run all tests
make test-all

# Run tests with race detection
make test-race

# Generate coverage report
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Full release preparation
make release-prep
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `build` | Build for current platform |
| `build-all` | Build for all supported platforms |
| `build-windows` | Build for Windows (.exe) |
| `build-linux` | Build for Linux |
| `build-darwin` | Build for macOS |
| `test` | Run unit tests |
| `test-integration` | Run integration tests |
| `test-all` | Run all tests |
| `clean` | Clean build artifacts |
| `install` | Install to $GOPATH/bin |

### Adding New Operations

1. Create a new handler in `pkg/operations/`
2. Implement the `OperationHandler` interface
3. Register it in `pkg/operations/applier.go`
4. Add comprehensive tests

### Dependencies

This project uses only Go standard library and has no external dependencies for the core functionality.

## Architecture

### Components

- **Parser**: Parses deltagram format into structured data
- **Operations**: Handles different types of file operations
- **Clipboard**: Cross-platform clipboard reading
- **CLI**: Command-line interface and version management

### Design Principles

- **Dependency Injection**: All components use interfaces for testability
- **Separation of Concerns**: Clear boundaries between parsing, operations, and I/O
- **Error Handling**: Comprehensive error reporting with context
- **Cross-platform**: Native support for major operating systems

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test-all`)
6. Format code (`make fmt`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code Standards

- Follow standard Go conventions
- Add tests for all new functionality
- Update documentation for user-facing changes
- Keep commits atomic and well-described

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- **[python-mimeogram](https://github.com/emcd/python-mimeogram)**: The original mimeogram implementation by Eric McDonald. We owe a huge debt of gratitude to Eric for his excellent work and inspiration that made this project possible. The deltagram format builds upon the solid foundation of mimeograms.
- **Git**: Inspiration for unified diff format
- **GNU Patch**: Reference implementation for patch application

## Support

- 📋 **Issues**: [GitHub Issues](https://github.com/developingjames/deltagrams/issues)
- 📖 **Documentation**: [deltagram_prompt.md](deltagram_prompt.md)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/developingjames/deltagrams/discussions)

---

**Note**: This tool is designed to work with LLMs and AI assistants that can generate deltagrams. Copy and paste the contents of [deltagram_prompt.md](deltagram_prompt.md) into your AI assistant's conversation to enable deltagram generation.