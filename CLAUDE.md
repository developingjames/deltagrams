# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **deltagrams** repository, which contains specifications and documentation for the mimeogram format - a structured format for exchanging collections of text files from hierarchical directory structures or disparate sources.

## Architecture

The project currently consists of:
- **mimeogram.md**: The complete format specification for mimeograms, including structure, boundary markers, headers, content guidelines, and processing instructions
- **LICENSE**: MIT license for the project

## Mimeogram Format Key Points

When working with mimeograms in this repository:

### Structure Requirements
- Parts separated by boundary markers: `--====MIMEOGRAM_{uuid}====`
- Final boundary must end with `====--`
- UUID must be exactly 32 lowercase hexadecimal characters
- Each part requires `Content-Location` and `Content-Type` headers

### Content Guidelines
- Message parts use `Content-Location: mimeogram://message`
- File parts use original filesystem paths or URLs
- Content normalized to UTF-8 with Unix (LF) line endings
- Always validate the five key points before delivering mimeograms (boundaries, UUID format, headers, line endings, no placeholders)

### Processing Notes
- Only provide mimeogram responses when explicitly requested
- Use artifacts/canvases for mimeogram output when available
- Treat mimeograms as plain text rather than Markdown when writing to artifacts
- Preserve original file paths unless explicitly asked to change them

## Development Notes

This appears to be a specification/documentation project rather than a traditional software project. There are no build scripts, dependency files, or code to compile. Changes would primarily involve updating the mimeogram specification or adding new documentation files.