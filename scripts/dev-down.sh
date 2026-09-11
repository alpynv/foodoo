#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
instance="${LIMA_INSTANCE:-docker}"

limactl shell "$instance" -- bash -lc '
  cd "$1"
  docker compose down
' -- "$project_root"
