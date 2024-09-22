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
