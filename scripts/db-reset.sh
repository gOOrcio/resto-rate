#!/usr/bin/env bash
set -euo pipefail

echo "==> Stopping containers and removing volumes..."
docker compose down -v

echo "==> Starting fresh (waiting for healthy)..."
docker compose up -d --wait

echo "==> Done. DB is clean."
