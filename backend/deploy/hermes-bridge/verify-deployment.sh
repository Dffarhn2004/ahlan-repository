#!/usr/bin/env bash
# Post-deploy acceptance checks. Run as an administrator (sudo). Changes nothing.
set -uo pipefail
APP=chatbot-pesantren; FRONTEND_UNIT="${FRONTEND_UNIT:-ahlan-frontend.service}"; SOCK=/run/ahlan-hermes/bridge.sock
ok=0; bad=0
pass(){ echo "  [OK]   $*"; ok=$((ok+1)); }
fail(){ echo "  [GAGAL] $*"; bad=$((bad+1)); }
chk(){ local label="$1"; shift; if "$@" >/dev/null 2>&1; then pass "$label"; else fail "$label"; fi; }
not(){ local label="$1"; shift; if "$@" >/dev/null 2>&1; then fail "$label"; else pass "$label"; fi; }

echo "== service aktif bersamaan"
for u in ahlan-hermes-bridge ahlan-api ahlan-worker; do chk "$u active" systemctl is-active --quiet "$u"; done
chk "frontend user-service ($FRONTEND_UNIT) masih aktif" systemctl --user -M "$APP"@ is-active --quiet "$FRONTEND_UNIT"
echo "== tepat satu worker & satu API; unit user lama berhenti"
WBIN="$(systemctl show ahlan-worker -p ExecStart --value | sed 's/.*path=\([^ ;]*\).*/\1/')"
n="$(pgrep -u "$APP" -fc "^$WBIN( |\$)")"; [[ "$n" == "1" ]] && pass "tepat satu proses worker ($WBIN)" || fail "proses worker = $n (harus 1)"
not "user-unit ahlan-worker tidak aktif" systemctl --user -M "$APP"@ is-active --quiet ahlan-worker.service
not "user-unit ahlan-api tidak aktif"    systemctl --user -M "$APP"@ is-active --quiet ahlan-api.service
echo "== direktori & socket (dicek setelah reboot juga)"
[[ "$(stat -c '%U:%G %a' /run/ahlan-hermes)" == "ahlan-bridge:ahlan-hermes 750" ]] && pass "/run/ahlan-hermes ahlan-bridge:ahlan-hermes 0750" || fail "/run/ahlan-hermes: $(stat -c '%U:%G %a' /run/ahlan-hermes 2>&1)"
[[ "$(stat -c '%U:%G %a' $SOCK)" == "ahlan-bridge:ahlan-hermes 660" ]] && pass "bridge.sock ahlan-bridge:ahlan-hermes 0660" || fail "bridge.sock: $(stat -c '%U:%G %a' $SOCK 2>&1)"
echo "== hanya yang berwenang dapat memakai bridge"
not "chatbot-pesantren (login, tanpa grup) ditolak" sudo -u "$APP" curl -sf --unix-socket $SOCK http://x/v1/health
not "user lain (nobody) ditolak" sudo -u nobody curl -sf --unix-socket $SOCK http://x/v1/health
chk "proses API (grup ahlan-hermes) dapat memakai bridge" sudo systemd-run --quiet --wait --pipe --uid="$APP" -p SupplementaryGroups=ahlan-hermes curl -sf --unix-socket $SOCK http://x/v1/health
echo "== API & Worker tidak punya jalur ke Docker"
id -nG "$APP" | tr ' ' '\n' | grep -qx docker && fail "$APP ada di grup docker" || pass "$APP tidak di grup docker"
sudo -l -U "$APP" 2>/dev/null | grep -q 'NOPASSWD\|(ALL' && fail "$APP punya sudo" || pass "$APP tanpa sudo"
for u in ahlan-api ahlan-worker; do
  pid="$(systemctl show -p MainPID --value $u)"
  [[ "$pid" != 0 ]] || { fail "$u tidak punya MainPID"; continue; }
  grps="$(ps -o supgrp= -p "$pid")"
  [[ ",$grps," == *",docker,"* ]] && fail "$u memiliki grup docker" || pass "$u tanpa grup docker (grup: $grps)"
  not "$u tidak dapat melihat docker.sock" nsenter -t "$pid" -m test -S /var/run/docker.sock
done
not "chatbot-pesantren: docker ps ditolak" sudo -u "$APP" docker ps
echo "== bridge menolak input berbahaya"
call(){ sudo systemd-run --quiet --wait --pipe --uid="$APP" -p SupplementaryGroups=ahlan-hermes curl -s -o /dev/null -w '%{http_code}' --unix-socket $SOCK "$@"; }
[[ "$(call -d '{"profile_key":"default","action":"start"}' http://x/v1/profile/action)" == 400 ]] && pass "profil 'default' ditolak" || fail "profil default tidak ditolak"
[[ "$(call -d '{"profile_key":"ponpes_a","action":"start","container":"x"}' http://x/v1/profile/action)" == 400 ]] && pass "override container ditolak" || fail "override container tidak ditolak"
[[ "$(call -d '{}' http://x/v1/exec)" == 404 ]] && pass "rute /v1/exec tidak ada" || fail "rute tak dikenal tidak 404"
echo; echo "Ringkasan: $ok lulus, $bad gagal"; exit $((bad>0))
