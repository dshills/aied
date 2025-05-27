# Changelog

All notable changes to AIED will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Core VIM-style modal editing (Normal, Insert, Visual, Command modes)
- Basic VIM navigation commands (h/j/k/l, w/b/e, 0/$, gg/G)
- Text manipulation (x, dd, yy, p, u, Ctrl-R)
- File operations (:w, :q, :wq, :q!, :e, :new)
- AI integration with multi-provider support:
  - OpenAI (GPT-4, GPT-3.5)
  - Anthropic (Claude 3.5 Sonnet)
  - Google (Gemini 1.5)
  - Ollama (Local models)
- AI commands (:ai, :aic, :aie, :air, :aip)
- Configuration system with YAML/JSON support
- Configuration commands (:config, :configgen, :configreload)
- Environment variable support for API keys
- Automatic AI provider fallback
- Context-aware AI assistance
- Line-based buffer system with undo/redo
- Cross-platform terminal UI using tcell

### Changed
- N/A (Initial release)

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- API keys are loaded from environment variables or config files
- Config files with API keys are gitignored by default

## [0.1.0] - TBD

Initial release of AIED - AI-Powered Terminal Editor

[Unreleased]: https://github.com/dshills/aied/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dshills/aied/releases/tag/v0.1.0