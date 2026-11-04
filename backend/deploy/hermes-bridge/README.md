# Hermes bridge — deployment untuk VPS Ahlan

Status: **belum dideploy, belum di-commit.** Semua perintah di bawah dijalankan manual oleh administrator server.

## 1. Gambaran

```
 Nginx ─► ahlan-api.service      (SYSTEM unit, User=chatbot-pesantren, SupplementaryGroups=ahlan-hermes, 127.0.0.1:8080)
          ahlan-worker.service   (SYSTEM unit, idem)
          ahlan-frontend         (tetap USER unit, tidak diubah)
             │ Unix socket /run/ahlan-hermes/bridge.sock  (ahlan-bridge:ahlan-hermes 0660, dir 0750)
             ▼
          ahlan-hermes-bridge.service  (User=ahlan-bridge + grup docker, tanpa jaringan, allow-list operasi)
             │ docker exec (hanya container hermes-ahlan)
             ▼
          hermes-ahlan  ──callback──►  <gateway-network-container>:8081  (listener sempit milik ahlan-api, hanya rute events)
```

Alasan API/Worker pindah ke system unit: `SupplementaryGroups=` tidak berlaku untuk `systemctl --user`, sedangkan akses
socket bridge dibatasi pada grup `ahlan-hermes`. `chatbot-pesantren` sendiri (user login) **tidak** masuk grup `docker`
maupun `ahlan-hermes`, dan tidak mendapat sudo. Hanya *proses* API dan Worker yang membawa grup `ahlan-hermes`.

Lapisan akses bridge: izin socket (grup) → UID peer `SO_PEERCRED` (`HERMES_BRIDGE_ALLOWED_UIDS`) → rute tetap →
JSON ketat → validasi per-field → validasi ulang di `DockerClient`. Tambahan di unit API/Worker:
`InaccessiblePaths=-/var/run/docker.sock -/run/docker.sock` — kedua proses *tidak dapat melihat* socket Docker
sekalipun ada bug.

## 2. Yang berubah pada revisi ini

| Item | Perubahan |
|---|---|
| `ahlan-api.service`, `ahlan-worker.service` (baru) | system unit pengganti user unit, lihat §5 |
| `ahlan-hermes-bridge.service` | `Type=notify` (siap hanya setelah socket listen), env dari `/etc/ahlan/hermes.env` |
| `cmd/hermes-bridge` | mengirim `READY=1` ke systemd |
| `cmd/api` | listener events mencoba bind ulang tiap 5 detik (gateway Docker bisa belum ada saat boot) |
| `internal/httpapi/telegram.go` | perbandingan `X-Hermes-Event-Key` constant-time |
| `hermes.env.example` | satu file env non-rahasia dipakai bridge + API + Worker (menggantikan dua file lama) |
| `detect-events-addr.sh` | membaca gateway **sebenarnya** dari Docker; tidak mengasumsikan `172.17.0.1` |
| `integration-test.sh` + `internal/hermes/integration_test.go` | uji Docker sungguhan (§8) |
| `verify-deployment.sh` | pemeriksaan acceptance setelah deploy / reboot |

## 3. Sandbox bridge vs Docker CLI (analisis)

Belum dijalankan di bawah systemd (mesin pengembangan ini tidak memakai systemd); cek akhir dilakukan `GET /v1/health`
bridge yang menjalankan `docker inspect` **di dalam sandbox sebenarnya**.

| Pengaturan | Efek terhadap `docker` CLI | Hasil analisis |
|---|---|---|
| `SupplementaryGroups=docker` | dibutuhkan agar boleh connect ke `/var/run/docker.sock` (root:docker 0660) | OK; izin file, tidak butuh capability |
| `PrivateNetwork=yes`, `RestrictAddressFamilies=AF_UNIX` | CLI hanya bicara ke socket **file**; namespace jaringan tidak memengaruhi | OK (akan rusak bila `DOCKER_HOST=tcp://…`) |
| `ProtectSystem=strict`, `ProtectHome=yes` | `/run`, `/usr` tetap terbaca; `connect()` ke socket di FS read-only tetap boleh | OK; `HOME`/`DOCKER_CONFIG` diarahkan ke `StateDirectory=ahlan-bridge` yang writable |
| `PrivateDevices`, `PrivateTmp` | `docker exec -i` (tanpa TTY) tidak butuh `/dev/pts` | OK |
| `CapabilityBoundingSet=` kosong, `NoNewPrivileges` | CLI tidak memakai capability | OK |
| `MemoryDenyWriteExecute`, `SystemCallFilter=@system-service` | CLI Go standar | OK (diharapkan) |
| **Docker dari snap** | confinement snap bentrok dengan sandbox | **Periksa** `command -v docker` (harus `/usr/bin/docker`) |

