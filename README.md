# URL Shortener

A fast, self-hosted URL shortener built with Go and SQLite.

Links use short Base62 IDs by default, or you can provide your own slug:

```text
https://example.com/a8K2xPq1
https://example.com/my-custom-link
```

Runs as a standalone binary or Docker container.

## Features

- Base62 short links.
- Custom slugs, including nested paths.
- In-memory LRU cache.
- Asynchronous click analytics.
- SQLite with WAL mode for concurrent access.
- Single binary with no external dependencies.

## Getting started

### Run with Docker

- Make [`docker-compose.yml`](./docker-compose.yml)
- Copy [`.env.example`](./.env.example) to `.env`. (Optional)

Then start the service:

```sh
docker compose up -d
```

### Run binary

You can either download a release or build the binary yourself.

1. Download or build
   - #### Download a release

     Download the latest release for your platform from [GitHub Releases](https://github.com/huffmanks/url-shortener/releases/latest).

   - #### Build from source

     Clone the repository:

     ```sh
     git clone https://github.com/huffmanks/url-shortener.git
     ```

     Build the binary:

     ```sh
     go build -o url-shortener .
     ```

2. Make the binary executable:

```sh
chmod +x url-shortener-*
```

3. Move it somewhere in your `PATH`:

```sh
sudo mv url-shortener-* /usr/local/bin/url-shortener
# OR
# mv url-shortener-* ~/.local/bin/url-shortener
```

4. Start the server:

```sh
url-shortener
```

By default, the database is stored in the system’s application data directory:

- macOS: `~/Library/Application Support/url-shortener/`
- Linux: `~/.config/url-shortener/`

#### Configuration

Set environment variables to override the defaults:

```sh
PORT=3011 DB_PATH=./custom_data/shortener.db url-shortener
```

## API

Create a short link:

```sh
curl -X POST http://localhost:3011/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/about/some-path"}'
```

Create a link with a custom slug:

```sh
curl -X POST http://localhost:3011/api/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/about/some-path",
    "custom_slug": "my-link"
  }'
```
