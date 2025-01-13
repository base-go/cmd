# Changelog

All notable changes to the Base CLI tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Support for attachment field type in module generation
- New `feed` command for importing data from JSON files
- Improved error messages in the `start` command
- Hot reload functionality using Air in `start` command
- Automatic `.air.toml` configuration generation
- Comprehensive documentation for hot reload feature in README

### Changed
- Updated module template to handle storage dependency more elegantly
- Improved search functionality in generated services
- Fixed path resolution in `start` command
- Refactored template generation for Air configuration
- Improved error handling in `start` command
- Enhanced code organization with dedicated Air utilities
- Improved Swagger documentation generation in hot reload mode

### Fixed
- Fixed directory path resolution in `start` command
- Fixed SQL query generation for search functionality
- Fixed storage handling in module generation
- Fixed "(no value) used as value" error in template generation
- Improved error handling in Air configuration setup
- Updated Air installation to use new repository (github.com/air-verse/air)
- Fixed hot reload loop by excluding docs directory and improving process management
- Fixed Swagger documentation triggering unnecessary rebuilds in hot reload mode

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

[Unreleased]: https://github.com/base-go/cmd/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/base-go/cmd/releases/tag/v1.0.0