Bila `health` gagal: `journalctl -u ahlan-hermes-bridge -n 50`, lalu longgarkan berurutan —
`SystemCallFilter` → `MemoryDenyWriteExecute` → `ProtectSystem=full` → `PrivateDevices` — satu per satu, jangan
sekaligus. `systemd-analyze security ahlan-hermes-bridge` menampilkan skor setelah dilonggarkan.

## 4. Callback Hermes → API (tanpa asumsi 172.17.0.1)

Container berada di network yang dikelola administrator; gateway-nya bisa `172.17.0.1` (bridge bawaan) atau
`172.18.0.1`, dst. (compose). Jalankan:

```bash
./deploy/hermes-bridge/detect-events-addr.sh hermes-ahlan
```

Skrip membaca network container, gateway, nama interface (`docker0`/`br-xxxx`), dan mencetak: nilai
`HERMES_EVENTS_LISTEN`/`HERMES_EVENTS_URL`, perintah `ufw` untuk interface itu saja, dan perintah uji dari dalam
container. Jika container memakai `network_mode: host`, skrip menyarankan `http://127.0.0.1:8080/...` tanpa listener.

Keamanan endpoint: (1) listener hanya di alamat gateway, bukan `0.0.0.0`; (2) firewall hanya mengizinkan interface itu;
(3) hanya satu rute (`POST /api/v1/internal/hermes/events`; ada tes yang menjamin rute lain 404); (4) wajib
`X-Hermes-Event-Key`, dibandingkan constant-time; (5) bridge hanya mau memasang hook dengan URL yang sama persis dengan
`HERMES_EVENTS_URL`. Tidak perlu `extra_hosts` karena URL memakai IP, jadi **container Hermes tidak perlu diubah**.

Catatan: jika network container nanti dibuat ulang dengan subnet lain, jalankan ulang skrip, perbarui
`hermes.env`, restart `ahlan-hermes-bridge ahlan-api ahlan-worker`, lalu `POST /api/v1/hermes/sync-all`.

## 5. Runbook deployment (urut; ganti `<…>` sesuai server)

### Fase 0 — pemeriksaan awal (hanya baca)
```bash
id chatbot-pesantren; getent group docker
sudo systemctl --user -M chatbot-pesantren@ cat ahlan-api.service ahlan-worker.service   # catat ExecStart, WorkingDirectory, Environment*
sudo systemctl --user -M chatbot-pesantren@ list-units 'ahlan-*'
sudo loginctl show-user chatbot-pesantren -p Linger        # harus Linger=yes (frontend tetap butuh)
command -v docker; ls -l /var/run/docker.sock; docker inspect -f '{{.State.Running}}' hermes-ahlan
```

### Fase 1 — backup (wajib sebelum apa pun)
```bash
BK=/root/ahlan-backup-$(date +%F-%H%M); sudo mkdir -p $BK
sudo cp -a /home/chatbot-pesantren/.config/systemd/user $BK/user-units
sudo cp -a /home/chatbot-pesantren/projects/backend/bin $BK/bin            # sesuaikan dengan lokasi binary lama
docker exec hermes-ahlan tar czf - /opt/data/profiles | sudo tee $BK/hermes-profiles.tgz >/dev/null
sudo ls -lh $BK; echo "BK=$BK"
```
(Tambahkan `pg_dump` database jika belum ada backup rutin. Rilis ini tidak mengubah skema.)

### Fase 2 — build (tidak mengganggu layanan yang berjalan)
```bash
cd /home/chatbot-pesantren/projects/backend            # source dari ./deploy-vps.sh backend
go test ./... && go vet ./...
sudo -u chatbot-pesantren go build -o bin/api.new    ./cmd/api
sudo -u chatbot-pesantren go build -o bin/worker.new ./cmd/worker
go build -o /tmp/hermes-bridge ./cmd/hermes-bridge
```
(Alternatif: cross-compile di PC dev `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build …` lalu `scp`.)

