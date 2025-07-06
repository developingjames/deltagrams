# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **deltagrams** repository, which contains a Go command-line tool for applying structured file changes using the deltagram format - an advanced evolution of the mimeogram format that supports delta operations.

## Architecture

The project is a standard Go application with the following structure:
- **cmd/deltagram/**: Main CLI application
- **pkg/**: Reusable packages (parser, operations, clipboard)
- **test/integration/**: Integration tests
- **internal/testutil/**: Test utilities  
- **deltagram_prompt.md**: Complete format specification and LLM instructions for deltagrams
- **Makefile**: Standard build automation
- **LICENSE**: MIT license

## Deltagram Format Key Points

When working with deltagrams in this repository:

### Structure Requirements
- Parts separated by boundary markers: `--====DELTAGRAM_{identifier}====`
- Final boundary must end with `====--`
- Identifier must be at least 8 characters using alphanumeric, underscore, or dash (a-z, A-Z, 0-9, _, -) for reasonable uniqueness
- Each part requires `Content-Location`, `Content-Type`, and optionally `Delta-Operation` headers

### Delta Operations Supported
- **create**: Create new files
- **delete**: Delete existing files  
- **copy**: Copy files to new locations
- **move**: Move/rename files
- **content**: Modify file content using unified diff format

### Content Guidelines
- Message parts use `Content-Location: deltagram://message`
- File parts use original filesystem paths or URLs
- Content normalized to UTF-8 with Unix (LF) line endings
- Content operations use standard unified diff format (`@@` hunks with `+`, `-`, ` ` prefixes)

### Processing Notes
- Operations are applied in the order they appear in the deltagram
- Only provide deltagram responses when explicitly requested
- Use artifacts/canvases for deltagram output when available
- Treat deltagrams as plain text rather than Markdown when writing to artifacts
- Always validate format before delivering deltagrams

## Development Notes

This is a production Go project with:
- **Standard Go project layout** following community conventions
- **Comprehensive test suite** with unit and integration tests
- **Cross-platform builds** for Linux, macOS, Windows
- **CI/CD pipelines** using GitHub Actions
- **Makefile automation** for common development tasks

Use `make help` to see available build targets. The main executable is built to `bin/deltagram`.