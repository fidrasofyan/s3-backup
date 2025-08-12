# s3backup

A command-line tool for backing up directories to Amazon S3-compatible APIs.

## Features

- Upload local directories to S3-compatible storage (e.g., AWS S3, Cloudflare R2)
- Simple YAML configuration

## Installation

```sh
make build
```
The binary will be available at `bin/s3backup`.

## Usage

Initialize a config file:
```sh
./bin/s3backup init
```

Upload files:
```sh
./bin/s3backup upload --config config.yaml
```

## Configuration

Edit `config.yaml`:
```yaml
aws:
	endpoint: https://your-s3-endpoint
	region: auto
	access_key_id: your-access-key
	secret_access_key: your-secret-key
	bucket: your-bucket
local_dir: /path/to/local/backup
remote_dir: remote/path/in/bucket
```

## Example

```sh
./bin/s3backup upload --config config.yaml
```

## License

MIT