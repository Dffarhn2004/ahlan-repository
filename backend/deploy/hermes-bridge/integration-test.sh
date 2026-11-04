#!/usr/bin/env bash
# Real-Docker integration test of the bridge (writes .env through stdin, knowledge, hooks, isolation, no secrets in argv).
# It creates and removes its OWN throw-away alpine container (ahlan-it-*); it never touches hermes-ahlan or any profile.
# Run on the VPS as an administrator who may use Docker, from the backend source directory:
#   cd ~/projects/backend && sudo -E env "PATH=$PATH" deploy/hermes-bridge/integration-test.sh
set -euo pipefail
cd "$(dirname "$0")/../.."
command -v go >/dev/null || { echo "Go belum terpasang"; exit 1; }
docker info >/dev/null || { echo "Docker tidak dapat diakses oleh user ini"; exit 1; }
go test -count=1 -tags integration -run Integration -v ./internal/hermes
