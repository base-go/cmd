# Changelog

All notable changes to the Base CLI tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - 2025-08-13

### Added
- **🎯 SelectOption Support for Frontend Select Boxes** - Enhanced `/all` endpoints
  - New `SelectOption` struct with just `id` and `name` fields for dropdowns
  - Intelligent field detection (prioritizes `name`, falls back to `title`)
  - Support for `translation.Field` types with automatic string extraction
  - Optimized database queries - only selects necessary fields
  - Alphabetically sorted results for better UX

### Changed
- **🔧 Template System Improvements**
  - Merged "_clean" template files into main templates for cleaner organization
  - Refactored `GenerateFileFromTemplate` to use `NamingConvention` struct
  - Updated all templates to use consistent naming convention fields
  - Fixed controller routes to be explicit (e.g., `/users` instead of `""`)

### Fixed
- **🛠️ GORM Field Optimization for MySQL**
  - Added proper size tags: `email` gets `size:255;index`, `string` gets `size:255`
  - Text fields use explicit `type:text` for longer content
  - URLs get `size:512` for longer links
  - Slugs get `size:255;uniqueIndex` for unique indexing
  - Decimal fields get `type:decimal(10,2)` for proper money precision
  - Foreign keys automatically get `index` for better performance
- **🐛 Template Field Resolution**
  - Fixed template errors with missing `.StructName` → `.Model` fields
  - Corrected `.LowerName` → `.ModelLower` field references
  - Resolved `.PluralName` → `.Plural` field mappings

## [v2.1.0] - 2025-08-13

### Added
- **🎯 Enhanced Swagger Documentation System** - Complete overhaul of API documentation
  - Auto-discovery of swagger annotations from controller files
  - Enhanced `base docs` command to generate comprehensive documentation  
  - Support for detailed @Param, @Success, @Failure, and @Security annotations
  - Automatic parsing of parameter types, request bodies, and responses
  - Integration with `base start -d` flag for automatic doc generation
- **🔧 API Route Prefix Fix** - Unified API structure  
  - All swagger routes now correctly prefixed with `/api`
  - Updated controller templates to include `/api` prefix automatically
  - Fixed swagger documentation to match actual API structure
  - New modules automatically generate with correct API routes
- **📋 Controller Template Improvements**
  - Consolidated to single controller template (removed duplicate)
  - Enhanced swagger annotations with detailed parameter information
  - Better error handling and response schemas
  - Auto-generation includes security requirements

### Fixed
- Template generation now uses correct `/api` prefixed routes
- Swagger documentation matches actual server endpoints
- Controller generation produces consistent API structure


### Added
- **🚀 MAJOR: Automatic Relationship Detection** - Revolutionary enhancement to code generation
  - Fields ending with `_id` and type `uint` automatically generate GORM relationships
  - Eliminates manual relationship specification - just use naming conventions!
  - Auto-generates both foreign key fields AND relationship fields
  - Includes proper GORM tags with `foreignKey` relationships
  - Works seamlessly with existing manual relationship syntax
- **Smart Field Processing**: Enhanced `ProcessField` function to detect relationship patterns
- **Template Enhancements**: Updated all templates to handle auto-detected relationships
- **Clean Code Generation**: Eliminated duplicate field generation issues
- **Major Version Upgrade Safety**: Added `--major` flag to upgrade command for controlled major version upgrades
- **Smart Upgrade Behavior**: Default `base upgrade` now only upgrades within the same major version (e.g., 1.x → 1.y)
- **Major Version Upgrade Warning**: Added interactive warning for major version upgrades with breaking change information
- **Enhanced Version Command**: Version command now shows specific information about major version upgrades
- **Upgrade Command Documentation**: Enhanced help text with examples for both minor and major upgrades

### Changed
- **BREAKING**: Enhanced relationship detection changes code generation behavior
- **Model Template**: Updated to handle both foreign key and relationship fields from auto-detection
- **Service Template**: Fixed Create/Update methods to prevent duplicate field assignments
- **Field Processing Logic**: `ProcessField` now returns multiple fields for `_id` patterns
- **Template Consistency**: All templates now work with enhanced relationship detection

### Fixed
- **Duplicate Fields**: Resolved issue where foreign key fields were generated multiple times
- **Service Layer**: Fixed Create and Update operations to properly handle relationship fields
- **Template Logic**: Corrected template conditions to avoid conflicts between manual and auto relationships

### Migration Guide
- **Existing projects**: Continue to work without changes
- **New projects**: Can use `_id` suffix convention for automatic relationships
- **Mixed usage**: Manual relationship syntax still works alongside auto-detection

### Examples

**Code Generation - Before v2.0.0:**
```bash
base g article title:string content:text author:belongsTo:Author category:belongsTo:Category
```

**Code Generation - v2.0.0 and later:**
```bash
base g article title:string content:text author_id:uint category_id:uint
```

**Upgrade Commands:**
```bash
base upgrade         # Safe: only upgrades within current major version (1.x → 1.y)
base upgrade --major # Allow major version upgrade (1.x → 2.x) with breaking changes warning
```

Both generate the same result, but v2.0.0 is much simpler!

**Generated Model (both approaches):**
```go
type Article struct {
    Id         uint     `json:"id" gorm:"primarykey"`
    Title      string   `json:"title"`
    Content    string   `json:"content"`
    AuthorId   uint     `json:"author_id"`
    Author     Author   `json:"author,omitempty" gorm:"foreignKey:AuthorId"`
    CategoryId uint     `json:"category_id"`
    Category   Category `json:"category,omitempty" gorm:"foreignKey:CategoryId"`
}
```

---

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

[v2.0.0]: https://github.com/BaseTechStack/basecmd/releases/tag/v2.0.0
[v1.2.2]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.2.2
[v1.2.1]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.2.1
[v1.2.0]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.2.0
[v1.1.9]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.9
[v1.1.8]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.8
[v1.1.7]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.7
[v1.1.6]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.6
[v1.1.5]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.5
[v1.1.4]: https://github.com/BaseTechStack/basecmd/releases/tag/v1.1.4
[v1.1.3]: https://github.com/base-go/cmd/releases/tag/v1.1.3
[v1.1.2]: https://github.com/base-go/cmd/releases/tag/v1.1.2
[v1.1.1]: https://github.com/base-go/cmd/releases/tag/v1.1.1
[v1.1.0]: https://github.com/base-go/cmd/releases/tag/v1.1.0
[v1.0.14]: https://github.com/base-go/cmd/releases/tag/v1.0.14
[v1.0.13]: https://github.com/base-go/cmd/releases/tag/v1.0.13
[v1.0.12]: https://github.com/base-go/cmd/releases/tag/v1.0.12
[v1.0.5]: https://github.com/base-go/cmd/releases/tag/v1.0.5
[1.0.0]: https://github.com/base-go/cmd/releases/tag/v1.0.0
