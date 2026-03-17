#!/usr/bin/env bash
# Deploy script voor de Mac.
# Draai via cron elke 10 minuten:
#   */10 * * * * /pad/naar/DartScheduler/scripts/deploy.sh >> /pad/naar/DartScheduler/scripts/deploy.log 2>&1
set -euo pipefail

# Cron runs with a minimal PATH — add the locations where Docker, git and
# curl live on macOS (both Intel /usr/local and Apple Silicon /opt/homebrew).
export PATH="/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:$PATH"

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="$REPO_DIR/backups"
COMPOSE="docker compose"
SERVICE="dartscheduler"
DB_PATH_IN_CONTAINER="/data/dartscheduler.db"

# ── 1. Zijn er nieuwe commits? ───────────────────────────────────────────────
git -C "$REPO_DIR" fetch origin master --quiet

LOCAL=$(git -C "$REPO_DIR" rev-parse HEAD)
REMOTE=$(git -C "$REPO_DIR" rev-parse origin/master)

if [ "$LOCAL" = "$REMOTE" ]; then
  exit 0  # Niets te doen
fi

echo "=== Nieuwe versie gevonden, deploy gestart: $(date) ==="
echo "    $LOCAL → $REMOTE"

# ── 2. Database backup ───────────────────────────────────────────────────────
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/dartscheduler-$(date +%Y%m%d-%H%M%S).db"

if $COMPOSE -f "$REPO_DIR/docker-compose.yml" ps --status running | grep -q "$SERVICE"; then
  echo "→ Database backup naar $BACKUP_FILE"
  $COMPOSE -f "$REPO_DIR/docker-compose.yml" cp \
    "$SERVICE:$DB_PATH_IN_CONTAINER" "$BACKUP_FILE"
  echo "  Backup klaar ($(du -sh "$BACKUP_FILE" | cut -f1))"
else
  echo "→ Container niet actief, backup overgeslagen"
fi

# Bewaar alleen de laatste 10 backups
ls -t "$BACKUP_DIR"/dartscheduler-*.db 2>/dev/null | tail -n +11 | xargs rm -f

# ── 3. Nieuwste code ophalen ─────────────────────────────────────────────────
echo "→ git pull"
git -C "$REPO_DIR" pull --ff-only origin master

# ── 4. Bouwen en herstarten ──────────────────────────────────────────────────
echo "→ docker compose build"
$COMPOSE -f "$REPO_DIR/docker-compose.yml" build --pull

echo "→ docker compose up"
$COMPOSE -f "$REPO_DIR/docker-compose.yml" up -d --remove-orphans

# ── 5. Health check ──────────────────────────────────────────────────────────
echo "→ Wachten op health check..."
for i in $(seq 1 24); do
  if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Server is up"
    break
  fi
  if [ "$i" -eq 24 ]; then
    echo "FOUT: server niet bereikbaar na 120s" >&2
    $COMPOSE -f "$REPO_DIR/docker-compose.yml" logs --tail=50 "$SERVICE" >&2
    exit 1
  fi
  sleep 5
done

echo "=== Deploy klaar: $(date) ==="
