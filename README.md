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
- [Base Blog Modules](#base-blog-modules)
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

Generate a new module with specified fields and relationships.

**Usage**:
```bash
base g <module-name> [field:type ...] [options]
```

**Field Types**:
- **Basic Types**:
  - `string`: String field
  - `text`: Text field (for longer content)
  - `int`: Integer field
  - `float`: Float field
  - `bool`: Boolean field
  - `datetime`: DateTime field (for date and time)

- **Relationships**:
  - `belongsTo`: Many-to-one relationship
  - `hasOne`: One-to-one relationship
  - `hasMany`: One-to-many relationship
  
- **Special Types**:
  - `attachment`: File attachment (handled by storage system)

**Example - Blog System**:
```bash
# Generate User model with relationships
base g User \
  username:string \
  email:string \
  password:string \
  avatar:attachment \
  profile:hasOne:Profile \
  posts:hasMany:Post \
  comments:hasMany:Comment

# Generate Profile model
base g Profile \
  bio:text \
  website:string \
  avatar:attachment \
  social_links:text \
  user:belongsTo:User

# Generate Category model with self-referential relationships
base g Category \
  name:string \
  description:text \
  image:attachment \
  parent:belongsTo:Category \
  subcategories:hasMany:Category \
  posts:hasMany:Post

# Generate Post model with multiple relationships
base g Post \
  title:string \
  content:text \
  excerpt:text \
  featured_image:attachment \
  gallery:attachment \
  published_at:datetime \
  author:belongsTo:User \
  category:belongsTo:Category \
  tags:hasMany:Tag \
  comments:hasMany:Comment

# Generate Tag model
base g Tag \
  name:string \
  posts:hasMany:Post

# Generate Comment model with self-referential relationships
base g Comment \
  content:text \
  author:belongsTo:User \
  post:belongsTo:Post \
  parent:belongsTo:Comment \
  replies:hasMany:Comment
```

This will generate:
- Models with proper GORM tags and relationships
- Services with CRUD operations
- Controllers with RESTful endpoints
- Response/Request structs
- Proper handling of circular dependencies
- File upload handling for attachments
- Automatic preloading of relationships

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

Update the Base framework's core components in your project.

**Usage**:
```bash
base update
```

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

# Generate the blog system models
base g User username:string email:string password:string avatar:attachment
base g Post title:string content:text author:belongsTo:User
base g Tag name:string posts:hasMany:Post

# Start the development server
base start
```

### Working with Image Uploads

The `attachment` type automatically handles file uploads:

```bash
base g Product \
  name:string \
  description:text \
  image:attachment \
  gallery:attachment
```

This generates:
- File upload handling in controllers
- Storage system integration
- Image processing capabilities
- Proper JSON serialization

---

## Base Blog Modules

The Base framework comes with a set of pre-built modules designed for building blog applications:

### Core Modules

1. **Post Module**
   - Full CRUD operations for blog posts
   - Support for rich text content
   - Image attachments
   - Relationships with categories, tags, and comments

2. **Category Module**
   - Hierarchical category structure
   - Parent-child relationships
   - Post categorization
   - Description and image support

3. **Comment Module**
   - Nested comment structure
   - Parent-child relationships for replies
   - Author and post associations
   - Content management

4. **User Module**
   - User authentication and authorization
   - Profile management
   - Role-based access control
   - Activity tracking

5. **Profile Module**
   - Extended user information
   - Avatar management
   - Social media links
   - Bio and personal details

6. **Tag Module**
   - Post tagging system
   - Tag management
   - Tag-based post filtering
   - Tag cloud generation

### Module Generation

To generate these modules in your project, use:

```bash
# Generate a single module
base g module post
base g module category
base g module comment
base g module user
base g module profile
base g module tag

# Generate all blog modules at once
base g blog
```

### Module Removal

To remove modules from your project:

```bash
# Remove a module
base d post
base d category
base d comment
base d user
base d profile
base d tag
```

Each module includes:
- Models with proper relationships
- Controllers with RESTful endpoints
- Services with business logic
- Migration support
- Swagger documentation
- Test templates

---

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
