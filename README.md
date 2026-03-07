# lyryx-backend

## MinIO avatar upload flow

В проекте используется двухшаговая схема загрузки аватарок:

1. Клиент вызывает `POST /v1/user/me/avatar/upload-url` и получает:
   - `upload_url` — presigned URL для прямой загрузки файла в MinIO;
   - `avatar_url` — URL, который нужно сохранить в профиле;
   - `object_key` — ключ объекта в бакете.
2. Клиент делает `PUT` файла на `upload_url` c `Content-Type` из ответа.
3. Клиент вызывает `PATCH /v1/user/me/avatar` c `{"avatar_url": "..."}`.

## Local setup (Postgres + MinIO)

```bash
docker compose up -d
```

MinIO поднимается на:
- API: `http://localhost:9000`
- Console: `http://localhost:9001`

Дефолтные креды (из `docker-compose.yml`):
- user: `minioadmin`
- password: `minioadmin`

## Environment variables

```env
PG_DSN=postgres://user:password@localhost:5432/lyryx?sslmode=disable
JWT_SECRET=supersecret

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=avatars
MINIO_USE_SSL=false
MINIO_PUBLIC_BASE_URL=http://localhost:9000/avatars
```

`MINIO_PUBLIC_BASE_URL` должен указывать на публичный путь к объектам бакета.
