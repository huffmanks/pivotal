# Pivotal

The turning point for your traffic. Adaptable routing for dynamic campaigns.

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

     Download the latest release for your platform from [GitHub Releases](https://github.com/huffmanks/pivotal/releases/latest).

   - #### Build from source

     Clone the repository:

     ```sh
     git clone https://github.com/huffmanks/pivotal.git
     ```

     Build the binary:

     ```sh
     go build -o pivotal ./cmd/server
     ```

2. Make the binary executable:

```sh
chmod +x pivotal-*
```

3. Move it somewhere in your `PATH`:

```sh
sudo mv pivotal-* /usr/local/bin/pivotal
# OR
# mv pivotal-* ~/.local/bin/pivotal
```

4. Start the server:

```sh
pivotal
```

By default, the database is stored in the system’s application data directory:

- macOS: `~/Library/Application Support/pivotal/`
- Linux: `~/.config/pivotal/`

#### Configuration

Set environment variables to override the defaults:

```sh
PORT=3011 DB_PATH=./custom_data/pivotal.db pivotal
```

## API

Create a short link:

```sh
curl -X POST http://localhost:3011/api/links \
  -H "Content-Type: application/json" \
  -d '{"destination_url":"https://example.com/about/some-path"}'
```

Create a link with a custom slug:

```sh
curl -X POST http://localhost:3011/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "destination_url": "https://example.com/about/some-path",
    "slug": "my-link"
  }'
```
