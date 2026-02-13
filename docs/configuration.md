# Configuration Guide - Database Schema Support

## Overview

PostgreSQL supports multiple schemas within a single database. Previously, EVO hardcoded the schema name to 'public'. Now you can configure the schema name via settings.

## Configuration

Add the SCHEMA field to your database configuration in config.yml:

```yaml
DATABASE:
  ENABLED: true
  TYPE: postgres
  SERVER: localhost:5432
  USERNAME: myuser
  PASSWORD: mypass
  DATABASE: mydb
  SCHEMA: public  # NEW: PostgreSQL schema name (defaults to 'public')
  SSLMODE: disable
```

## Example: Multi-tenant Setup

```yaml
# Tenant A  
DATABASE:
  SCHEMA: tenant_a

# Tenant B
DATABASE:
  SCHEMA: tenant_b
```

## Usage

The framework automatically uses the configured schema for all database operations:

```go
// No code changes needed!
db.AutoMigrate(&User{})  // Creates table in configured schema
db.Find(&users)          // Queries configured schema
```

## MySQL Note

MySQL does not support schemas. In MySQL, the SCHEMA setting is ignored.

## Support

For questions: https://github.com/getevo/evo/issues
