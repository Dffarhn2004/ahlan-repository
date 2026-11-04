#!/usr/bin/env bash
# Diagnoses why the Hermes log hook cannot reach the API ("telegram log hook delivery failed: <urlopen error timed out>").
# Read-only: it only runs docker inspect/exec, ss and ufw status. Run as an administrator who may use Docker.
#   sudo ./check-events-path.sh [container] [env-file]     (defaults: hermes-ahlan, /etc/ahlan/hermes.env)
#
# A TIMEOUT (not "refused") normally means a firewall silently drops the packets; "refused" means nothing listens there.
set -uo pipefail
C="${1:-hermes-ahlan}"
ENVFILE="${2:-/etc/ahlan/hermes.env}"
ok()   { printf '[OK]    %s\n' "$*"; }
bad()  { printf '[GAGAL] %s\n' "$*"; FAILED=1; }
info() { printf '        %s\n' "$*"; }
FAILED=0

[[ -r "$ENVFILE" ]] || { echo "Tidak bisa membaca $ENVFILE" >&2; exit 2; }
URL="$(grep -E '^HERMES_EVENTS_URL=' "$ENVFILE" | tail -1 | cut -d= -f2-)"
LISTEN="$(grep -E '^HERMES_EVENTS_LISTEN=' "$ENVFILE" | tail -1 | cut -d= -f2-)"
[[ -n "$URL" ]] || { echo "HERMES_EVENTS_URL kosong di $ENVFILE (API akan memakai host.docker.internal, yang tidak ada di Linux)." >&2; exit 2; }
docker inspect "$C" >/dev/null 2>&1 || { echo "Container '$C' tidak ditemukan." >&2; exit 2; }

HOSTPORT="$(printf '%s' "$URL" | sed -E 's#^https?://([^/]+).*#\1#')"
HOST="${HOSTPORT%%:*}"
PORT="${HOSTPORT##*:}"
[[ "$PORT" == "$HOSTPORT" ]] && PORT=80
echo "URL callback : $URL"
echo "Listener API : ${LISTEN:-<tidak diatur>}"
echo

# 1. Does the URL point at the gateway the container really uses?
NET="$(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}} {{end}}' "$C" | awk '{print $1}')"
GW="$(docker inspect -f "{{(index .NetworkSettings.Networks \"$NET\").Gateway}}" "$C" 2>/dev/null)"
MODE="$(docker inspect -f '{{.HostConfig.NetworkMode}}' "$C")"
if [[ "$MODE" == "host" ]]; then
  ok "container memakai network host; callback lewat loopback"
elif [[ -n "$GW" && "$HOST" == "$GW" ]]; then
  ok "host di URL ($HOST) sama dengan gateway network '$NET'"
else
  bad "host di URL ($HOST) BUKAN gateway network '$NET' (${GW:-?}). Jalankan detect-events-addr.sh dan samakan hermes.env."
fi

# 2. Is anything listening on that address?
if ss -ltn 2>/dev/null | awk '{print $4}' | grep -qE "(^|:|\])${PORT}$"; then
  BOUND="$(ss -ltn | awk '{print $4}' | grep -E ":${PORT}$" | tr '\n' ' ')"
  ok "ada listener di port $PORT: $BOUND"
  if [[ "$HOST" != "127.0.0.1" ]] && ! printf '%s' "$BOUND" | grep -qE "(^| )(0\.0\.0\.0|\*|\[::\]|${HOST//./\\.}):${PORT}( |$)"; then
    bad "listener ada, tetapi tidak terikat ke $HOST. Cek HERMES_EVENTS_LISTEN dan restart ahlan-api."
  fi
else
  bad "tidak ada listener di port $PORT. Pastikan ahlan-api jalan dan HERMES_EVENTS_LISTEN terisi (journalctl -u ahlan-api | grep -i events)."
fi

# 3. Firewall.
if command -v ufw >/dev/null 2>&1; then
  if ufw status 2>/dev/null | grep -qi 'inactive'; then
    ok "ufw tidak aktif"
  elif ufw status 2>/dev/null | grep -qE "(^|[^0-9])${PORT}(/tcp)?[[:space:]]+(on[[:space:]]+[A-Za-z0-9._-]+[[:space:]]+)?ALLOW"; then
    ok "ufw punya aturan ALLOW untuk port $PORT"
  else
    bad "ufw aktif tetapi tidak ada aturan ALLOW untuk port $PORT. Paket dari container kemungkinan di-DROP (gejala: timeout)."
    info "Perintah dari detect-events-addr.sh:  sudo ufw allow in on <interface> to $HOST port $PORT proto tcp"
  fi
fi

# 4. The URL baked into each profile's hook must match hermes.env exactly.
mapfile -t HOOKURLS < <(docker exec "$C" sh -c 'grep -rhoE "https?://[^\"'"'"' ]+events" /opt/data/profiles/*/hooks/ponpes-telegram-log/handler.py 2>/dev/null' | sort -u)
if [[ ${#HOOKURLS[@]} -eq 0 ]]; then
  bad "tidak menemukan URL di handler.py hook log; hook belum terpasang (jalankan Sinkronkan semua)."
else
  for u in "${HOOKURLS[@]}"; do
    if [[ "$u" == "$URL" ]]; then ok "hook memakai URL yang sama: $u"; else bad "hook memakai URL LAMA: $u  (hermes.env: $URL) -> jalankan Sinkronkan semua"; fi
  done
fi

# 5. The real test: a request from inside the container. 401 = reachable (no key sent), which is what we want.
RESULT="$(docker exec -i "$C" python3 - "$URL" <<'PY' 2>&1 | tail -1
import sys, urllib.request, urllib.error, socket
url = sys.argv[1]
try:
    urllib.request.urlopen(urllib.request.Request(url, data=b"{}"), timeout=5)
    print("STATUS 200")
except urllib.error.HTTPError as e:
    print("STATUS %d" % e.code)
except (socket.timeout, TimeoutError):
    print("TIMEOUT")
except urllib.error.URLError as e:
    r = e.reason
    print("TIMEOUT" if isinstance(r, (socket.timeout, TimeoutError)) or "timed out" in str(r) else "ERROR %s" % r)
except Exception as e:
    print("ERROR %s" % e)
PY
)"
case "$RESULT" in
  "STATUS 401"|"STATUS 403") ok "dari dalam container: endpoint terjangkau (HTTP ${RESULT#STATUS }; ditolak karena tanpa kunci, itu wajar)";;
  "STATUS 200") ok "dari dalam container: endpoint terjangkau (HTTP 200)";;
  TIMEOUT) bad "dari dalam container: TIMEOUT. Paket di-DROP; periksa langkah 3 (firewall) dan langkah 2 (listener).";;
  *) bad "dari dalam container: $RESULT";;
esac

echo
if [[ $FAILED -eq 0 ]]; then echo "Semua pemeriksaan lulus."; else echo "Ada pemeriksaan yang gagal; perbaiki dari yang paling atas."; fi
exit $FAILED
