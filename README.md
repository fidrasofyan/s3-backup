# db-backup

A simple command-line tool for backing up databases to S3-compatible storage (AWS S3, Cloudflare R2, DigitalOcean Spaces, etc.).

## Features

- **Database Backup**: Support for multiple MySQL/MariaDB database backups.
- **Compression**: Automatic Gzip compression for database dumps.
- **Retention Policy**: Automatically delete old backups to save storage space.
- **YAML Configuration**: Easy to configure with a single file.

## Prerequisites

For database backups, ensure one of the following is installed in your system:

- `mysqldump`
- `mariadb-dump`

## Installation

### Download binary

Download the pre-compiled binary for your platform from the [Releases](https://github.com/fidrasofyan/db-backup/releases/latest) page.

### Build from source

```sh
make build
```

The binary will be available at `bin/db-backup`.

## Usage

### Initialize Configuration

Generate a default `config.yaml` template:

```sh
./bin/db-backup init
```

### Backup Databases & Upload

Backup all configured databases, rotate old ones, and upload to S3:

```sh
# Basic usage
./bin/db-backup backup-db --config config.yaml

# Keep only the last 5 backups
./bin/db-backup backup-db --config config.yaml --keep 5

# Create local backup only (no upload)
./bin/db-backup backup-db --config config.yaml --no-upload
```

## Configuration

Edit `config.yaml` to match your environment:

```yaml
aws:
  endpoint: https://your-r2-or-s3-endpoint.com
  region: auto
  access_key_id: your-access-key
  secret_access_key: your-secret-key
  bucket: your-bucket-name

backup_db:
  - type: mariadb
    host: 127.0.0.1
    port: 3306
    user: db_user
    password: db_password
    dbname: first_database
  - type: mariadb
    host: 127.0.0.1
    port: 3306
    user: db_user
    password: db_password
    dbname: second_database

local_dir: ./backup
remote_dir: production/backups
```

### Configuration Fields

| Field          | Description                                               |
| -------------- | --------------------------------------------------------- |
| `aws.endpoint` | S3-compatible API endpoint (e.g., Cloudflare R2 endpoint) |
| `aws.region`   | AWS Region (use `auto` for R2)                            |
| `backup_db`    | List of databases to backup                               |
| `local_dir`    | Local path where database dumps are stored before upload  |
| `remote_dir`   | Destination path in your S3 bucket                        |

## License

This project is licensed under the [MIT License](LICENSE).
