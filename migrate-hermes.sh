#!/usr/bin/env bash
# Migrasi STAGING konfigurasi Hermes (laptop Docker) -> VPS (~/hermes-migrate/).
# Jalankan dari root project (Git Bash):  ./migrate-hermes.sh [opsi]
#
# Opsi:
#   --no-upload         buat + scan arsip di laptop saja, tanpa upload
#   --approve-reviewed  lanjut walau scan menemukan pola mirip secret
#                       (HANYA setelah .hermes-staging/secret-findings.txt Anda tinjau)
#
# Yang TIDAK dilakukan script ini: ekstrak ke ~/.hermes, restart Hermes,
# ubah gateway, aktifkan cron, menyentuh data Hermes VPS.
set -euo pipefail
export MSYS_NO_PATHCONV=1   # cegah Git Bash merusak path /data, /out

SSH_HOST="ahlan-server"
VOLUME="deploy_hermes-data"
IMAGE="alpine:3.20"
REMOTE_DIR="hermes-migrate"          # relatif ke home user di VPS
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STAGE="$ROOT/.hermes-staging"
STAMP="$(date +%Y%m%d-%H%M%S)"
ARCHIVE="hermes-ahlan-$STAMP.tar.gz"

UPLOAD=1; APPROVED=0
for a in "$@"; do
  case "$a" in
    --no-upload) UPLOAD=0 ;;
    --approve-reviewed) APPROVED=1 ;;
    *) echo "Opsi tidak dikenal: $a"; exit 1 ;;
  esac
done

echo "==> [1/6] Prasyarat"
docker info >/dev/null 2>&1 || { echo "Docker tidak berjalan"; exit 1; }
docker volume inspect "$VOLUME" >/dev/null 2>&1 || { echo "Volume $VOLUME tidak ditemukan"; exit 1; }
if [[ $UPLOAD -eq 1 ]]; then
  ssh -G "$SSH_HOST" >/dev/null 2>&1 || { echo "Alias SSH $SSH_HOST tidak ada"; exit 1; }
fi

rm -rf "$STAGE"; mkdir -p "$STAGE"
STAGE_WIN="$(cygpath -w "$STAGE")"

