# Go Echo PostgreSQL API

A production-ready Go API with Echo, PostgreSQL, JWT authentication, and modern development tools.

## Features

- 🚀 **Echo** - High performance HTTP framework
- 🗄️ **PostgreSQL** - Database with connection pooling
- 🔐 **JWT Authentication** - Secure token-based auth
- ⚙️ **Viper** - Configuration management
- 📝 **Structured Logging** - Zap logger
- 🛡️ **CORS & Security** - Middleware protection
- 🔄 **Hot Reload** - Air for development
- 🧪 **SQLC** - Type-safe SQL queries
- 📦 **Taskfile** - Build automation
- 🐳 **Docker** - Containerization
- 🔄 **Migrations** - Database schema management

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Task (optional)

### Setup

1. **Clone and install dependencies:**

   ```bash
   git clone <repository>
   cd go-echo-pg
   go mod download
   ```

   ## 🚀 **How to Use This Setup**

1. **Install tools:**
   ```bash
   go install github.com/air-verse/air@latest
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   go install github.com/go-task/task/v3/cmd/task@latest
   ```