### Fase 3 — pengguna, grup, berkas konfigurasi
```bash
getent group ahlan-hermes >/dev/null || sudo groupadd --system ahlan-hermes
id ahlan-bridge >/dev/null 2>&1 || sudo useradd --system --no-create-home --shell /usr/sbin/nologin --gid ahlan-hermes ahlan-bridge
sudo usermod -aG docker ahlan-bridge                 # SATU-SATUNYA akun non-admin dengan akses Docker
sudo install -d -m 0755 /opt/ahlan/bin /etc/ahlan
sudo install -m 0755 /tmp/hermes-bridge /opt/ahlan/bin/hermes-bridge
sudo install -m 0640 -g ahlan-hermes deploy/hermes-bridge/hermes.env.example /etc/ahlan/hermes.env
sudo sed -i "s/__UID_OF_CHATBOT_PESANTREN__/$(id -u chatbot-pesantren)/" /etc/ahlan/hermes.env
```

### Fase 4 — alamat callback & firewall (§4)
```bash
./deploy/hermes-bridge/detect-events-addr.sh hermes-ahlan      # salin nilai ke /etc/ahlan/hermes.env, jalankan perintah ufw yang dicetak
sudoedit /etc/ahlan/hermes.env                                 # ganti __GATEWAY_IP__ ; pastikan tak ada tanda __ tersisa
grep -n '__' /etc/ahlan/hermes.env && echo "MASIH ADA PLACEHOLDER" || echo ok
```

### Fase 5 — nyalakan bridge saja, uji (API/Worker lama masih berjalan)
```bash
sudo install -m 0644 deploy/hermes-bridge/ahlan-hermes-bridge.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now ahlan-hermes-bridge
systemctl status ahlan-hermes-bridge --no-pager | head -15
ls -ld /run/ahlan-hermes; ls -l /run/ahlan-hermes/bridge.sock    # drwxr-x--- ahlan-bridge ahlan-hermes ; srw-rw---- ahlan-bridge ahlan-hermes
sudo systemd-run --quiet --wait --pipe --uid=chatbot-pesantren -p SupplementaryGroups=ahlan-hermes \
  curl -s --unix-socket /run/ahlan-hermes/bridge.sock http://x/v1/health        # {"status":"ok"} = Docker CLI jalan di dalam sandbox
```
Bila gagal → §3. Jangan lanjut sebelum `health` berhasil.

### Fase 6 — uji integrasi Docker nyata (container uji sendiri, tidak menyentuh hermes-ahlan)
```bash
sudo -E env "PATH=$PATH" deploy/hermes-bridge/integration-test.sh      # harus: --- PASS: TestIntegrationBridgeAgainstRealDocker
```

### Fase 7 — pasang system unit (belum dijalankan)
```bash
# Sesuaikan WorkingDirectory/ExecStart/Environment dengan keluaran Fase 0 SEBELUM menyalin:
sudo install -m 0644 deploy/hermes-bridge/ahlan-api.service deploy/hermes-bridge/ahlan-worker.service /etc/systemd/system/
sudoedit /etc/systemd/system/ahlan-api.service /etc/systemd/system/ahlan-worker.service
sudo systemd-analyze verify /etc/systemd/system/ahlan-api.service /etc/systemd/system/ahlan-worker.service
sudo systemctl daemon-reload            # JANGAN enable/start dulu
```

### Fase 8 — cutover (tidak boleh ada dua Worker)
Urutan wajib: **hentikan lama → pastikan mati → baru nyalakan sistem**. Job yang sedang berjalan akan diambil ulang
Worker baru setelah lease (±1 menit) habis. Jendela henti API: beberapa detik.
```bash
U='sudo systemctl --user -M chatbot-pesantren@'
# --- Worker
$U disable --now ahlan-worker.service
pgrep -u chatbot-pesantren -af 'worker' && echo "MASIH ADA WORKER LAMA - STOP, JANGAN LANJUT" || echo "worker lama mati"
mv /home/chatbot-pesantren/projects/backend/bin/worker /home/chatbot-pesantren/projects/backend/bin/worker.prev
mv /home/chatbot-pesantren/projects/backend/bin/worker.new /home/chatbot-pesantren/projects/backend/bin/worker
sudo systemctl enable --now ahlan-worker
journalctl -u ahlan-worker -n 20 --no-pager        # "Hermes control ready mode=bridge", "provisioning worker ready"
# --- API (port 8080 dipakai bersama, jadi lama harus berhenti dulu)
$U disable --now ahlan-api.service
pgrep -u chatbot-pesantren -af '/bin/api' && echo "API LAMA MASIH HIDUP - STOP" || echo "api lama mati"
mv /home/chatbot-pesantren/projects/backend/bin/api /home/chatbot-pesantren/projects/backend/bin/api.prev
mv /home/chatbot-pesantren/projects/backend/bin/api.new /home/chatbot-pesantren/projects/backend/bin/api
sudo systemctl enable --now ahlan-api
curl -fsS http://127.0.0.1:8080/healthz
# --- frontend: tidak disentuh
$U is-active ahlan-frontend.service                # nama unit frontend sesuai server
```
Jangan `mask` unit user lama: berkasnya di `~/.config/systemd/user/` (mask akan gagal/menimpa). `disable --now` cukup;
berkasnya dibiarkan untuk rollback.

