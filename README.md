# Base - Command Line Tool for Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework. It provides scaffolding, module generation, and utilities to accelerate your Go application development.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Commands](#commands)
  - [`base new`](#base-new)
  - [`base g`](#base-generate-or-base-g)
  - [`base start` or `base s`](#base-start-or-base-s)
  - [`base update`](#base-update)
- [Examples](#examples)
  - [Generating a New Project](#generating-a-new-project)
  - [Generating Modules](#generating-modules)
  - [Seeding Data](#seeding-data)
- [Contributing](#contributing)
- [License](#license)

## Installation

Install the Base CLI tool by running:

```bash
curl -sSL https://raw.githubusercontent.com/base-go/cmd/main/install.sh | bash
```

This script downloads and installs the latest version of the Base CLI.

Alternatively, install via `go install`:

```bash
go install github.com/base-go/cmd@latest
```

## Getting Started

After installation, verify the CLI is working:

```bash
base --help
```

This displays the help menu with all available commands and options.

## Commands

### `base new`

Creates a new Base framework project.

**Usage:**

```bash
base new <project-name>
```

**Example:**

```bash
base new myapp
```

### `base generate` or `base g`

Generates a new module with the specified name and fields.

**Usage:**

```bash
base g <module-name> [field:type ...] [options]
```

**Parameters:**

- `<module-name>`: Name of the module (e.g., `User`, `Post`).
- `[field:type ...]`: List of fields with types.
- `[options]`: Additional flags like `--admin` to generate admin interface.

**Supported Field Types:**

- **Primitive Types:**
  - `string`
  - `text`
  - `int`
  - `bool`
  - `float`
  - `time`
- **Relationships:**
  - `belongsTo`: `user:belongsTo:User`
  - `hasOne`: `profile:hasOne:Profile`
  - `hasMany`: `posts:hasMany:Post`

**Example:**

```bash
base g User name:string email:string password:string profile:hasOne:Profile
```

### `base start` or `base s`

Starts the development server.

**Usage:**

```bash
base s
```

### `base update`

Updates the Base CLI to the latest version.

**Usage:**

```bash
base update
```

## Examples

### Generating a New Project

Create a new project called `myapp`:

```bash
base new myapp
cd myapp
go mod tidy
```

### Generating Modules

#### Example: Blog System

Let's create a simple blog system with users, posts, and comments:

```bash
# Generate User module
base g User name:string email:string password:string

# Generate Post module
base g Post title:string content:text published_at:time author:belongsTo:User

# Generate Comment module
base g Comment content:text user:belongsTo:User post:belongsTo:Post

# Generate Category module with admin interface
base g Category name:string description:text --admin
```

### Seeding Data

The Base CLI automatically generates seed files for each module. To seed your database with initial data, use the following command:

```bash
base seed
```

**Important Note on Seeding Relationships:**

When seeding data for modules with relationships, ensure that the parent records exist before seeding the child records. This should be reflected in the order of seeders in your `app/seed.go` file.

Example of correct seeder initialization order in `app/seed.go`:

```go
func InitializeSeeders() []module.Seeder {
    return []module.Seeder{
        &user.UserSeeder{},        // Parent
        &category.CategorySeeder{},// Independent
        &post.PostSeeder{},        // Child of User
        &comment.CommentSeeder{},  // Child of User and Post
        // Add other seeders in the correct order
        // SEEDER_INITIALIZER_MARKER - Do not remove this comment
    }
}
```

This order ensures that:
1. Users are seeded first (parent)
2. Categories are seeded (independent)
3. Posts are seeded (child of User)
4. Comments are seeded last (child of both User and Post)

When customizing seed data, maintain this order to avoid foreign key constraint violations.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the Repository
2. Create a Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

**Reporting Issues:**

- Use the [GitHub Issues](https://github.com/base-go/cmd/issues) page to report bugs or request features.
- Provide detailed information to help us understand and address the issue promptly.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

© 2024 Basecode LLC. All rights reserved.

For more detailed information on the Base framework and its capabilities, please refer to the official documentation.