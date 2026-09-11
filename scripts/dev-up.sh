#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
instance="${LIMA_INSTANCE:-docker}"

if ! command -v limactl >/dev/null; then
  echo "limactl is required but was not found in PATH." >&2
  exit 1
fi

if ! limactl list "$instance" --format '{{.Status}}' 2>/dev/null | grep -qx 'Running'; then
  limactl start "$instance"
fi

limactl shell "$instance" -- bash -lc '
  cd "$1"
  docker compose up --build -d
  docker compose ps
' -- "$project_root"

echo "Order API: http://localhost:8080/healthz"
