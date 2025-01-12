
# Base - Command Line Tool for the Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework.
It offers scaffolding, module generation, and utilities to accelerate Go application development.
You can to seed data, import JSON files, and more with a few simple commands.

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
- **Relationships**: `belongsTo`, `hasOne`, `hasMany`

**Example**:
```bash
base g User name:string email:string password:string profile:hasOne:Profile
```

### `base destroy` or `base d`

Destroy a module and its associated files.
---

### `base start` or `base s`

Start the development server.

**Usage**:
```bash
base s
```

---

### `base update`

Update the Base Core package to the latest version.

**Usage**:
```bash
base update
```

### `base upgrade`

Upgrade the Base CLI tool to the latest version.

---

## Examples

### Generating a New Project

Create a new project called `myapp`:

```bash
base new myapp
cd myapp
go mod tidy
```

---

### Generating Modules

#### Blog System Example:

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

---

### Seeding Data

Base CLI automatically generates seed files for each module. To seed your database with initial data, use:

```bash
base seed
```

To reset and seed fresh data:

```bash
base replant
```

**Important Note on Seeding Relationships**:
Ensure parent records are seeded before child records. Adjust the order in `app/seed.go` accordingly.

Example:

```go
func InitializeSeeders() []module.Seeder {
    return []module.Seeder{
        &user.UserSeeder{},        // Parent
        &category.CategorySeeder{},// Independent
        &post.PostSeeder{},        // Child of User
        &comment.CommentSeeder{},  // Child of User and Post
    }
}
```

---
### Feeding Data
Base CLI provides a flexible way to import JSON data into your database. You can map JSON fields to database columns using the `base feed` command.

## Base Feed Command

The `base feed` command imports JSON data into your database with flexible field mapping options.

### Basic Syntax

```bash
base feed <table_name>[:<json_path>] [field_mappings...]
```

- `<table_name>`: Database table to insert data into.
- `<json_path>` (optional): Path to the JSON file.
- `[field_mappings...]` (optional): Mappings for JSON fields to database columns.

### Usage Examples

1. **Basic usage**:
   ```bash
   base feed users
   ```

2. **Using a custom JSON file**:
   ```bash
   base feed users:custom_data/my_users.json
   ```

3. **Simple field mapping**:
   ```bash
   base feed users name:full_name email:user_email
   ```

4. **Mapping one source to multiple columns**:
   ```bash
   base feed users username:full_name username:login_name
   ```

5. **Concatenating multiple fields**:
   ```bash
   base feed users "first_name last_name":full_name email:contact_email
   ```

6. **Combining all types of mappings**:
   ```bash
   base feed users "first_name last_name":full_name username:login username:display_name email:contact_email
   ```

---

## Contributing

Contributions are welcome! Follow these steps:

1. Fork the repository.
2. Create a branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a pull request.

To report issues, use the [GitHub Issues](https://github.com/base-go/cmd/issues) page, and provide detailed information to help us address the issue promptly.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

© 2024 Basecode LLC. All rights reserved.

For more information on the Base framework, refer to the official documentation.