### Fase 9 — verifikasi & uji E2E
```bash
sudo FRONTEND_UNIT=<nama-unit-frontend>.service deploy/hermes-bridge/verify-deployment.sh   # semua [OK]
```
E2E dari dashboard: buat tenant baru → "Buat asisten":
```bash
journalctl -u ahlan-worker -f              # "provisioning job completed", tanpa "permission denied"
journalctl -u ahlan-hermes-bridge -f       # satu baris "bridge op" per tahap, tanpa token/isi
docker exec hermes-ahlan ls /opt/data/profiles           # ponpes_<slug> ada
docker exec hermes-ahlan ls /opt/data/profiles/ponpes_<slug>/hooks    # ponpes-telegram-log, ponpes-welcome
```
Lalu di dashboard tenant itu: simpan token Telegram, start/stop, WhatsApp (QR muncul), unggah dokumen + sinkronisasi,
kirim pesan ke bot dan pastikan percakapan tercatat (membuktikan callback sampai). Terakhir, `POST /api/v1/hermes/sync-all`
sebagai superadmin memasang ulang hook semua profil lama dengan URL callback baru.

### Fase 10 — uji reboot (poin direktori/socket)
`/run` ada di tmpfs, jadi direktori dibuat ulang tiap boot oleh `RuntimeDirectory=` milik bridge (owner `ahlan-bridge`,
group `ahlan-hermes`, 0750), dan socket dibuat 0660 saat bridge listen. `ahlan-api`/`ahlan-worker` `After=`+`Wants=`
bridge bertipe `notify`, sehingga mereka baru start setelah socket benar-benar ada.
```bash
sudo reboot
# setelah naik:
sudo FRONTEND_UNIT=<nama-unit-frontend>.service deploy/hermes-bridge/verify-deployment.sh
```
Perilaku bila bridge mati: API/Worker tetap hidup, operasi Hermes membalas 502 / job gagal `PROFILE_CREATE_FAILED`
sampai bridge pulih (sengaja memakai `Wants`, bukan `Requires`, supaya restart bridge tidak ikut me-restart API).

## 5b. Update rutin biner (tanpa mengulang Fase 0-10)

Dari root proyek di PC (Git Bash), jalankan `./deploy-vps.sh backend` (atau `all` bersama frontend). Skrip ini
mengunggah source ke `chatbot-ahlan-server`, menjalankan `go test`, membangun **`ahlan-api`, `ahlan-worker` dan
`hermes-bridge`** (Go hanya terpasang untuk user `chatbot-pesantren`, tidak untuk `daffa`), lalu lewat `ahlan-server`
(user `daffa`, sudo) memasang bridge ke `/opt/ahlan/bin/hermes-bridge` (biner lama disimpan sebagai `.prev`), me-restart
`ahlan-hermes-bridge` lebih dulu, baru `ahlan-worker` dan `ahlan-api`. Gagal health check = rollback otomatis ke `.prev`.

Urutan bridge-dulu itu penting: rilis yang menambah operasi bridge baru (mis. `POST /v1/gateway/restart`) membuat API
memanggil rute yang belum ada (404) bila bridge lama masih berjalan. Cek setelah deploy: PID `ahlan-hermes-bridge`
berubah (`systemctl show ahlan-hermes-bridge -p MainPID -p ActiveEnterTimestamp`) dan `ls -l /opt/ahlan/bin/hermes-bridge`
bertanggal baru. Restart bridge tidak memutus bot. Bila source sudah terunggah, skrip diagnosis ada di
`/home/chatbot-pesantren/projects/backend/deploy/hermes-bridge/` (jalankan sebagai `daffa` dengan `sudo bash`).

