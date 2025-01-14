# Changelog

All notable changes to the Base CLI tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.13] - 2025-01-14

### Added
- Cross-platform build support for all architectures (amd64, arm64)
- New `build-all` make target for building all platform binaries

### Fixed
- Version string handling in Makefile to avoid double 'v' prefix
- Upgrade command now properly handles compressed binaries (.tar.gz, .zip)
- Installation path now matches install.sh behavior
- Binary architecture detection for M1/M2 Macs
- Fixed unused variable in upgrade command

### Changed
- Improved release process to support all platforms consistently
- Installation now uses ~/.base directory for binary storage

## [v1.0.12] - 2025-01-14

### Added
- Optional Swagger documentation generation with `--docs` flag
- Support for specialized attachment types (`image` and `file`) with validation
- Type-specific file validation (size limits and allowed extensions)

### Changed
- Made Swagger documentation generation optional (use `--docs` flag)
- Updated installation path to use /usr/local/bin
- Improved version information handling
- Enhanced attachment field generation to avoid duplicates

### Fixed
- Fixed version information display in CLI
- Fixed duplicate field generation in attachment handling
- Fixed polymorphic association tags in model generation
- Fixed attachment field validation in create/update requests

## [v1.0.5] - 2025-01-14

### Added
- PowerShell installation script for Windows users
- Support for multiple platforms (Windows, macOS, Linux) in installation scripts
- Automated release script for creating new releases
- Support for multiple architectures (amd64, arm64)

### Changed
- Improved upgrade command to use GitHub release assets
- Updated installation process to use pre-built binaries
- Enhanced version command to show release notes
- Simplified installation instructions in README

### Fixed
- Version information display in CLI output
- Binary installation paths for different operating systems
- Upgrade process on Windows systems

## [1.0.0] - 2024-01-13

### Added
- Initial release of Base CLI
- Core commands: new, generate, start, destroy, update, upgrade
- Module generation with various field types
- Relationship support: belongsTo, hasOne, hasMany
- Basic CRUD operations in generated modules
- Search functionality in list endpoints
- Pagination support
- Storage system for file uploads
- Basic authentication system
- Database migrations
- Module templates
- Service layer with CRUD operations
- Controller layer with REST endpoints
- Model layer with GORM integration
- Basic project structure

[v1.0.13]: https://github.com/base-go/cmd/releases/tag/v1.0.13
[v1.0.12]: https://github.com/base-go/cmd/releases/tag/v1.0.12
[v1.0.5]: https://github.com/base-go/cmd/releases/tag/v1.0.5
[1.0.0]: https://github.com/base-go/cmd/releases/tag/v1.0.0
