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

**Field Types**:
- `string`: String field
- `text`: Text field (for longer content)
- `int`: Integer field
- `float`: Float field
- `bool`: Boolean field
- `date`: Date field
- `datetime`: DateTime field
- `time`: Time field
- `attachment`: File upload field
- `belongsTo`: One-to-one relationship
- `hasMany`: One-to-many relationship
- `hasOne`: One-to-one relationship
- `manyToMany`: Many-to-many relationship

### `base start` or `base s`

Run your application with hot reload:

```bash
base start
```

Features:
- Automatic rebuild on file changes
- Supports Go files, templates, HTML, and environment files
- Excludes common directories like assets, tmp, vendor, and node_modules
- Clean process management
- Configurable through `.air.toml`

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
