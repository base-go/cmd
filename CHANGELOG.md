# Changelog

All notable changes to the Base CLI tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.2.2] - 2025-03-18

### Changed
- Updated model template to improve database compatibility:
  - Changed attachment field gorm tag from `polymorphic:Model` to `foreignKey:ModelId;references:Id`
  - Updated table name and model name functions to use `ToSnakeCase` instead of `toLower` for proper snake_case formatting
- Fixed boolean field handling in Update requests to use pointer types

## [v1.2.1] - 2025-03-18

### Fixed
- Fixed field type handling in service template for Update method:
  - Boolean fields now use pointer check instead of string comparison
  - Date/time fields now use IsZero() method for proper validation
  - Integer and float fields use appropriate zero value checks
  - String fields continue to use empty string check
- Updated model template to use pointer types for boolean fields in Update request struct
- This fixes issues with boolean fields being incorrectly compared to empty strings and compiler errors related to nil checks on non-pointer boolean fields

## [v1.2.0] - 2025-02-06

### Added
- New `base package` command for better package management
  - `base package add [namespace/package]` to install packages from GitHub (e.g., `base package add base-packages/gamification`)
  - `base package add [URL]` to install packages from any Git repository
  - `base package remove [package]` to remove installed packages
- Automatic module initialization in start.go when adding packages
- List official packages when running `base package add` without arguments

### Changed
- Reorganized package management commands from `base add` to `base package add/remove`
- Improved package installation with better error handling and import organization

## [v1.1.9] - 2025-02-06

### Added
- Support for installing packages via GitHub URLs using `base add https://github.com/org/package`
- Automatic module initialization in start.go when adding packages
- List official packages when running `base add` without arguments

## [v1.1.8] - 2025-02-04

### Changed
- Updated Controller to accept Bearer token for authentication


## [v1.1.7] - 2025-01-29

### Changed
- Updated route generation to use kebab-case for better URL consistency
- Changed datetime field type to use types.DateTime instead of time.Time
- Added default pagination values in service template GetAll method

## [v1.1.6] - 2025-01-29

### Fixed
- Fixed version comparison to handle 'v' prefix correctly, resolving incorrect update notifications

## [v1.1.5] - 2025-01-29

### Changed
- Updated templates and utils for improved code generation
- Enhanced controller template handling

## [v1.1.4] - 2025-01-21

### Fixed
- Fixed service template to use correct package name instead of pluralized name

## [v1.1.3] - 2025-01-21

### Fixed
- Fixed logger interface consistency across modules
- Fixed attachment deletion in controllers to use direct field access
- Improved template code to handle logger interfaces correctly
- Simplified field processing in template generation
- Removed unused code and simplified relationship handling

## [v1.1.1] - 2025-01-14

### Fixed
- Fixed version comparison showing update message when already on latest version
- Added explicit "up-to-date" message when running version command
- Improved version check messages across all commands

## [v1.1.0] - 2025-01-14

### Added
- Improved cross-platform upgrade command with better Windows support
- OS-specific binary handling (base.exe for Windows)
- Better error handling and user feedback during upgrade
- Centralized version management in version package

### Changed
- Unified installation behavior between install.sh and upgrade command
- Enhanced directory structure handling for different operating systems
- Improved sudo handling for Unix systems
- Consolidated version handling across all commands

### Fixed
- Windows installation path now uses correct user profile directory
- Binary permissions are now properly set during upgrade
- Symlink handling improved for Unix systems
- Version consistency across version and upgrade commands

## [v1.0.14] - 2025-01-14

### Fixed
- Fixed unused variable in upgrade command
- Improved upgrade process reliability

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

[v1.2.0]: https://github.com/base-go/cmd/releases/tag/v1.2.0
[v1.1.9]: https://github.com/base-go/cmd/releases/tag/v1.1.9
[v1.1.8]: https://github.com/base-go/cmd/releases/tag/v1.1.8
[v1.1.7]: https://github.com/base-go/cmd/releases/tag/v1.1.7
[v1.1.6]: https://github.com/base-go/cmd/releases/tag/v1.1.6
[v1.1.5]: https://github.com/base-go/cmd/releases/tag/v1.1.5
[v1.1.4]: https://github.com/base-go/cmd/releases/tag/v1.1.4
[v1.1.3]: https://github.com/base-go/cmd/releases/tag/v1.1.3
[v1.1.2]: https://github.com/base-go/cmd/releases/tag/v1.1.2
[v1.1.1]: https://github.com/base-go/cmd/releases/tag/v1.1.1
[v1.1.0]: https://github.com/base-go/cmd/releases/tag/v1.1.0
[v1.0.14]: https://github.com/base-go/cmd/releases/tag/v1.0.14
[v1.0.13]: https://github.com/base-go/cmd/releases/tag/v1.0.13
[v1.0.12]: https://github.com/base-go/cmd/releases/tag/v1.0.12
[v1.0.5]: https://github.com/base-go/cmd/releases/tag/v1.0.5
[1.0.0]: https://github.com/base-go/cmd/releases/tag/v1.0.0