Lalu, sebagai superadmin, buka **Admin > Hermes > Sinkronkan semua** (centang "Mulai ulang asisten yang sedang aktif").

### Restart gateway bersama (`POST /v1/gateway/restart`)
Hermes menyimpan overlay locale yang sudah dibaca di memori selama proses gateway hidup. Restart *per profile* hanya
memutus-sambung adapter, jadi `id.yaml` yang baru ditulis tidak terbaca sampai proses gateway di-restart. Bridge kini
punya satu operasi tanpa argumen untuk itu (`hermes gateway restart` di container tetap). Pengamanan: rute tetap, tanpa
input, dibatasi satu kali per 2 menit (`429 BUSY` bila diulang; di sisi API otomatis jatuh ke restart per profile), dan
dicatat di audit log sebagai `HERMES_GATEWAY_RESTART`. Efek samping yang disengaja: **semua** bot tenant offline kira-kira
satu menit. Karena itu hanya dipanggil dari "Sinkronkan semua" dengan opsi restart, tidak dari tindakan per-tenant.

### "Sinkronkan semua" berjalan sebagai job latar belakang
`POST /api/v1/hermes/sync-all` langsung membalas `202 {"job_id": ...}`; halaman memantau
`GET /api/v1/hermes/sync-all/:job_id` (status, fase, `done/total`, hasil per asisten). Hanya satu job berjalan; permintaan
kedua mengembalikan job yang sedang berjalan (`already_running: true`). Job disimpan di memori API (10 terakhir): bila API
di-restart di tengah jalan, job hilang dan cukup dijalankan ulang (aman diulang). Dengan ini request HTTP tidak lagi
menunggu menit-an, jadi **tidak perlu** menaikkan `proxy_read_timeout` Nginx untuk endpoint ini. Endpoint lain yang masih
lama (mis. sinkronisasi pengetahuan per tenant) tetap perlu diperhatikan bila timeout Nginx rendah.

### Diagnosis callback log percakapan
```bash
sudo bash /home/chatbot-pesantren/projects/backend/deploy/hermes-bridge/check-events-path.sh hermes-ahlan /etc/ahlan/hermes.env
```
Memeriksa berurutan: host di `HERMES_EVENTS_URL` = gateway network container; ada listener di port itu; aturan ufw;
URL di `handler.py` tiap profil sama dengan `hermes.env`; dan uji nyata dari dalam container (401 = terjangkau).
**TIMEOUT** (bukan "refused") umumnya berarti firewall men-*drop* paket. Handler hook kini mengirim di thread latar
belakang dengan antrean terbatas, sehingga API yang tak terjangkau tidak lagi menunda balasan pengunjung (sebelumnya
menambah ±3 detik per balasan karena `timeout=3` yang sinkron). Hook lama baru diganti setelah "Sinkronkan semua".

## 6. Rollback (tanpa kehilangan data/profil)

Data profil tenant ada di volume Hermes dan database; rilis ini **tidak mengubah skema** maupun format profil, jadi
rollback hanya soal service dan binary. Backup Fase 1 tersedia bila diperlukan.
```bash
U='sudo systemctl --user -M chatbot-pesantren@'
sudo systemctl disable --now ahlan-worker ahlan-api
pgrep -u chatbot-pesantren -af 'worker|/bin/api' && echo "MASIH ADA PROSES - STOP"
B=/home/chatbot-pesantren/projects/backend/bin
mv $B/worker $B/worker.new-rollback; mv $B/worker.prev $B/worker
mv $B/api    $B/api.new-rollback;    mv $B/api.prev    $B/api
$U enable --now ahlan-worker.service ahlan-api.service          # urutan: worker boleh bersamaan, API setelah system API mati
curl -fsS http://127.0.0.1:8080/healthz
# opsional: matikan bridge, tetapi biarkan berkas dan user-nya
sudo systemctl disable --now ahlan-hermes-bridge
```
Konsekuensi jujur: binary lama memanggil Docker langsung, jadi setelah rollback keadaan kembali seperti sebelumnya —
dashboard hidup, tetapi provisioning/Telegram/WhatsApp/knowledge **kembali `permission denied`**. Hook log percakapan
di profil sudah menunjuk URL baru; untuk memulihkan hook lama ekstrak dari `$BK/hermes-profiles.tgz`
(`hooks/ponpes-telegram-log/`) atau jalankan sync-all lagi setelah roll-forward. Profil, `.env`, dan knowledge tidak berubah.
Jika hanya ingin mundur dari system unit tetapi tetap memakai binary baru: tidak didukung (binary baru di produksi
butuh bridge, dan user unit tidak bisa membawa grup `ahlan-hermes`).

