# FR — Pertanyaan tak terjawab dan kontak

Sumber: `internal/questions/questions.go`, `httpapi/questions.go`, `internal/contacts/contacts.go`, `00016_tenant_contacts.sql`, `00017_tenant_questions.sql`.

| ID | Deskripsi | Prioritas | Status | Sumber | Kriteria penerimaan |
|---|---|---|---|---|---|
| FR-QC-001 | Status pertanyaan harus salah satu dari `OPEN`, `ANSWERED`, `RESOLVED`, `IGNORED`. | Must | [CONFIRMED] | konstanta `questions.go`, `validQuestionStatus` | PATCH status `CLOSED`: 400 `INVALID_REQUEST`. |
| FR-QC-002 | Sistem harus menyimpan paling banyak 200 pertanyaan `OPEN` per tenant (yang terlama dibuang) dan paling banyak 100 jawaban yang diserahkan ke asisten. | Must | [CONFIRMED] | `MaxOpen = 200`, `MaxAnswered = 100` | Pertanyaan OPEN ke-201 menyebabkan baris OPEN terlama tidak lagi termasuk batas `LIMIT` penghapusan di `RecordUnanswered`. Jawaban ke-101 tidak menambah persona di atas batas repository. |
| FR-QC-003 | Balasan asisten yang berisi frasa `belum tersedia di data kami` harus dianggap tidak terjawab, dan kalimat baku yang diberikan ke persona harus `Mohon maaf, informasi tersebut belum tersedia di data kami.` | Must | [CONFIRMED] | `Marker`, `Sentence`, `IsUnanswered` | Balasan berisi marker tercatat sebagai pertanyaan. Balasan tanpa marker tidak lewat `IsUnanswered`. |
| FR-QC-004 | Kontak publik per tenant harus paling banyak 20. Nama peran kontak dipotong 80 karakter. Telepon dan email harus lolos normalisasi di paket `contacts` sebelum disimpan. | Must | [CONFIRMED] | `MaxPerSite = 20`, `contacts.go` `Clean` peran di `httpapi/contacts.go` | Kontak ke-21: `ErrContactLimit`. Izin tulis: `tenant:update`. VIEWER dapat GET. |

### Alur pertanyaan

**Normal.** Asisten menjawab dengan marker. `RecordUnanswered` menyimpan atau menaikkan hitungan. Admin mengubah status ke `ANSWERED` beserta jawaban. Jawaban masuk teks persona lewat `PersonaSection`.

**Normalisasi kunci.** Pertanyaan dibanding case-insensitive, tanda baca dibuang, spasi digabung, paling banyak 200 rune (`Key`). Dua kalimat yang sama setelah normalisasi itu tidak menjadi dua baris.

Daftar UI: `/ponpes/questions`.
