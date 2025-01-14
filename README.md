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
- `string`: String field
- `int`: Integer field
- `bool`: Boolean field
- `float`: Float field
- `text`: Text field (for longer strings)

Special Types:
- `image`: Image attachment with validation (5MB limit, image extensions)
- `file`: File attachment with validation (50MB limit, document extensions)
- `attachment`: Generic attachment (10MB limit, mixed extensions)
- `time`: Time field
- `date`: Date field
- `datetime`: DateTime field

Relationship Types:
- `belongs_to`: One-to-one relationship
- `has_one`: One-to-one relationship
- `has_many`: One-to-many relationship
- `many2many`: Many-to-many relationship

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
  comments:hasMany:Comment

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
