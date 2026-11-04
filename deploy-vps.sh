#!/usr/bin/env bash
# Deploy Ahlan (backend Go + frontend Next.js) ke VPS: upload -> build -> migrasi DB -> restart -> health check -> rollback otomatis.
# Jalankan dari root project (Git Bash):  ./deploy-vps.sh [backend|frontend|all]   (default: all)
# Env opsional: SKIP_TESTS=1 (lewati go test), FORCE_NPM_CI=1 (paksa npm ci)
set -euo pipefail

SSH_HOST="chatbot-ahlan-server"   # user chatbot-pesantren (upload + build)
SSH_ADMIN="ahlan-server"          # user daffa (sudo systemctl untuk backend)
REMOTE_HOME="/home/chatbot-pesantren"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${1:-all}"
SKIP_TESTS="${SKIP_TESTS:-0}"
FORCE_NPM_CI="${FORCE_NPM_CI:-0}"

# Bridge: berjalan sebagai service milik root/ahlan-bridge dari /opt/ahlan/bin, jadi dipasang lewat sudo (user daffa).
BRIDGE_BIN="/opt/ahlan/bin/hermes-bridge"
BRIDGE_STAGED="$REMOTE_HOME/projects/backend/bin/hermes-bridge"
BRIDGE_INSTALL="{ sudo cp -f $BRIDGE_BIN $BRIDGE_BIN.prev || true; } && sudo install -m 0755 $BRIDGE_STAGED $BRIDGE_BIN && sudo systemctl restart ahlan-hermes-bridge && sleep 2 && sudo systemctl is-active --quiet ahlan-hermes-bridge"
BRIDGE_ROLLBACK="if [ -f $BRIDGE_BIN.prev ]; then sudo install -m 0755 $BRIDGE_BIN.prev $BRIDGE_BIN && sudo systemctl restart ahlan-hermes-bridge; fi"

EXCLUDES=(
  --exclude=.git --exclude=node_modules --exclude=.next --exclude=bin
  --exclude=.env --exclude='.env.*' --exclude=dist --exclude=build
  --exclude=.turbo --exclude=coverage --exclude='*.exe'
  --exclude='*.tsbuildinfo' --exclude='*.pem' --exclude='*.key'
)

case "$TARGET" in
  all) PARTS=(backend frontend) ;;
  backend|frontend) PARTS=("$TARGET") ;;
  *) echo "Pakai: $0 [backend|frontend|all]"; exit 1 ;;
esac

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "==> [1/7] Verifikasi project"
[[ -f "$ROOT/backend/go.mod" ]]        || { echo "backend/go.mod tidak ada"; exit 1; }
[[ -f "$ROOT/frontend/package.json" ]] || { echo "frontend/package.json tidak ada"; exit 1; }

echo "==> [2/7] Kompres source"
for p in "${PARTS[@]}"; do
  tar -czf "$TMP/ahlan-$p.tar.gz" "${EXCLUDES[@]}" -C "$ROOT/$p" .
  echo "    ahlan-$p.tar.gz  $(du -h "$TMP/ahlan-$p.tar.gz" | cut -f1)"
done

echo "==> [3/7] Cek file sensitif di dalam arsip"
for p in "${PARTS[@]}"; do
  if tar -tzf "$TMP/ahlan-$p.tar.gz" | grep -E '(^|/)(\.env|id_[a-z0-9_]+$)|\.pem$|\.key$|node_modules/|(^|/)\.git/'; then
    echo "DIBATALKAN: file sensitif ditemukan di arsip $p"; exit 1
  fi
done
echo "    bersih"

echo "==> [4/7] Upload ke $SSH_HOST"
(cd "$TMP" && for p in "${PARTS[@]}"; do scp -q "ahlan-$p.tar.gz" "$SSH_HOST:$REMOTE_HOME/"; done)

