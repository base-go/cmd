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

We also provide a `replant` command to reset the database and seed fresh data:

```bash
base replant
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

# Base Feed Command

The `base feed` command is a flexible tool for importing JSON data into your database tables. It offers various ways to map JSON fields to database columns.

## Basic Syntax

```
base feed <table_name>[:<json_path>] [field_mappings...]
```

- `<table_name>`: The name of the database table to insert data into.
- `<json_path>` (optional): The path to the JSON file. If omitted, it defaults to `data/<table_name>.json`.
- `[field_mappings...]` (optional): Specifications for how to map JSON fields to database columns.

## Usage Examples

1. **Basic usage (no mappings)**

   ```
   base feed users
   ```
   This will read from `data/users.json` and insert all fields into the `users` table as-is.

2. **Specifying a custom JSON file**

   ```
   base feed users:custom_data/my_users.json
   ```
   This will read from `custom_data/my_users.json` and insert all fields into the `users` table as-is.

3. **Simple field mapping**

   ```
   base feed users name:full_name email:user_email
   ```
   This will map the "name" field from JSON to the "full_name" column, and "email" to "user_email". Other fields will be inserted as-is.

4. **Multiple mappings from one source**

   ```
   base feed users username:full_name username:login_name
   ```
   This will map the "username" field from JSON to both "full_name" and "login_name" columns in the database.

5. **Concatenating multiple fields**

   ```
   base feed users "first_name last_name":full_name email:contact_email
   ```
   This will concatenate "first_name" and "last_name" from JSON (with a space between) and map to the "full_name" column. It will also map "email" to "contact_email".

6. **Combining all types of mappings**

   ```
   base feed users "first_name last_name":full_name username:login username:display_name email:contact_email
   ```
   This command demonstrates all types of mappings:
   - Concatenation: "first_name" and "last_name" to "full_name"
   - Multiple targets: "username" to both "login" and "display_name"
   - Simple mapping: "email" to "contact_email"

## Behavior Notes

- Unmapped fields: Any JSON fields not explicitly mapped will be inserted using their original field names.
- Missing fields: If a JSON field specified in the mapping doesn't exist, it's silently ignored.
- Data types: The command attempts to preserve the data types from JSON. Ensure your database schema can handle the incoming data types.
- Concatenation: When concatenating fields, only string values are used. Non-string values in concatenation are silently ignored.

## Error Handling

- Invalid JSON: If the JSON file can't be parsed, an error is displayed and the operation is aborted.
- Database errors: Any errors during database insertion are displayed, but the operation continues with the next record.
- Invalid mappings: Mappings not in the format `source:target` are ignored with a warning message.

This flexible command allows for a wide range of data import scenarios, from simple one-to-one mappings to complex field concatenations and duplications.
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
