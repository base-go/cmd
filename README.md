# Base - Command Line Tool for the Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework.
It offers scaffolding, module generation, and utilities to accelerate Go application development.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Commands](#commands)
  - [`base new`](#base-new)
  - [`base g`](#base-generate-or-base-g)
  - [`base d`](#base-destroy-or-base-d)
  - [`base start`](#base-start-or-base-s)
- [Blog Example](#blog-example)
- [Contributing](#contributing)
- [License](#license)

## Installation

You can install the Base CLI tool using one of the following methods:

1. **Using the install script** (Recommended):
   ```bash
   curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | bash
   ```

2. **From Source**:
   ```bash
   git clone https://github.com/base-go/cmd.git
   cd cmd
   go build -o base
   sudo mv base /usr/local/bin/
   ```

## Getting Started

Verify your installation by running:

```bash
base --help
```

This displays the help menu with all available commands and options.

## Commands

### `base new`

Create a new project using the Base framework.

**Usage**:
```bash
base new <project-name>
```

**Example**:
```bash
base new myblog
cd myblog
go mod tidy
```

### `base generate` or `base g`

Generate a new module with specified fields and types.

**Usage**:
```bash
base g <module-name> [field:type ...] [options]
```

**Supported Field Types**:
- **Basic Types**: 
  - `string`: For short text
  - `text`: For long text content
  - `int`: For numbers
  - `float`: For decimal numbers
  - `bool`: For true/false values
  - `time`: For dates and timestamps
  - `file`: For file uploads
  - `image`: For image uploads
  
- **Relationship Types**:
  - `belongsTo`: One-to-one relationship (child side)
  - `hasOne`: One-to-one relationship (parent side)
  - `hasMany`: One-to-many relationship
  - `sort`: For sortable records

### `base destroy` or `base d`

Remove a module and its associated files.

**Usage**:
```bash
base d <module-name>
```

### `base start` or `base s`

Start the development server.

**Usage**:
```bash
base s
```

## Blog Example

Let's create a complete blog system with users, posts, categories, and comments.

### 1. Create a New Project

```bash
base new myblog
cd myblog
```

### 2. Generate the User Module

```bash
base g user name:string email:string password:string bio:text avatar:image
```

This creates:
- User model with name, email, password, bio fields
- File upload handling for avatar
- CRUD API endpoints
- Service layer with search functionality

### 3. Generate the Category Module

```bash
base g category name:string description:text sort:sort
```

Features:
- Sortable categories
- Full CRUD operations
- Search by name and description

### 4. Generate the Post Module

```bash
base g post title:string content:text published_at:time featured_image:image author:belongsTo:User category:belongsTo:Category
```

Creates:
- Post model with relationships to User and Category
- Image upload handling for featured_image
- Timestamps for publishing
- Full text search
- CRUD operations with relationship handling

### 5. Generate the Comment Module

```bash
base g comment content:text user:belongsTo:User post:belongsTo:Post
```

Features:
- Comments linked to both users and posts
- CRUD operations with relationship validations
- Nested relationship handling

### 6. Test the Generated API

Start the server:
```bash
base s
```

Example API calls:

```bash
# Create a category
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Technology", "description": "Tech articles"}'

# Create a post
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "First Post",
    "content": "Hello World!",
    "author_id": 1,
    "category_id": 1
  }'

# Get all posts with pagination and search
curl "http://localhost:8080/api/posts?page=1&limit=10&search=technology"
```

### 7. Clean Up (Optional)

To remove a module:

```bash
base d post    # Removes the post module
base d comment # Removes the comment module
```

## Contributing

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.