# ---------------------------------------------------------------- BACKEND
deploy_backend() {
  echo "==> [5/7] Backend: ekstrak + test + build + migrasi database (binary lama tetap jalan)"
  ssh "$SSH_HOST" bash -s -- "$SKIP_TESTS" <<'REMOTE'
set -euo pipefail
export PATH="$PATH:/usr/local/go/bin:$HOME/.local/go/bin:$HOME/go/bin"
mkdir -p ~/projects/backend/bin
tar -xzf ~/ahlan-backend.tar.gz -C ~/projects/backend
rm -f ~/ahlan-backend.tar.gz
cd ~/projects/backend
[[ "$1" == "1" ]] || go test ./...
go build -o bin/ahlan-api.new ./cmd/api
go build -o bin/ahlan-worker.new ./cmd/worker
go build -o bin/hermes-bridge.new ./cmd/hermes-bridge
# Migrasi database (idempoten: yang sudah diterapkan dilewati). Berjalan SEBELUM binary diganti: bila gagal, set -e
# menghentikan deploy dan binary lama tetap melayani. Membaca DATABASE_URL dari .env di folder ini.
echo "    migrasi database"
go run ./cmd/migrate
# simpan versi sebelumnya untuk rollback, lalu ganti binary
for b in ahlan-api ahlan-worker hermes-bridge; do
  [[ -f bin/$b ]] && cp -f bin/$b bin/$b.prev
  mv -f bin/$b.new bin/$b
done
echo "    binary baru terpasang"
REMOTE

  echo "==> [6/7] Backend: pasang bridge, lalu restart service (mungkin diminta password sudo)"
  # Bridge lebih dulu: API/Worker baru memanggil rute bridge yang baru. Biner lama disimpan sebagai .prev untuk rollback.
  # Restart bridge tidak memutus bot (ia hanya perantara).
  if ! ssh -t "$SSH_ADMIN" "$BRIDGE_INSTALL && sudo systemctl restart ahlan-worker ahlan-api"; then
    echo "!! restart gagal, rollback"; rollback_backend; return 1
  fi

  echo "==> [7/7] Backend: health check"
  if ! ssh "$SSH_HOST" 'for i in $(seq 1 15); do curl -fsS http://127.0.0.1:8080/healthz >/dev/null && exit 0; sleep 2; done; exit 1'; then
    echo "!! health check gagal, rollback ke binary sebelumnya"; rollback_backend; return 1
  fi
  echo "    backend OK"
}

rollback_backend() {
  ssh "$SSH_HOST" 'cd ~/projects/backend/bin && for b in ahlan-api ahlan-worker hermes-bridge; do [[ -f $b.prev ]] && cp -f $b.prev $b; done' || true
  ssh -t "$SSH_ADMIN" "$BRIDGE_ROLLBACK; sudo systemctl restart ahlan-worker ahlan-api" || true
  echo "    rollback selesai (cek: sudo journalctl -u ahlan-api -n 50)"
}

# --------------------------------------------------------------- FRONTEND
deploy_frontend() {
  echo "==> [5/7] Frontend: ekstrak + install (jika perlu) + build"
  ssh "$SSH_HOST" bash -s -- "$FORCE_NPM_CI" <<'REMOTE'
set -euo pipefail
export PATH="$PATH:/usr/local/bin:$HOME/.local/bin"
[[ -s "$HOME/.nvm/nvm.sh" ]] && source "$HOME/.nvm/nvm.sh"
mkdir -p ~/projects/frontend
cd ~/projects/frontend
OLD_HASH="$(cat .deps-hash 2>/dev/null || true)"
tar -xzf ~/ahlan-frontend.tar.gz -C ~/projects/frontend
rm -f ~/ahlan-frontend.tar.gz
NEW_HASH="$(sha256sum package.json package-lock.json 2>/dev/null | sha256sum | cut -d' ' -f1)"
if [[ "$1" == "1" || "$OLD_HASH" != "$NEW_HASH" || ! -d node_modules ]]; then
  npm ci --maxsockets=2 --no-audit --no-fund
  echo "$NEW_HASH" > .deps-hash
else
  echo "    dependency tidak berubah, lewati npm ci"
fi
# backup build lama supaya bisa rollback
rm -rf .next.prev
[[ -d .next ]] && cp -a .next .next.prev
if ! npm run build; then
  echo "!! build gagal, memulihkan .next lama"
  rm -rf .next; [[ -d .next.prev ]] && mv .next.prev .next
  exit 1
fi
REMOTE

  echo "==> [6/7] Frontend: restart service"
  ssh "$SSH_HOST" 'systemctl --user restart ahlan-frontend'

  echo "==> [7/7] Frontend: health check"
  if ! ssh "$SSH_HOST" 'for i in $(seq 1 15); do curl -fsSI http://127.0.0.1:3000 >/dev/null && exit 0; sleep 2; done; exit 1'; then
    echo "!! health check gagal, rollback .next"
    ssh "$SSH_HOST" 'cd ~/projects/frontend && [[ -d .next.prev ]] && rm -rf .next && mv .next.prev .next; systemctl --user restart ahlan-frontend' || true
    return 1
  fi
  echo "    frontend OK"
}

for p in "${PARTS[@]}"; do
  echo; echo "########## $p ##########"
  "deploy_$p"
done

echo; echo "Selesai. Deploy ${PARTS[*]} berhasil."
