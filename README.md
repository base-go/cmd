
# Base - Command Line Tool for Base Framework

Base is a powerful command-line tool designed to streamline development with the Base framework. It provides scaffolding, module generation, and utilities to accelerate your Go application development.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Commands](#commands)
  - [`base new`](#base-new)
  - [`base generate` or `base g`](#base-generate-or-base-g)
  - [`base server` or `base s`](#base-server-or-base-s)
  - [`base update`](#base-update)
- [Examples](#examples)
  - [Generating a New Project](#generating-a-new-project)
  - [Generating Modules](#generating-modules)
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
base new <project-name> [options]
```

**Options:**

- `--module`: Specify the Go module path (e.g., `--module=github.com/username/project`).

**Example:**

```bash
base new myapp --module=github.com/username/myapp
```

---

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

---

### `base server` or `base s`

Starts the development server.

**Usage:**

```bash
base s [options]
```

**Options:**

- `--port`: Specify server port (default is 8080).

**Example:**

```bash
base s --port=3000
```

---

### `base update`

Updates the Base CLI to the latest version.

**Usage:**

```bash
base update
```

---

## Examples

### Generating a New Project

Create a new project called `myapp`:

```bash
base new myapp --module=github.com/yourusername/myapp
cd myapp
go mod tidy
```

---

### Generating Modules

#### Example 1: User Module with Relationships

Generate a `User` module with various fields and relationships:

```bash
base g User   name:string   email:string   password:string   is_active:bool   last_login:time   profile:hasOne:Profile   settings:hasOne:UserSettings
```

---

#### Example 2: Post Module with Relationships

Generate a `Post` module that belongs to `User` and `Category`, and has a `hasOne` relationship with `Image`:

```bash
base g Post   title:string   content:text   published:bool   view_count:int   published_at:time   author:belongsTo:User   category:belongsTo:Category   featured_image:hasOne:Image
```

---

#### Example 3: Comment Module

Generate a `Comment` module that belongs to `User` and `Post`:

```bash
base g Comment   content:text   approved:bool   user:belongsTo:User   post:belongsTo:Post
```

---

#### Example 4: Generating Admin Interface

Generate a `Category` module with an admin interface:

```bash
base g Category name:string description:text --admin
```

---

#### Example 5: Complex Relationships

Generate `NewsletterSubscription` module:

```bash
base g NewsletterSubscription   user:belongsTo:User   newsletter:belongsTo:Newsletter
```

---

#### Example 6: Starting the Server

Start the development server on default port 8080:

```bash
base s
```

Start the server on port 3000:

```bash
base s --port=3000
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the Repository**: Click the "Fork" button at the top right of the [repository page](https://github.com/base-go/cmd).
2. **Create a Branch**: Create a new branch for your feature or bugfix.
   ```bash
   git checkout -b feature/new-feature
   ```
3. **Make Changes**: Implement your feature or fix the bug.
4. **Commit Changes**: Commit your changes with clear and concise messages.
   ```bash
   git commit -am "Add new feature"
   ```
5. **Push to Fork**: Push your changes to your forked repository.
   ```bash
   git push origin feature/new-feature
   ```
6. **Submit Pull Request**: Go to the original repository and submit a Pull Request.

**Reporting Issues:**

- Use the [GitHub Issues](https://github.com/base-go/cmd/issues) page to report bugs or request features.
- Provide detailed information to help us understand and address the issue promptly.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

© 2024 Basecode LLC. All rights reserved.

# Base Framework Documentation

For more detailed information on the Base framework and its capabilities, please refer to the official documentation.

---

Thank you for using the Base CLI! If you have any questions or need further assistance, feel free to reach out to the community or maintainers.
