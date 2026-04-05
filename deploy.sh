#!/usr/bin/env bash
set -euo pipefail

SRC_VOLUME="dartscheduler_dart_data"
DST_VOLUME="dartschedulepublic_dart_data_public"
ENV_FILES=(
  "frontend/src/environments/environment.ts"
  "frontend/src/environments/environment.prod.ts"
)

echo "==> 1. git pull"
git pull

echo "==> 2. Data kopiëren: $SRC_VOLUME → $DST_VOLUME"
docker run --rm \
  -v "${SRC_VOLUME}:/source:ro" \
  -v "${DST_VOLUME}:/dest" \
  alpine sh -c "cp -a /source/. /dest/"

echo "==> 3. Versie instellen op 'dev'"
for f in "${ENV_FILES[@]}"; do
  sed -i.bak "s/version: '.*'/version: 'dev'/" "$f"
done

echo "==> 4. Docker: down → build → up"
docker compose -f docker-compose-public.yml down
docker compose -f docker-compose-public.yml build --no-cache
docker compose -f docker-compose-public.yml up -d

echo "==> Herstel originele versie-bestanden"
for f in "${ENV_FILES[@]}"; do
  mv "${f}.bak" "$f"
done

echo "==> Klaar"
