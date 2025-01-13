# Changelog

All notable changes to the Base CLI tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Support for attachment field type in module generation
- New `feed` command for importing data from JSON files
- Improved error messages in the `start` command

### Changed
- Updated module template to handle storage dependency more elegantly
- Improved search functionality in generated services
- Fixed path resolution in `start` command

### Fixed
- Fixed directory path resolution in `start` command
- Fixed SQL query generation for search functionality
- Fixed storage handling in module generation

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
