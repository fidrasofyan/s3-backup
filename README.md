# s3-backup

A command-line tool for backing up directories to Amazon S3-compatible APIs.

## Features

- Upload local directories to S3-compatible storage (e.g., AWS S3, Cloudflare R2)
- Backup database and upload to S3-compatible storage
- Simple YAML configuration

## Installation

### Download binary

Download the binary from [Releases](https://github.com/fidrasofyan/s3-backup/releases/latest)

### Build from source

```sh
make build
```
The binary will be available at `bin/s3-backup`.

## Usage

Initialize a config file:
```sh
./bin/s3-backup init
```

Upload files:
```sh
./bin/s3-backup upload --config config.yaml
```

Backup database:
```sh
./bin/s3-backup backup-db --config config.yaml
```

Use `--help` for more options.

## Configuration

Edit `config.yaml`:
```yaml
aws:
  endpoint: https://your-s3-endpoint
  region: auto
  access_key_id: your-access-key
  secret_access_key: your-secret-key
  bucket: your-bucket
backup_db:
  - type: mysql
    host: 127.0.0.1
    port: 3306
    user: user
    password: password
    dbname: database_name
local_dir: /path/to/local/backup
remote_dir: remote/path/in/bucket
```

## License

MIT