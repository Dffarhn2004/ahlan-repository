#!/usr/bin/env bash
# Finds the address the Hermes container can really use to reach the host, instead of assuming 172.17.0.1.
# Run as an administrator who may use Docker (NOT as chatbot-pesantren). Read-only: it only runs `docker inspect`.
#   ./detect-events-addr.sh [container]        (default: hermes-ahlan)
set -euo pipefail
C="${1:-hermes-ahlan}"
PORT=8081

docker inspect "$C" >/dev/null 2>&1 || { echo "Container '$C' tidak ditemukan / tidak bisa diakses." >&2; exit 1; }
MODE="$(docker inspect -f '{{.HostConfig.NetworkMode}}' "$C")"
if [[ "$MODE" == "host" ]]; then
  cat <<TXT
Container memakai network_mode: host. Ia menjangkau API langsung di loopback host; listener tambahan TIDAK diperlukan.
Isi /etc/ahlan/hermes.env:
  HERMES_EVENTS_URL=http://127.0.0.1:8080/api/v1/internal/hermes/events
  (hapus baris HERMES_EVENTS_LISTEN)
TXT
  exit 0
fi

mapfile -t NETS < <(docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{"\n"}}{{end}}' "$C" | sed '/^$/d')
[[ ${#NETS[@]} -gt 0 ]] || { echo "Container tidak terhubung ke network mana pun." >&2; exit 1; }
[[ ${#NETS[@]} -gt 1 ]] && echo "PERINGATAN: container ada di ${#NETS[@]} network (${NETS[*]}); memakai yang pertama." >&2

NET="${NETS[0]}"
GW="$(docker inspect -f "{{(index .NetworkSettings.Networks \"$NET\").Gateway}}" "$C")"
NID="$(docker network inspect -f '{{.Id}}' "$NET")"
IFACE="$(docker network inspect -f '{{index .Options "com.docker.network.bridge.name"}}' "$NET" 2>/dev/null || true)"
[[ -n "$IFACE" && "$IFACE" != "<no value>" ]] || { [[ "$NET" == "bridge" ]] && IFACE=docker0 || IFACE="br-${NID:0:12}"; }

[[ -n "$GW" ]] || { echo "Gateway network '$NET' kosong." >&2; exit 1; }
ip -4 addr show dev "$IFACE" 2>/dev/null | grep -q "inet $GW/" || echo "PERINGATAN: $GW belum terpasang di interface $IFACE (network belum aktif?). API akan mencoba bind ulang tiap 5 detik." >&2

cat <<TXT
Network container : $NET   (interface $IFACE, gateway $GW)

1) Isi /etc/ahlan/hermes.env:
  HERMES_EVENTS_LISTEN=$GW:$PORT
  HERMES_EVENTS_URL=http://$GW:$PORT/api/v1/internal/hermes/events

2) Firewall - hanya interface itu yang boleh mencapai listener:
  sudo ufw allow in on $IFACE to $GW port $PORT proto tcp
  sudo ufw deny $PORT/tcp

3) Uji dari DALAM container setelah API jalan (harus 401, bukan timeout/refused):
  docker exec $C sh -c 'wget -qO- --post-data="{}" http://$GW:$PORT/api/v1/internal/hermes/events || true'
  docker exec $C sh -c 'python3 -c "import urllib.request as u;print(u.urlopen(u.Request(\"http://$GW:$PORT/api/v1/internal/hermes/events\",data=b\"{}\")).status)"' 2>&1 | tail -1
TXT
