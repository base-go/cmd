# Base - Command Line Tool for the Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework.
It offers scaffolding, module generation, and utilities to accelerate Go application development.

## Installation

### macOS and Linux

```bash
curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | bash
```

If you need to install in a protected directory (like `/usr/local/bin`), use:

```bash
curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | sudo bash
```

### Windows

#### Option 1: Using PowerShell (Recommended)

1. Open PowerShell as Administrator
2. Run:
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/base-go/cmd/main/install.ps1'))
```

#### Option 2: Using Git Bash

```bash
curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | bash
```

## Commands

### `base new <project-name>`

Create a new project using the Base framework.

```bash
base new myapp
```

### `base generate` or `base g`

Generate a new module with fields and relationships.

```bash
base g <module-name> [field:type ...] [options]
```

### `base start` or `base s`

Start the Base application server.

Options:
- `--hot-reload`, `-r`: Enable hot reloading using air
- `--docs`, `-d`: Generate Swagger documentation

Examples:
```bash
# Start the server normally
base start

# Start with hot reloading
base start -r

# Start with Swagger documentation
base start -d

# Start with both hot reloading and Swagger docs
base start -r -d
```


### `base docs`

Generate OpenAPI 3.0 documentation by scanning controller annotations and create static files.

```bash
base docs [flags]
```

Options:
- `-o, --output string`: Output directory for generated files (default "docs")
- `-s, --static`: Generate static swagger files (JSON, YAML, docs.go) (default true)
- `--no-static`: Skip generating static files

Examples:
```bash
# Generate docs in default 'docs' directory
base docs

# Generate docs in custom directory
base docs --output api-docs

# Skip static file generation (legacy mode)
base docs --no-static
```

Generated files:
- `swagger.json`: OpenAPI 3.0 specification in JSON format
- `swagger.yaml`: OpenAPI 3.0 specification in YAML format
- `docs.go`: Go package with embedded OpenAPI spec for programmatic access

Notes:
- Static files are served at `/docs/` when running the server
- You can also run `base start -d` to auto-generate docs before starting the server and serve Swagger UI at `/swagger/`
- All swagger info (title, version, description) is extracted from main.go annotations

### `base d` or `base destroy`

Destroy (delete) one or more existing modules.

```bash
base d [name1] [name2] ... [flags]
```

Examples:
```bash
# Destroy a single module
base d user

# Destroy multiple modules at once
base d user customer order

# Alternative command name
base destroy user customer
```

What gets removed:
- Module directory (`app/modulename/`)
- Model file (`app/models/modulename.go`)
- Import and registration from `app/init.go`

Notes:
- Requires confirmation before destroying modules
- Will attempt to clean up orphaned entries even if module directory doesn't exist
- Shows progress for each module when destroying multiple modules

### `base update`

Update framework core components:

```bash
base update
```

### `base upgrade`

Upgrade the Base CLI tool:

```bash
base upgrade
```

### `base version`

Display version information:

```bash
base version
```

## Field Types

Base supports various field types for model generation:

Basic Types:
- `string`
- `bool`
- `int`, `uint` (also `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`)
- `float`, `float32`, `float64`
- `text` (stored as string with appropriate DB type)

Special Types and Aliases (mapping shown on the right):
- `email`, `url`, `slug` → string
- `datetime`, `time`, `date` → `types.DateTime`
- `decimal`, `float` → `float64`
- `sort` → `int`
- `translation`, `translatedField` → `translation.Field`
- `image`, `file`, `attachment` → `*storage.Attachment`

Notes:
- Attachment fields are handled via dedicated upload endpoints and are not included in JSON create/update payloads.
- Email/URL/Slug are strings; GORM tags may add size/indexing automatically.
- Datetime types use Base `types.DateTime` under the hood.

Relationship Types (both snake_case and camelCase accepted):
- `belongs_to` (or `belongsTo`): one-to-one with FK on this model
- `has_one` (or `hasOne`): one-to-one with FK on the other model
- `has_many` (or `hasMany`): one-to-many
- `to_many` (or `toMany`): many-to-many with join table

Relationship auto-detection:
- Defining a field as `<name>_id:uint` will also generate the corresponding `belongs_to` relationship for `<name>` automatically.

Type inference (when no explicit type is given):
- `<name>_id` → `uint` (FK)
- Contains one of: `description`, `content`, `body`, `notes`, `comment`, `summary`, `bio`, `about` → `text`
- Prefix or contains: `is_`, `has_`, `can_`, `enabled`, `active`, `published`, `verified`, `confirmed` → `bool`
- Contains: `price`, `amount` → `decimal`; other numeric-like names (`count`, `quantity`, `number`, `rating`, `score`, `weight`, `height`, `width`) → `int`
- Suffix `_at`, `_on`, `_date` or contains common datetime terms (`date`, `time`, `created_at`, `updated_at`, `deleted_at`, `published_at`, `expires_at`) → `datetime`
- Contains `email` → `email` (string)
- Contains `url` or `link` → `url` (string)
- Contains `image`, `photo`, `picture`, `avatar` → `image` (attachment)
- Contains `file`, `document`, `attachment` → `file` (attachment)
- Contains `translation`, `i18n`, `locale` → `translatedField`
- Otherwise → `string`

Example:
```bash
# Generate a post module with title and image
base g post title:string cover:image

# Generate a document module with title and file attachment
base g document title:string file:file
```

## Example: Building a Blog System

Here's a comprehensive example of building a blog system with categories, posts, tags, and comments:

```bash
# Generate Category model
base g Category \
  name:string \
  description:text \
  image:attachment \
  parent:belongsTo:Category \
  posts:hasMany:Post

# Generate Post model
base g Post \
  title:string \
  content:text \
  excerpt:text \
  featured_image:attachment \
  gallery:attachment \
  published_at:datetime \
  author:belongsTo:users.User \
  category:belongsTo:Category \
  comments:hasMany:Comment \
  tags:toMany:Tag

# Generate Tag model
base g Tag \
  name:string \
  slug:string 
 

# Generate Comment model
base g Comment \
  content:text \
  author:belongsTo:users.User \
  post:belongsTo:Post
```

This will create:
- Full CRUD operations for all models
- RESTful API endpoints with Swagger documentation
- File upload handling for images
- Proper relationships and preloading
- Authentication and authorization integration

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
