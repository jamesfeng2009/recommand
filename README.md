# recommand

A Go-based backend platform for cross-source news collection, parsing, deduplication, storage, Elasticsearch indexing, and search.

- Collection/processing pipeline is decoupled by Kafka topics (`news.raw` → `news.parsed`).
- Writes go to PostgreSQL; queries are served primarily by Elasticsearch.

For a detailed module/data-flow description, see: `ARCHITECTURE.md`.

## Tech Stack

- Go 1.21
- Gin
- Kafka
- PostgreSQL
- Elasticsearch 8.x (via `go-elasticsearch/v8`)
- goquery

## Repository Layout

- `cmd/crawler-service` - Main HTTP service (source/task management + search API)
- `cmd/parsed-producer` - Consumes `news.raw`, parses HTML, produces `news.parsed`
- `cmd/news-sink` - Consumes `news.parsed`, UPSERT into PostgreSQL `news`
- `cmd/es-sync` - Incremental sync from PostgreSQL `news` to Elasticsearch (Bulk API)
- `cmd/raw-consumer` - Debug consumer for `news.raw`
- `cmd/backfill-hash` - Backfill `hash` for historical `news` rows

## Prerequisites

- Go 1.21+
- PostgreSQL
- Kafka
- Elasticsearch

This repo currently does not include a `docker-compose.yml`. You can run dependencies using your own docker/compose setup or existing infrastructure.

## Configuration (Environment Variables)

All services share the same env config (with defaults):

- `HTTP_LISTEN_ADDR` (default `:8080`)
- `DB_DSN` (default `postgres://user:password@localhost:5432/crawler?sslmode=disable`)
- `KAFKA_BROKERS` (default `localhost:9092`)
- `KAFKA_TOPIC_RAW` (default `news.raw`)
- `KAFKA_TOPIC_PARSED` (default `news.parsed`)
- `ES_ADDRESS` (default `http://localhost:9200`)
- `ES_USERNAME` (default `elastic`)
- `ES_PASSWORD` (default empty)
- `ES_INDEX` (default `news`)

## Quick Start (Local)

### 1) Start dependencies

Start PostgreSQL, Kafka, and Elasticsearch.

### 2) Create Kafka topics

Create the two topics used by the pipeline:

- `news.raw`
- `news.parsed`

(Use your Kafka admin tooling / scripts. Topic partitions can be increased for higher throughput.)

### 3) Initialize PostgreSQL schema

Create the minimal tables used by the current code:

```sql
CREATE TABLE IF NOT EXISTS news_sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  base_url TEXT NOT NULL,
  language TEXT,
  category TEXT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  crawl_interval_minutes INT NOT NULL DEFAULT 60,
  max_concurrency INT NOT NULL DEFAULT 1,
  last_crawl_at TIMESTAMPTZ,
  last_crawl_status TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS crawl_tasks (
  task_id TEXT PRIMARY KEY,
  source_id BIGINT NOT NULL,
  source_name TEXT,
  mode TEXT NOT NULL,
  since TIMESTAMPTZ,
  max_pages INT,
  status TEXT NOT NULL,
  progress DOUBLE PRECISION NOT NULL DEFAULT 0,
  pages_crawled INT NOT NULL DEFAULT 0,
  articles_found INT NOT NULL DEFAULT 0,
  articles_saved INT NOT NULL DEFAULT 0,
  duplicates_skipped INT NOT NULL DEFAULT 0,
  errors INT NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  error_message TEXT,
  created_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS news (
  id BIGSERIAL PRIMARY KEY,
  hash TEXT,
  source_code TEXT,
  url TEXT,
  title TEXT,
  content TEXT,
  publish_time TIMESTAMPTZ,
  crawl_time TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_news_hash ON news(hash);
CREATE INDEX IF NOT EXISTS idx_news_source_code_publish_time ON news(source_code, publish_time);
CREATE INDEX IF NOT EXISTS idx_news_updated_at ON news(updated_at);
```

### 4) Run services

Open multiple terminals and run:

- Main HTTP service:

```bash
go run ./cmd/crawler-service
```

- Kafka pipeline workers:

```bash
go run ./cmd/parsed-producer
```

```bash
go run ./cmd/news-sink
```

- PostgreSQL → Elasticsearch sync:

```bash
go run ./cmd/es-sync
```

(Optional) Observe raw messages:

```bash
go run ./cmd/raw-consumer
```

## HTTP APIs

Base prefix: `/api/v1`

### Health

- `GET /api/v1/crawler/health`

### News Sources

- `GET /api/v1/crawler/sources`
- `POST /api/v1/crawler/sources`
- `PUT /api/v1/crawler/sources/:id`
- `PUT /api/v1/crawler/sources/:id/status`

### Crawl Tasks

- `POST /api/v1/crawler/tasks`
- `GET /api/v1/crawler/tasks`
- `GET /api/v1/crawler/tasks/:task_id`
- `POST /api/v1/crawler/tasks/:task_id/stop`

### Search

- `GET /api/v1/search?query=xxx`

Example:

```bash
curl 'http://localhost:8080/api/v1/search?query=military'
```

## Notes

- Current `crawler-service` includes a demo/fake crawler engine and only performs a real HTTP fetch for specific configured sources (see `ARCHITECTURE.md`).
- Elasticsearch index mapping is not managed by this repo yet. Create your index/mapping based on your production needs (text fields for `title/content`, keyword fields for `source_code/url/hash`, date fields for `publish_time/crawl_time/updated_at`).

## Troubleshooting

- If Elasticsearch is HTTPS with self-signed certs, ensure `ES_ADDRESS/ES_USERNAME/ES_PASSWORD` are correct for your environment.
- If Kafka connection fails, verify `KAFKA_BROKERS` and that topics exist.
- If DB errors occur, verify `DB_DSN` and schema initialization.
