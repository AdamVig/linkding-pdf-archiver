# Linkding PDF Archiver

> [!NOTE]
> Linkding now has [native support for capturing PDFs](https://github.com/sissbruecker/linkding/pull/1271), so this tool is no longer necessary!

> **Note:** This is a fork of [proog/linkding-media-archiver](https://github.com/proog/linkding-media-archiver) with additional modifications by [AdamVig](https://github.com/AdamVig).

Automatically download PDFs for your [Linkding](https://linkding.link/) bookmarks

## What it is

Linkding can automatically create HTML snapshots of your bookmarks to guard against link rot. Linkding PDF Archiver supplements this feature by automatically downloading PDF files for your bookmarks and adding them to Linkding as additional assets.

## How it works

Linkding PDF Archiver retrieves bookmarks that do not already have a PDF file attached and attempts to download one. If successful, the file is uploaded to Linkding as a bookmark asset. This process repeats on a configurable schedule with any bookmarks that have been added or changed since the previous run. If Linkding PDF Archiver is restarted, it will retrieve all bookmarks again.

## Usage

Linkding PDF Archiver requires a [Linkding](https://linkding.link/) instance to work. The easiest way to run it is by using [the Docker image](https://github.com/AdamVig/linkding-pdf-archiver/pkgs/container/linkding-pdf-archiver). See `docker-compose.example.yml` for an example Docker Compose setup that combines Linkding and Linkding PDF Archiver. Alternatively, it can be run as a binary by cloning the repository and compiling from source.

```sh
# Docker Compose (preferred, see docker-compose.example.yml)
docker compose up

# Docker
docker run --rm -e LDPA_BASEURL="http://localhost:9090" -e LDPA_TOKEN="abcd1234" AdamVig/linkding-pdf-archiver [-n] [-s]

# Binary
go build -o ./linkding-pdf-archiver ./cmd
LDPA_BASEURL="http://localhost:9090" LDPA_TOKEN="abcd1234" ./linkding-pdf-archiver [-n] [-s]
```

### Flags

- `-n` Dry run: download PDFs but do not actually upload it to Linkding
- `-s` Single run: exit after processing bookmarks once

### Environment variables

| Name                 | Example                            | Default                | Description                                                                                                                                         |
| -------------------- | ---------------------------------- | ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `LDPA_BASEURL`       | `http://linkding.example.com:9090` | None **(required)**    | Base URL of your Linkding instance                                                                                                                  |
| `LDPA_TOKEN`         | `{random 40 char token}`           | None **(required)**    | Auth token from the Linkding integration page                                                                                                       |
| `LDPA_TAGS`          | `to-archive pdf`              | None (all bookmarks)   | Only process bookmarks with any of these tags (space separated, omit the #)                                                                         |
| `LDPA_BUNDLE_ID`     | `42`                               | None (all bookmarks)   | Only process bookmarks matching this [bundle](https://github.com/sissbruecker/linkding/pull/1097) (get the id from the url when editing the bundle) |
| `LDPA_SCAN_INTERVAL` | `600` (10 mins)                    | `3600` (1 hour)        | Schedule to check for new bookmarks                                                                                                                 |
| `LDPA_LOG_LEVEL`     | `DEBUG`                            | `INFO`                 | Log level, useful for troubleshooting                                                                                                               |
| `LDPA_LOG_FORMAT`    | `text`                             | `text` when running in a terminal, `json` otherwise | Log output format, allows you to choose between human-readable or machine-readable logs |