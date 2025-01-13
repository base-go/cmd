# Base - Command Line Tool for the Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework.
It offers scaffolding, module generation, and utilities to accelerate Go application development.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Commands](#commands)
  - [`base new`](#base-new)
  - [`base generate` or `base g`](#base-generate-or-base-g)
  - [`base destroy` or `base d`](#base-destroy-or-base-d)
  - [`base start` or `base s`](#base-start-or-base-s)
  - [`base update`](#base-update)
  - [`base upgrade`](#base-upgrade)
  - [`base version`](#base-version)
- [Examples](#examples)
  - [Generating a New Project](#generating-a-new-project)
  - [Generating Modules](#generating-modules)
  - [Working with Image Uploads](#working-with-image-uploads)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

You can install the Base CLI tool using one of the following methods:

1. **Using the install script**:
   ```bash
   curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | bash
   ```

## Getting Started

Verify your installation by running:

```bash
base --help
```

This displays the help menu with all available commands and options.

---

## Commands

### `base new`

Create a new project using the Base framework.

**Usage**:
```bash
base new <project-name>
```

**Example**:
```bash
base new myapp
```

---

### `base generate` or `base g`

Generate a new module with specified fields and options.

**Usage**:
```bash
base g <module-name> [field:type ...] [options]
```

- `<module-name>`: Name of the module (e.g., `User`, `Post`)
- `[field:type ...]`: List of fields with types
- `[options]`: Additional flags, such as `--admin` for generating an admin interface

**Supported Field Types**:
- **Primitive Types**: `string`, `text`, `int`, `bool`, `float`, `time`
- **Relationships**: `belongsTo`, `hasOne`, `hasMany`, `attachment`

**Example**:
```bash
base g User name:string email:string posts:hasMany:Post
```

---

### `base destroy` or `base d`

Destroy a module and its associated files.

**Usage**:
```bash
base d <module-name>
```

**Example**:
```bash
base d User
```

---

### `base start` or `base s`

Start the development server.

**Usage**:
```bash
base s
```

---

### `base update`

Update the Base framework's core components in your project. This command:
- Updates the core directory with the latest version
- Maintains your custom modifications
- Updates core interfaces and utilities
- Preserves your application code

**Usage**:
```bash
base update
```

**What Gets Updated**:
- Core interfaces and types
- Base utilities and helpers
- Storage system components
- Authentication system
- Database utilities
- Event system
- Logging system
- Middleware components
- Error handling
- Configuration management
- Testing utilities

**Example Output**:
```bash
$ base update
Updating Base framework components...
✓ Backing up current core directory
✓ Downloading latest core components
✓ Updating interfaces and types
✓ Updating utilities and helpers
✓ Preserving custom modifications
✓ Cleaning up temporary files
Update completed successfully!
```

**Note**: This command only updates the framework's core components. To update the CLI tool itself, use `base upgrade`.

---

### `base upgrade`

Upgrade the Base CLI tool to the latest version.

**Usage**:
```bash
base upgrade
```

---

### `base version`

Display version information for the Base CLI tool.

**Usage**:
```bash
base version
```

**Example Output**:
```bash
Base CLI v1.0.0
Commit: abc123d
Built: 2025-01-13 00:55:29
Go version: go1.21.0
```

---

## Examples

### Generating a New Project

```bash
# Create a new project named 'blog'
base new blog

# Change into the project directory
cd blog

# Start the development server
base s
```

### Generating Modules

Base provides a powerful module generation system that supports various field types and relationships.

#### Basic Module with Simple Fields
```bash
# Generate a basic module with string and text fields
base g Post title:string content:text

# Generate a module with various field types
base g Product \
  name:string \
  description:text \
  price:float \
  quantity:int \
  is_active:bool \
  published_at:time
```

#### Modules with File Attachments
```bash
# Generate a module with image support
base g Profile \
  name:string \
  bio:text \
  avatar:attachment

# Multiple attachments in one module
base g Gallery \
  title:string \
  description:text \
  cover:attachment \
  image:attachment
```

#### Modules with Relationships
```bash
# One-to-Many Relationship (Category has many Posts)
base g Category \
  title:string \
  content:text \
  image:attachment \
  posts:hasMany:Post

# Belongs-To Relationship (Post belongs to Category)
base g Post \
  title:string \
  content:text \
  image:attachment \
  category:belongsTo:Category

# One-to-One Relationship
base g User \
  name:string \
  email:string \
  profile:hasOne:Profile

# Multiple Relationships
base g Comment \
  title:string \
  content:text \
  user:belongsTo:User \
  post:belongsTo:Post \
  replies:hasMany:Comment
```

#### Complex Module Example
```bash
# Blog system with all features
base g User \
  username:string \
  email:string \
  password:string \
  avatar:attachment \
  profile:hasOne:Profile \
  posts:hasMany:Post \
  comments:hasMany:Comment

base g Profile \
  bio:text \
  website:string \
  avatar:attachment \
  social_links:text \
  user:belongsTo:User

base g Category \
  name:string \
  description:text \
  image:attachment \
  parent:belongsTo:Category \
  subcategories:hasMany:Category \
  posts:hasMany:Post

base g Post \
  title:string \
  content:text \
  excerpt:text \
  featured_image:attachment \
  gallery:attachment \
  published_at:time \
  author:belongsTo:User \
  category:belongsTo:Category \
  tags:hasMany:Tag \
  comments:hasMany:Comment

base g Comment \
  content:text \
  author:belongsTo:User \
  post:belongsTo:Post \
  parent:belongsTo:Comment \
  replies:hasMany:Comment
```

Each generated module includes:
- Model with GORM configuration
- Service layer with CRUD operations
- Controller with REST endpoints
- Automatic migrations
- Search functionality
- Pagination
- File upload endpoints (for attachment fields)
- Relationship handling
- Type-safe request/response structs

The generated code follows best practices:
- Clean architecture principles
- Dependency injection
- Interface-based design
- Proper error handling
- Input validation
- Secure file handling
- Efficient database queries
- Proper relationship loading

### Working with Image Uploads

Base provides a flexible storage system for handling file uploads. You can use different storage providers:

#### 1. Local Storage (Default)
Files are stored in your local filesystem.

```bash
# Generate a module with image support
base g Profile name:string bio:text avatar:attachment

# Configuration in config.yaml
storage:
  provider: local
  path: "./storage"
  baseURL: "http://localhost:8080/storage"
```

#### 2. S3 Compatible Storage
Store files in AWS S3 or any S3-compatible service (like MinIO, DigitalOcean Spaces).

```yaml
# Configuration in config.yaml
storage:
  provider: s3
  bucket: "my-bucket"
  region: "us-east-1"
  endpoint: "https://s3.amazonaws.com"
  baseURL: "https://my-bucket.s3.amazonaws.com"
  apiKey: "your-access-key"
  apiSecret: "your-secret-key"
```

#### 3. Cloudflare R2
Store files in Cloudflare R2 with optional CDN support.

```yaml
# Configuration in config.yaml
storage:
  provider: r2
  bucket: "my-bucket"
  endpoint: "https://<account-id>.r2.cloudflarestorage.com"
  baseURL: "https://cdn.example.com"  # If using Cloudflare CDN
  apiKey: "your-access-key"
  apiSecret: "your-secret-key"
```

#### Usage in API
Once configured, the upload endpoints are automatically available:

```bash
# Upload an image
curl -X POST -F "avatar=@image.jpg" http://localhost:8080/api/profiles/1/upload/avatar

# The response includes the file URL
{
  "url": "http://localhost:8080/storage/profiles/1/avatar/image.jpg"
}
```

The storage system automatically:
- Validates file types (default: jpg, jpeg, png, gif)
- Handles file size limits (default: 10MB)
- Generates unique filenames
- Creates optimized versions for images
- Cleans up old files when updated
- Provides secure URLs for access

## Contributing

We welcome contributions to Base! Here's how you can help:

1. Fork the repository.
2. Create a branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