## 7. Konfigurasi (`/etc/ahlan/hermes.env`)

| Variabel | Dipakai | Nilai |
|---|---|---|
| `HERMES_CONTAINER` | bridge | `hermes-ahlan` (API mengabaikannya di mode bridge) |
| `HERMES_BRIDGE_SOCKET_MODE` | bridge | `0660` (bit "other" ditolak oleh bridge) |
| `HERMES_BRIDGE_ALLOWED_UIDS` | bridge | `id -u chatbot-pesantren` |
| `HERMES_EXEC_MODE` | API, Worker | `bridge` (default otomatis bila `APP_ENV=production`; mode tak dikenal = gagal) |
| `HERMES_BRIDGE_SOCKET` | API, Worker | `/run/ahlan-hermes/bridge.sock` (dipakai bridge juga) |
| `HERMES_EVENTS_URL` | API, Worker, bridge | dari `detect-events-addr.sh` |
| `HERMES_EVENTS_LISTEN` | API | dari `detect-events-addr.sh` |

Rahasia (`DATABASE_URL`, `HERMES_API_KEY`, …) tetap di env aplikasi yang sekarang; bridge tidak menerimanya.

## 8. Uji integrasi Linux nyata

`internal/hermes/integration_test.go` (tag `integration`, hanya Linux) memutar rantai penuh
`BridgeClient → Unix socket → BridgeServer → DockerClient → docker exec` pada container Alpine sekali pakai
(`ahlan-it-*`) dengan stub CLI `hermes`. Yang dibuktikan: penulisan `.env` via stdin (baris lain tetap, token lama
terganti, tanpa `.env.tmp`); persona/hook/lockdown; sync → extract → ready → prune; **isolasi** (prune satu tenant
tidak menyentuh tenant lain, traversal ke profil lain ditolak, persona hanya memuat dokumen tenant itu); URL callback di
luar allow-list ditolak; dan log argumen `docker` menunjukkan **tidak ada token/kunci di command line** serta hanya
subcommand `exec`/`inspect`.

Hasil pada sesi ini: **LULUS** terhadap Docker Engine Linux 29.8.2 (biner uji dikompilasi untuk Linux dan dijalankan
dalam container dengan socket Docker). Belum dijalankan di VPS Ahlan dan belum terhadap image Hermes asli — CLI
`hermes` memang di-stub; jalur `hermes profile create` dsb. hanya dibuktikan argumennya. Jalankan Fase 6 di VPS, lalu
E2E Fase 9 untuk Hermes asli.

## 9. Risiko tersisa
- `ahlan-bridge` ada di grup `docker` (setara root atas host). Dibatasi: bukan akun login, tanpa jaringan/capability,
  FS read-only, allow-list operasi, satu container tetap. Mitigasi lanjutan: Docker rootless, atau socket-proxy.
- Isolasi tenant tetap di level aplikasi (API yang memetakan tenant → `profile_key`); bridge hanya menjamin path di
  bawah profile key yang valid. Semua tenant berbagi satu container & gateway Hermes. `update` dan `delete` tetap
  global/destruktif dan `update` masih bisa dipanggil pemanggil API yang berwenang (perilaku lama) — sebaiknya dibatasi
  ke superadmin terpisah.
- Callback memakai HTTP polos di jaringan Docker lokal; pelindungnya firewall + kunci event (yang juga tertulis di
  `handler.py` tiap profil). Gunakan kunci acak panjang.
- Token Telegram/WhatsApp tetap teks biasa di `.env` profil (desain Hermes), kini tidak lewat argv/log.
- Sandbox bridge belum dibuktikan di bawah systemd nyata (§3); Fase 5 adalah pembuktiannya.
- Nama/path unit frontend, lokasi binary, dan isi `Environment` unit lama saya tidak dapat melihatnya; unit baru berisi
  tebakan yang ditandai `# EDIT` dan harus disamakan dengan keluaran Fase 0.