# ---- pola mirip-secret (dipakai scan di dalam container) -------------------
cat > "$STAGE/patterns.txt" <<'PATTERNS'
sk-[A-Za-z0-9_-]{20,}
[0-9]{8,10}:[A-Za-z0-9_-]{35}
AKIA[0-9A-Z]{16}
-----BEGIN [A-Z ]*PRIVATE KEY-----
gh[pousr]_[A-Za-z0-9]{30,}
xox[baprs]-[A-Za-z0-9-]{10,}
AIza[0-9A-Za-z_-]{35}
eyJ[A-Za-z0-9_-]{15,}\.[A-Za-z0-9_-]{15,}\.
(api[_-]?key|secret|token|passw(or)?d|authorization|bearer|credential)[A-Za-z0-9_"']{0,20}[ ]*[:=][ ]*["']?[A-Za-z0-9_./+=-]{16,}
PATTERNS

# ---- script yang jalan di dalam container (volume read-only) ---------------
cat > "$STAGE/inner.sh" <<'INNER'
set -eu
cd /data
OUT=/out
LIST=$OUT/filelist.raw
: > "$LIST"

add_file() { if [ -f "$1" ]; then echo "$1" >> "$LIST"; fi; }
add_tree() { if [ -d "$1" ]; then find "$1" -type f >> "$LIST"; fi; }

# ===== ALLOWLIST =====
add_file SOUL.md
add_file config.yaml
add_tree skills

for d in profiles/*; do
  [ -d "$d" ] || continue
  for f in SOUL.md PROFILE_DESCRIPTION.md config.yaml enabled_skills.json; do add_file "$d/$f"; done
  for t in knowledge skills hooks plugins platforms locales memories; do add_tree "$d/$t"; done
  # cron: hanya definisi job (json/yaml), bukan executions.db / output / lock
  for f in "$d"/cron/*.json "$d"/cron/*.yaml "$d"/cron/*.yml; do add_file "$f"; done
  # kode bridge WhatsApp (tanpa node_modules / session, difilter di bawah)
  add_tree "$d/scripts/whatsapp-bridge"
done

# ===== DENYLIST sebagai sabuk pengaman kedua =====
NAME='(^|/)(\.env[^/]*|auth\.json[^/]*|creds\.json|id_[a-z0-9_]+|[^/]*\.(pem|key|db|db-shm|db-wal|sqlite[0-9]*|lock))$'
PATHS='(^|/)(node_modules|\.git|session|sessions|auth_info[^/]*|logs|cache|backups|\.curator_backups|whatsapp)/'
sort -u "$LIST" > "$OUT/filelist.sorted"
grep -Ev "$NAME" "$OUT/filelist.sorted" | grep -Ev "$PATHS" > "$OUT/filelist.txt" || true
rm -f "$OUT/filelist.raw" "$OUT/filelist.sorted"

# ===== SCAN pola secret (hanya file:baris, TIDAK menampilkan isi) =====
: > "$OUT/secret-findings.txt"
tr '\n' '\0' < "$OUT/filelist.txt" \
  | xargs -0 grep -nHsiE -f "$OUT/patterns.txt" 2>/dev/null \
  | cut -d: -f1,2 > "$OUT/secret-findings.txt" || true

echo "$(wc -l < "$OUT/filelist.txt") file lolos allowlist"
echo "$(wc -l < "$OUT/secret-findings.txt") baris mirip-secret"

if [ -s "$OUT/secret-findings.txt" ] && [ "${APPROVED:-0}" != "1" ]; then
  echo "SCAN: ditemukan kandidat secret, arsip TIDAK dibuat."
  exit 0
fi
tar -czf "$OUT/$ARCHIVE" -T "$OUT/filelist.txt"
INNER

echo "==> [2/6] Build allowlist + scan secret (volume read-only)"
docker run --rm \
  -e "ARCHIVE=$ARCHIVE" -e "APPROVED=$APPROVED" \
  -v "$VOLUME:/data:ro" \
  -v "$STAGE_WIN:/out" \
  "$IMAGE" sh /out/inner.sh

if [[ -s "$STAGE/secret-findings.txt" && $APPROVED -ne 1 ]]; then
  echo
  echo "DIBATALKAN: kandidat secret per file (jumlah baris cocok):"
  cut -d: -f1 "$STAGE/secret-findings.txt" | sort | uniq -c | sort -rn | head -40
  echo
  echo "Daftar lengkap file:baris: $STAGE/secret-findings.txt"
  echo "Tinjau file-file itu (banyak yang mungkin false positive, mis. contoh di dokumentasi skill)."
  echo "Jika sudah aman, jalankan ulang:  ./migrate-hermes.sh --approve-reviewed"
  exit 2
fi
[[ -f "$STAGE/$ARCHIVE" ]] || { echo "Arsip tidak terbentuk"; exit 1; }

echo "==> [3/6] Verifikasi isi arsip (laptop)"
tar -tzf "$STAGE/$ARCHIVE" > "$STAGE/archive-contents.txt"
if grep -Eai '(^|/)(\.env|auth\.json|creds\.json|id_[a-z0-9_]+)|\.(pem|key|db|sqlite)|node_modules/|whatsapp/session' "$STAGE/archive-contents.txt"; then
  echo "DIBATALKAN: arsip memuat path terlarang (daftar di atas)"; exit 1
fi
echo "    $(wc -l < "$STAGE/archive-contents.txt") entri, bersih dari path terlarang"
echo "    Isi per profile:"
grep -E '^profiles/[^/]+/' "$STAGE/archive-contents.txt" | cut -d/ -f2 | sort | uniq -c | sed 's/^/      /'

echo "==> [4/6] Checksum SHA-256"
( cd "$STAGE" && sha256sum "$ARCHIVE" > "$ARCHIVE.sha256" && cat "$ARCHIVE.sha256" && du -h "$ARCHIVE" | cut -f1 )

if [[ $UPLOAD -eq 0 ]]; then
  echo "Mode --no-upload: selesai. Arsip ada di $STAGE/$ARCHIVE"
  exit 0
fi

echo "==> [5/6] Upload ke $SSH_HOST:~/$REMOTE_DIR/"
ssh "$SSH_HOST" "mkdir -p ~/$REMOTE_DIR && chmod 700 ~/$REMOTE_DIR"
( cd "$STAGE" && scp "$ARCHIVE" "$ARCHIVE.sha256" "$SSH_HOST:$REMOTE_DIR/" )

echo "==> [6/6] Verifikasi checksum di VPS + info versi (read-only)"
ssh "$SSH_HOST" bash -s -- "$REMOTE_DIR" "$ARCHIVE" <<'REMOTE'
set -eu
cd ~/"$1"
sha256sum -c "$2.sha256"
ls -la
echo "--- container Hermes di VPS (read-only) ---"
if command -v docker >/dev/null 2>&1; then
  docker ps --format '{{.Names}}  {{.Image}}  {{.Status}}' 2>&1 | head -20 || true
  for c in $(docker ps --format '{{.Names}} {{.Image}}' 2>/dev/null | awk '/hermes/{print $1}'); do
    echo "$c: $(docker exec "$c" hermes --version 2>&1 | head -1)"
  done
else
  echo "docker tidak tersedia untuk user ini"
fi
REMOTE

cat <<EOF

Selesai. Arsip staging: ~/$REMOTE_DIR/$ARCHIVE (checksum OK bila baris di atas 'OK').
Belum ada yang diekstrak atau diaktifkan. Laptop versi Hermes: v0.22.5.
Langkah berikut (manual): backup config Hermes VPS, bandingkan versi, lalu merge SATU profile.
EOF
