# Blueprint Analisis Implementasi GoCroot pada Backend Research Data Analysis

## Pendahuluan dan Ringkasan Eksekutif

Dokumen ini menyajikan rencana analisis teknis yang komprehensif untuk mengevaluasi tingkat kesesuaian implementasi backend Research Data Analysis terhadap praktik GoCroot yang ideal. Fokus penilaian diarahkan pada empat area inti: struktur codebase, sistem routing, implementasi controller, dan titik integrasi. Analisis disusun sebagai blueprint naratif yang dapat dijalankan: dimulai dari pemahaman kondisi saat ini (what), diturunkan menjadi evaluasi masalah dan gaps (how), dan diakhiri dengan rekomendasi perbaikan beserta roadmap eksekusi (so what).

Berdasarkan pembacaan terhadap artefak inti—entry point aplikasi, agregator routing, kumpulan controller, paket config, dan helper—kami menemukan beberapa karakteristik dominan yang perlu Ditata ulang. Pertama, routing dijalankan secara manual dengan switch-case pada satu fungsi pusat; pendekatan ini mudah diterapkan untuk sekadar “bekerja”, tetapi rentan terhadap duplikasi, inkonsistensi penamaan, dan sulit dikembangkan ke arah versioning atau modularisasi per domain. Kedua, controller memadukan terlalu banyak tanggung jawab: ia menangani parsing request, otentikasi, akses database, pemanggilan layanan eksternal, compose konteks untuk AI, hingga orchestrasi alur analysis dan export. Pola ini menyulitkan pengujian, meningkatkan beban kognitif, dan membuat evolusi kontrak HTTP terasa berisiko.

Ketiga, beberapa pattern kunci dari GoCroot belum diterapkan: ketiadaan DI (dependency injection) dengan wire, absennya lapisan service/usecase yang eksplisit, belum adanya error wrapper terpusat, dan konfigurasi yang masih di-setup per request tanpa struktur default/validasi yang tegas. Pada sisi integrasi, terjadi ketergantungan langsung pada helper atdb dan vertexai di lapisan controller, sementara CORS di-handle di router tanpa pipeline middleware yang jelas.

Secara bisnis, konsekuensi dari kondisi saat ini adalah peningkatan biaya pemeliharaan, terhambatnya kecepatan pengembangan fitur baru, serta risiko reliability yang tidak terdeteksi dini. secara teknis, sulit melakukan pengujian terisolasi untuk logika bisnis dan meningkatkan effort ketika harus mengganti implementasi infrastruktur (misalnya, database, AI provider, atau format export).

Rekomendasi prioritasi perbaikan meliputi: migrasi ke router berbasis gorilla/mux atau chi dengan grouping per domain; refactor controller menjadi thin handlers yang hanya mengorkestrasi use-case; introductions layers service/usecase dan repository; penerapan DI dengan google wire; error wrapper terpusat dan logging terstruktur; standarisasi struktur project per domain; serta peningkatan observability (health checks, metrik, correlation ID). Semua perubahan disusun dalam roadmap bertahap yang meminimalkan downtime dan regresi, termasuk strategi rollout aman untuk Cloud Run.

Untuk mengilustrasikan skala tantangan di setiap domain, Tabel 1 merangkum readiness kualitatif per area.

Tabel 1 — Ringkasan Readiness per Domain (indikatif)

| Domain            | Readiness | Catatan Utama                                                                 |
|-------------------|-----------|--------------------------------------------------------------------------------|
| Auth              | Rendah    | belum ada handler terverifikasi; header login/secret non-standar               |
| Project           | Sedang    | route dan sebagian controller belum terlihat; DTO perlu ditegakkan             |
| Upload            | Rendah    | ketergantungan langsung ke helper; belum jelas validasi & storage abstraction  |
| Analysis          | Rendah    | controller tebal; orchestrasi AI & DB耦合; belum ada service/usecase           |
| Export            | Sedang    | logika export masih procedural di controller                                   |
| Metadata/Preview  | Rendah    | helper used directly; belum ada kontrak DTO yang konsisten                     |

Dari perspektif operasional Cloud Run, entry point memetakan PORT dari environment dan menjalankan server HTTP dengan http.HandleFunc terhadap fungsi router utama. Tanpa middleware autentikasi terpadu, pipeline CORS di bagian router menimbulkan dua masalah: preflight handling yang tercampur dengan logika routing dan header allow-origin yang secara default jatuh ke “*” ketika origin tidak ditemukan dalam daftar,增加了 keamanan risiko untuk environment non-dev. Dengan demikian, penataan ulang routing dan middleware menjadi prioritas pertama.

### Tujuan & Nilai Bisnis

- Mengurangi biaya pemeliharaan melalui pemisahan concerns yang tegas dan eliminasi duplikasi logika.
- Meningkatkan kecepatan pengembangan fitur dengan struktur router modular, controller tipis, dan layer service/usecase yang jelas.
- Meminimalkan downtime dan regressi saat rilis melalui refactor bertahap, kontrak stabil, dan suite uji yang memadai.
- Menurunkan risiko insiden operasional dengan observability yang kuat (logging terstruktur, correlation ID, metrik, health checks) dan konfigurasi environment yang aman.

### Rekomendasi Prioritas Tinggi

- Routing: adopsi gorilla/mux atau chi, lakukan grouping per domain, gunakan middleware terpadu untuk auth, CORS, rate limiting, dan logging.
- Controller: refactor menjadi thin handlers; pindahkan business logic ke service/usecase; tegakkan kontrak Request/Response dengan DTO.
- Konfigurasi & DI: gunakan google wire untuk memisahkan konstruksi dependency; sediakan struktur default dan validasi env.
- Error Handling & Observability: error wrapper terpusat; logging terstruktur; health/readiness endpoints; metrik latency dan error rate.
- Struktur Project: susun folder per domain (router, controller, service, repository, model), batasi shared kernel untuk utilitas umum.
- Integrasi Eksternal: repository abstraction untuk DB, Facade untuk Vertex AI, dan strategi retry/timeout/circuit breaker.

## Metodologi dan Sumber Data

Analisis ini dilakukan melalui pembacaan artefak backend yang tersedia dan reprduksi alur permintaan berdasarkan kode yang teramati. Keterbatasan utama adalah tidak semua endpoint/handler ditampilkan, sehingga beberapa kesimpulan mengenai domain tertentu masih bersifat indikatif dan memerlukan validasi lanjutan.

Tabel 2 — Inventaris Artefak yang Dianalisis

| Artefak                        | Peran                                                            | Catatan Kunci                                                         |
|-------------------------------|------------------------------------------------------------------|------------------------------------------------------------------------|
| main-cloudrun.go              | Entry point; PORT dari env; registrasi router                    | Cloud Run-friendly; belum ada graceful shutdown eksplisit              |
| route/route.go                | Router utama; switch-case; CORS;SetEnv dipanggil per request     | Preload env di router; middleware pipeline belum terstruktur           |
| controller/base.go            | GetHome, NotFound, MethodNotAllowed                              | Kontrak response konsisten                                             |
| controller/analysis.go        | Analysis handlers (recommend, process, results, refine, export)  | Controller tebal; akses DB & AI langsung; belum DTO/resolver eksplisit |
| config/config.go              | Vars; SetEnv; Mongo init; CORS headers                           | Init DB per permintaan; CORS default “*” untuk origin takknown         |
| helper/at/at.go               | Utilitas HTTP, URL matching                                      | Custom matcher berbasis segment; tidak mendukung regex                 |
| helper/atdb/atdb.go           | DB helper generic                                                | Akses Mongo langsung dari controller                                  |
| helper/vertexai/vertexai.go   | Integrasi Vertex AI (Gemini)                                     | Fallback token dari env; error handling sederhana                      |
| model/model.go                | Domain models & request DTO                                      | Banyak DTO tersedia namun belum ditegakkan pola validasi/transform     |

### Artefak yang Dianalisis

- Entry point aplikasi menjalankan server HTTP dan mendelegasikan seluruh permintaan ke fungsi router utama.
- Router melakukan CORS, memuat environment variables, serta mencocokkan method dan path dengan switch-case; untuk parameter dinamis menggunakan bantuan utilitas URL matching custom.
- Controller menampilkan dua karakteristik: handler root dan notfound sederhana; dan handlers untuk analisis yang mengakses database dan Vertex AI secara langsung.
- Config menyediakan pemuatan environment variables, inisialisasi MongoDB, serta CORS headers.
- Helper at/atdb/vertexai menyederhanakan operasi HTTP, DB, dan AI, namun menciptakan ketergantungan langsung di controller.
- Model menyediakan struktur data/domain dan DTO request/response yang konsisten untuk seluruh aplikasi.

### Kriteria Evaluasi

- Struktur folder dan pemisahan concerns.
- Routing: konsistensi, modularitas, dan kemudahan versioning.
- Controller: ketipisan lapisan, kontrak DTO, dan penurunan ketergantungan ke infrastruktur.
- Konfigurasi & DI: keamanan, decoupled, dan testability.
- Observabilitas: logging, error handling, health checks.
- Cloud Run readiness: port, graceful shutdown, dan rollout aman.

## Struktur Backend Saat Ini

Secara garis besar, struktur telah memisahkan entry point, router, controller, config, dan helper. Namun, boundary antar lapisan belum tegas. Helper yang semestinya berada di infrastruktur justru digunakan langsung oleh controller, dan router menjalankan CORS dan setup env yang ideal-nya bagian dari bootstrap aplikasi.

Tabel 3 — Peta Komponen → Peran → Ketergantungan (indikatif)

| Komponen           | Peran                                                | Ketergantungan Kunci                         |
|--------------------|------------------------------------------------------|-----------------------------------------------|
| main-cloudrun.go   | Bootstrapping server                                 | route.URL, http server                        |
| route/route.go     | Routing + CORS + env loading                         | controller/*, config/*, helper/at/*           |
| controller/*       | Presentation layer + business orchestration          | helper/atdb/*, helper/vertexai/*, config/*    |
| service/usecase    | Belum ada (target refactor)                          | repository, external service (via interface)  |
| repository         | Belum ada (target refactor)                          | DB client, cache, external APIs               |
| config/*           | Env loading, DB init, CORS                           | Mongo client, env vars                        |
| helper/*           | Infrastruktur/utilities                              | HTTP, DB, AI client                           |

### Entry Point & Konfigurasi

Aplikasi membaca PORT dari environment dan default ke 8080 untuk lokal. SetEnv dipanggil dari router, bukan dari main, sehingga inisialisasi lingkungan terjadi pada setiap permintaan—kondisi yang berpotensi浪费 dan menimbulkan race Mongo init. CORS di-router memungkinkan preflight langsung退出 tanpa pipeline yang jelas, sementara keamanan header Allow-Origin ditetapkan ke “*” ketika origin tidak sesuai daftar.

### Router & Routing

Router saat ini menjalankan semua route dalam satu fungsi melalui switch-case. Parameter dinamis bervariabel seperti :id, :projectId, :uploadId, :analysisId dikelola melalui utilitas custom yang hanya mencocokkan segmen, tidak mendukung wildcard kompleks atau regex. Tidak ada middleware pipeline, versioning path, atau grouping per domain.

Tabel 4 — Ringkasan Endpoint (indikatif; perlu validasi)

| Method | Path                                   | Handler             | Catatan                                    |
|--------|----------------------------------------|---------------------|--------------------------------------------|
| GET    | /                                      | GetHome             | Root                                       |
| POST   | /auth/register                         | Register            | belum terlihat implementasi                 |
| POST   | /auth/login                            | Login               | belum terlihat implementasi                 |
| GET    | /auth/profile                          | GetProfile          | belum terlihat implementasi                 |
| POST   | /api/project                           | CreateProject       | belum terlihat implementasi                 |
| GET    | /api/project                           | GetProjects         | belum terlihat implementasi                 |
| GET    | /api/project/:id                       | GetProjectByID      | parameterized                               |
| PUT    | /api/project                           | UpdateProject       | belum terlihat implementasi                 |
| DELETE | /api/project                           | DeleteProject       | belum terlihat implementasi                 |
| POST   | /api/upload/:projectId                 | UploadData          | parameterized                               |
| GET    | /api/preview/:uploadId                 | GetDataPreview      | parameterized                               |
| GET    | /api/stats/:uploadId                   | GetDataStats        | parameterized                               |
| POST   | /api/recommend/:projectId              | GetRecommendations  | parameterized; AI & DB access               |
| POST   | /api/process                           | ProcessAnalysis     | Orchestrasi AI & DB                         |
| GET    | /api/results/:analysisId               | GetAnalysisResults  | parameterized; ownership check              |
| POST   | /api/refine/:analysisId                | RefineAnalysis      | parameterized; iterasi baru                 |
| GET    | /api/export/:analysisId                | ExportResults       | parameterized; format=pdf/csv/json          |

Catatan: beberapa endpoint membutuhkan validasi lebih lanjut karena handler belum terlihat dicontroller.

### Controller & Logic

Controller analysis memusatkan banyak tanggung jawab: otentikasi header, pemetaan parameter, akses DB, compose konteks untuk AI, orchestrasi alur analisis, hingga membangun response.没有 lapisan service/usecase dan repository yang eksplicit, sehingga pengujian unit sulit dilakukan tanpa melibatkan infrastruktur secara langsung.

### Model & DTO

Model menyediakan struktur domain dan DTO request/response. Walau demikian, aturan validasi dan transformer belum ditegakkan; banyak handler tetap melakukan decoding manual dari request dan membangun response secara langsung.

### Helper & Integrasi

Helper atdb menyediakan operasi CRUD generic yang ergonomis, namun menggoda controller untuk mengakses DB langsung. Helper vertexai menyederhanakan panggilan ke Vertex AI tetapi menambah ketergantungan konkrit di controller. Idealnya, akses DB dan AI dibungkus dalam interface dan di-inject melalui wire.

## Identifikasi Masalah

Kondisi saat ini menimbulkan beberapa masalah utama. Routing manual berbasis switch-case sulit diskalakan dan rawan inkonsistensi. Controller yang tebal menyulitkan pengujian dan menjaga kontrak HTTP. DI belum diterapkan, sehingga testability rendah dan konfigurasi tersebar. Error handling tidak terstandarisasi dan observability minim.

Tabel 5 — Matriks Issue → Akar Penyebab → Dampak → Prioritas → Usulan Perbaikan

| Issue                                   | Akar Penyebab                                                | Dampak                                     | Prioritas | Usulan Perbaikan                                                   |
|-----------------------------------------|---------------------------------------------------------------|--------------------------------------------|-----------|--------------------------------------------------------------------|
| Routing manual switch-case              | Satu fungsi pusat, tanpa router library                      | Sulit масштабировать; inkonsistensi        | Tinggi    | gorilla/mux atau chi dengan grouping per domain                    |
| CORS di router tanpa middleware pipeline| Preflight exit langsung; lack of middleware composability    | Keamanan dan manutenção sulit               | Tinggi    | Middleware CORS terpadu; preflight di pipeline                     |
| Controller tebal                        | Tidak ada lapisan service/usecase                            | Sulit ditest; risiko regresi               | Tinggi    | Refactor ke handlers tipis; pindahkan logika ke service            |
| Akses DB/AI langsung dari controller    | Helper called inline, tanpa interface                         | Coupling tinggi; sulit di-mock              | Tinggi    | Repository & Facade via interface; inject via wire                 |
| Konfigurasi per request                 | SetEnv di router; init DB on-demand                          | Potensi race/overhead                       | Sedang    | Bootstrap di main; default & validate env                           |
| Error handling ad-hoc                   | Tidak ada error wrapper                                       | Respons tidak konsisten                     | Sedang    | Centralized error wrapper                                           |
| Observability minim                     | Logging minimal; health/readiness absent                     | Sulit diagnose insiden                      | Sedang    | Structured logging; metrik; health endpoints                        |
| URL matcher terbatas                    | Custom tanpa regex                                            | Sukar mendukung pola kompleks               | Rendah    | Router library; path variable terstandar                            |
| CORS Allow-Origin “*” untuk unknown     | Fallback tanpa strict origin                                  | Risiko keamanan non-dev                     | Sedang    | Validasi strict origin; environment-based policy                    |

### Masalah Routing

Switch-case pada satu fungsi tidak масштабируем untuk banyak domain.Tidak ada grouping/versioning, sehingga penamaan path rentan tidak konsisten. Custom URL matcher tidak mendukung regex atau pola kompleks. CORS di-router menyebabkan preflight退出 dini tanpa pipeline, menyulitkan penegakan kebijakan keamanan yang seragam.

### Masalah Controller

Handlers seperti GetRecommendations, ProcessAnalysis, GetAnalysisResults, RefineAnalysis, dan ExportResults memadukan terlalu banyak peran: otentikasi, validasi, akses DB, panggilan AI, compose konteks, dan pembuatan response. без lapisan service, handler sulit diuji dan perubahan infrastruktur dapat merambat ke presentasi.

### Masalah Konfigurasi & DI

SetEnv dijalankan di router dan Mongo koneksi diinisialisasi saat diperlukan tanpa jaminan idempotensi. Secrets management tidak eksplisit, dan DI tidak diterapkan. Kondisi ini menghambat pengujian dengan mock dan memperberat konfigurasi saat startup.

### Masalah Observabilitas

Logging terstruktur, correlation ID, metrik latency/error rate, serta health/readiness endpoints belum terlihat. tanpa itu,Diagnosis masalah di produksi akan sulit, dan upaya optimasi performa tidak terukur.

## Perbandingan dengan Praktik GoCroot

Praktik GoCroot强调 структурную ясность: presentation-controller yang tipis, application/service untuk бизнес логика, domain модели, dan infrastructure untuk akses teknologi. DI dengan wire memisahkan konstruksi dependency dari bisnis, sehingga testability meningkat. Routing menggunakan library standar dengan grouping/versioning, dan middleware dideklarasikan secara eksplisit. Error handling di-wrapper terpusat agar kontrak HTTP tetap konsisten. Logging terstruktur, correlation ID, dan health checks adalah bagian integral dari operasional.

Tabel 6 — Kondisi Saat Ini vs GoCroot → Gap → Aksi Perbaikan

| Kondisi Saat Ini                                      | Praktik GoCroot                                         | Gap                            | Aksi Perbaikan                                                       |
|-------------------------------------------------------|----------------------------------------------------------|--------------------------------|----------------------------------------------------------------------|
| Router switch-case satu fungsi                        | Router library + grouping/versioning                     | Modularitas & konsistensi      | gorilla/mux atau chi; grouping per domain; versioning prefix         |
| CORS di-router tanpa middleware                       | Middleware pipeline terstruktur                          | Pipeline terpadu               | Middleware CORS, auth, logging, rate limit                           |
| Controller tebal                                      | Thin handlers + service/usecase                          | Business logic berpindah       | Refactor ke service; DTO & resolver; kontrak Request/Response tegas  |
| Helper atdb/vertexai dipanggil langsung               | Repository/Facade via interface + DI wire                | Decoupling & testability       | Interface repository & AI client; provider wire                      |
| SetEnv per request; Mongo init on-demand              | Bootstrap + env default/validate                         | Idempotensi & keamanan         | Init di main; validasi env; secrets manager                          |
| Error handling ad-hoc                                 | Centralized error wrapper                                | Konsistensi kontrak            | Wrapper dengan kode & pesan terstandar                               |
| Logging & health minimal                              | Structured logging + correlation ID + health endpoints   | Observabilitas                 | Logger terstruktur; metrik; health/readiness/liveness endpoints      |

### Struktur Folder

Target struktur adalah per domain, misalnya: auth/, project/, upload/, analysis/, export/. Setiap domain memiliki router, controller, service/usecase, repository, dan model. Utilitas umum ditempatkan di shared/ atau infrastructure/, dibatasi agar tidak mencemari domain logic.

### Routing & Middleware

Gunakan gorilla/mux atau chi. Terapkan grouping per domain, versioning melalui prefix path (misal /api/v1), dan pipeline middleware eksplisit: CORS, autentikasi, rate limiting, dan logging dengan correlation ID.

### Controller & UseCase

Handler hanya menerima request, melakukan validasi ringan, dan memanggil use-case. Response menggunakan DTO yang konsisten. Business logic dan orkestrasi dipindahkan ke service/usecase, sehingga perubahan infrastruktur tidak afects controller.

### DI & Konfigurasi

google wire digunakan untuk membangun graph dependency. Struktur env dibedakan antar lingkungan (dev/staging/prod) dengan default dan validasi. Secrets Manager menyimpan data sensitif, bukan hardcoded di repository.

### Error Handling & Observabilitas

Error wrapper memetakan error domain ke HTTP status yang konsisten. Logging menggunakan format terstruktur dengan correlation ID. metrik seperti latency dan error rate dihimpun, dan health/readiness/liveness endpoints disediakan untuk operasional Cloud Run.

## Titik Integrasi yang Perlu Diperbaiki

Integrasi perlu ditata untuk mengurangi ketergantungan langsung dan meningkatkan decoupling.

Tabel 7 — Matriks Integrasi → Perubahan → Risiko → Mitigasi

| Integrasi            | Perubahan Diperlukan                                           | Risiko                            | Mitigasi                                         |
|----------------------|----------------------------------------------------------------|-----------------------------------|--------------------------------------------------|
| Router → Controller  | Grouping & delegasi handler per domain                         | Rute patah saat migrasi           | Canary; OpenAPI kontrak; uji E2E                |
| Middleware           | Pipeline CORS, auth, logging, rate limit                       | Kebijakan tidak konsisten         | Library middleware; dokumentasi; test integrasi |
| DB/Repository        | Interface repository + wire providers                          | Coupling infrastruktur            | Mock di test; implementasi terpisah             |
| Vertex AI            | Facade interface + strategi retry/timeout/circuit breaker      | Cascade failure                   | Backoff & breaker; observabilitas               |
| Export               | Service export terpisah; format strategy                       | Perubahan affect kontrak          | DTO kontrak stabil; test regresi                |

### Middleware & Pipeline

CORS dipindah ke middleware, bukan di-router handler. Autentikasi menggunakan token yang tervalidasi, dan rate limiting dibatasi untuk endpoint kritikal. Logging menyertakan correlation ID dan payload ringkas, tanpa data sensitif.

### Database & Repository

Tambahkan lapisan repository untuk memisahkan kontrak akses data dari controller. Provider wire menyuntik dependency DB ke service/usecase. Unit test menggunakan mock repository agar cepat dan stabil.

### Vertex AI & External Services

Facade/interface untuk Vertex AI membuat panggilan dapat di-swap untuk test atau provider lain. Implementasikan retry dengan backoff, timeout, dan circuit breaker untuk mencegah cascade failure. Logging untuk 응답 AI harus dibatasi dan disensor.

### Export & Storage

Service export menangani format PDF/CSV/JSON. Jika dibutuhkan, integrasi storage (misalnya GCS) dibungkus dalam interface dan tidak dipanggil langsung dari controller.

## Perubahan yang Diperlukan (Actionable Refactoring)

Perubahan disusun untuk meminimalkan downtime dan regressi. Kontrak endpoint dijaga melalui OpenAPI; regresi dicegah melalui suite uji dan canary release.

Tabel 8 — Backlog Refactoring → Estimasi Effort → Dependensi → Prioritas

| Item Refactoring                                            | Effort (indikatif) | Dependensi                    | Prioritas |
|-------------------------------------------------------------|--------------------|-------------------------------|-----------|
| Adopsi router library (gorilla/chi)                        | Sedang             | Routing middleware            | Tinggi    |
| Grouping per domain + versioning                            | Sedang             | Router library                | Tinggi    |
| Middleware pipeline (CORS, auth, logging, rate limit)      | Sedang             | Router                        | Tinggi    |
| Refactor handlers ke thin layer                            | Tinggi             | DTO & resolver                | Tinggi    |
| Introduce service/usecase & repository                     | Tinggi             | Wire DI                       | Tinggi    |
| Wire providers untuk DB, AI, export                         | Sedang             | Service/repository            | Tinggi    |
| Error wrapper + response standard                           | Rendah             | Handler refactor              | Sedang    |
| Logging terstruktur + correlation ID                        | Sedang             | Middleware pipeline           | Sedang    |
| Health/readiness endpoints                                  | Rendah             | Router                        | Sedang    |
| Validasi env & secrets manager                              | Sedang             | Config/DI                     | Sedang    |
| Test suite (unit/integration/E2E)                           | Tinggi             | Service/repository            | Tinggi    |

### Routing

Bangun router per domain, tidak lagi di satu fungsi. Gunakan path variables dan naming konvensi. Middleware dirangkai secara eksplisit: CORS → auth → rate limit → logging → handler.

### Controller

Refactor handlers menjadi tipis: validasi input ringan, transform ke DTO, panggil use-case, dan kirim response standar. Pertimbangkan opsi DTO resolver untuk mengurangi duplikasi decode/encode.

### Service/Usecase & Repository

Pindahkan business logic ke service. Tentukan kontrak repository dan implementasi di infrastruktur. Wire DI menyediakan instance DB, AI client, dan service lain. Tulis unit test untuk service/usecase dengan mock repository.

### Konfigurasi & DI

Pindahkan SetEnv dan inisialisasi koneksi ke bootstrap (main). Gunakan wire untuk membangun graph dependency. Validasi env dan sediakan default. Simpan secrets di Secret Manager; hindari hardcode.

### Error Handling & Observabilitas

Centralized error wrapper mengonversi error domain ke kode & pesan HTTP konsisten. Logging terstruktur dengan correlation ID. Sediakan metrik latency/error rate. Tambahkan health/readiness/liveness endpoints untuk Cloud Run.

### Keamanan

Terapkan validasi strict origin untuk CORS di non-dev. Gunakan rate limiting di endpoint yang berisiko. Audit & redaksi fields sensitif dalam log. Pastikan headers keamanan (misalnya Content-Type dan lain-lain) ditangani konsisten.

## Rencana Migrasi Bertahap

Migrasi dilakukan per fase dengan Definition of Done (DoD) dan rencana rollback yang jelas. Setiap fase menyertakan uji regresi untuk mencegah penurunan kualitas.

Tabel 9 — Roadmap Fase → Aktivitas → DoD → Risiko → Rollback Plan

| Fase | Aktivitas                                                                 | DoD                                                   | Risiko Utama                              | Rollback Plan                               |
|------|---------------------------------------------------------------------------|-------------------------------------------------------|-------------------------------------------|---------------------------------------------|
| 0    | Baseline & arsitektur target                                              | Dokumen baseline & target disetujui                   | Scope creep                               | Freeze scope; review mingguan               |
| 1    | Router + middleware + endpoint mapping                                    | Semua route migrated; pipeline middleware aktif       | Rute patah sementara                      | Canary; revert via tag                      |
| 2    | Refactor handlers ke thin + DTO & error wrapper                           | ≥80% handlers refactor; test handler memadai          | Regressi logika bisnis                    | Feature flags; rollback ke versi sebelumnya |
| 3    | Service/usecase + repository + wire DI                                   | DI via wire; tidak ada konstruktor eksplisit di handler | Build kompleks                           | Pipeline CI/CD diperkuat                    |
| 4    | Observabilitas + testing                                                  | Logging & health OK; E2E smoke pass                   | Data sensitif dalam log                   | Redaksi fields; audit logging               |
| 5    | Cloud Run readiness & rollout                                             | Cold start optimal; shutdown graceful                 | Deploy belum stabil                       | Blue/green; canary; rollback otomatis       |

### Fase 0 — Persiapan

Lakukan baseline audit, tentukan arsitektur target, dan sepakati acceptance criteria. Susun backlog refactor dan peta dependensi.

### Fase 1 — Router & Middleware

Migrasikan routing ke library standar. Terapkan middleware pipeline dan pastikan endpoint mapping stabil. Validasi dengan OpenAPI.

### Fase 2 — Handler & DTO

Refactor handler ke thin layer. Standarisasi DTO & response. Centralized error wrapper untuk konsistensi.

### Fase 3 — Service & DI

Bentuk lapisan service/usecase & repository. Implementasi wire DI untuk dependency graph. Tulis unit test.

### Fase 4 — Observabilitas & Testing

Implementasi structured logging, metrik, dan health/readiness endpoints. Lengkapi pyramid tes: unit → integration → E2E.

### Fase 5 — Cloud Run Readiness & Rollout

Pastikan port binding dan shutdown graceful. Optimasi cold start. Siapkan strategi blue/green/canary, dan monitoring pra-rilis.

## Risiko dan Mitigasi

Setiap refactor membawa risiko. Tabel berikut memaparkan risiko utama dan mitigasi yang disarankan.

Tabel 10 — Risiko → Dampak → Mitigasi

| Risiko                               | Dampak                                 | Mitigasi                                                                 |
|--------------------------------------|-----------------------------------------|--------------------------------------------------------------------------|
| Regressi fungsional                  | Fitur existing rusak                    | Suite uji otomatis; canary; rollback cepat                               |
| Ketidakstabilan DI                   | Build gagal atau runtime error          | Wire providers jelas; pipeline CI/CD diperkuat                           |
| Kebocoran data via log               | Risiko keamanan & compliance            | Redaksi fields; kebijakan akses log; audit berkala                       |
| Ketidakseragaman route               | Bug antar domain                        | Pedoman naming; linting; review rute                                     |
| Ketergantungan eksternal (AI/DB)     | Cascade failure; timeout                | Retry dengan backoff; circuit breaker; observabilitas                    |
| Cold start Cloud Run                 | Latensi tinggi di awal                  | Warm-up; readiness; optimasi image                                       |
| Downtime saat deploy                 | Service tidak tersedia                  | Blue/green; canary; orchestrasi CI/CD aman                               |

### Risiko Teknis

DI dan konfigurasi yang berubah perlu pengujian menyeluruh. Perubahan route berpotensi regressi di jalur kritikal; mitigasi dengan kontrak OpenAPI dan E2E smoke test.

### Risiko Operasional

Logging harus bebas data sensitif. Cold start harus dioptimasi untuk pengalaman pengguna. Di Cloud Run, readiness dan graceful shutdown wajib.

### Risiko Organisasi

Koordinasi lintas tim penting. Dokumentasi dan onboarding membantu menyamakan pemahaman. Dengan DoD yang jelas, perubahan tetap fokus dan dapat diaudit.

## Metrik Keberhasilan & Penerimaan

Keberhasilan diukur melalui indikator kualitas, tes, performa, dan reliabilitas. Target difokuskan pada tren peningkatan, bukan angka absolut.

Tabel 11 — KPI → Target (indikatif) → Cara Ukur

| KPI                      | Target (indikatif)                              | Cara Ukur                                         |
|--------------------------|--------------------------------------------------|---------------------------------------------------|
| Kualitas arsitektur      | Penurunan anti-pattern; peningkatan konsistensi | Review kode; linting; audit arsitektur            |
| Cakupan tes              | Kenaikan signifikan                              | Laporan coverage unit/integration/E2E             |
| Defect density           | Tren menurun                                     | Tracker insiden; QA laporan                       |
| Lead time for change     | Lebih singkat                                    | Waktu dari commit ke deploy stabil                |
| Deployment frequency     | Meningkat tanpa gangguan                         | Frekuensi rilis per periode                       |
| MTTR                     | Lebih cepat                                      | Durasi pemulihan insiden                          |
| Latency p95/p99          | Stabil/ membaik                                  | Metrik latency endpoint                           |
| Error rate               | Turun dan terkendali                             | Persentase error per request                      |
| Cold start time          | Menurun                                          | Pengukuran waktu inisialisasi di Cloud Run        |

### Metrik Teknis

Peningkatan coverage tes dan penurunan jumlah anti-pattern adalah indikator kualitas. Performa endpoint, terutama pada persentil tinggi, diharapkan stabil setelah optimasi cold start dan refactor handler.

### Metrik Operasional

Jumlah insiden menurun dan MTTR membaik grâce observabilitas. Frekuensi rilis meningkat tanpa mengorbankan stabilitas, mengindikasikan proses rilis yang halus dan terkontrol.

## Lampiran

### Inventaris Endpoint (indikatif; perlu validasi)

- GET / → GetHome
- POST /auth/register → Register
- POST /auth/login → Login
- GET /auth/profile → GetProfile
- POST /api/project → CreateProject
- GET /api/project → GetProjects
- GET /api/project/:id → GetProjectByID
- PUT /api/project → UpdateProject
- DELETE /api/project → DeleteProject
- POST /api/upload/:projectId → UploadData
- GET /api/preview/:uploadId → GetDataPreview
- GET /api/stats/:uploadId → GetDataStats
- POST /api/recommend/:projectId → GetRecommendations
- POST /api/process → ProcessAnalysis
- GET /api/results/:analysisId → GetAnalysisResults
- POST /api/refine/:analysisId → RefineAnalysis
- GET /api/export/:analysisId → ExportResults

### DTO & Model (contoh, melihat model/model.go)

- Response standar dengan Status, Message, Data.
- User, Variables, Project, DataSummary, Upload.
- Recommendation, MethodResult, Figure, Analysis.
- AuditLog, LoginRequest, RegisterRequest, ProjectRequest.
- RecommendRequest, ProcessRequest, RefineRequest, ExportRequest.

### Checklist Refactoring

- Routing: adopsi router library; grouping & versioning.
- Middleware: CORS, auth, rate limit, logging terstruktur.
- Handler: thin handlers; DTO/resolver; error wrapper.
- Service/Usecase: business logic; orkestrasi.
- Repository: interface; implementasi infrastruktur.
- DI: wire providers; tanpa konstruktor eksplisit di handler.
- Observabilitas: correlation ID; metrik; health endpoints.
- Konfigurasi: env default & validate; secrets manager.
- Testing: unit/integration/E2E; coverage target.
- Keamanan: strict origin; audit log; rate limit.

## Informasi yang Belum Tersedia (untuk Validasi Lanjutan)

- Implementasi handler untuk auth, project, upload, dan sebagian export belum terlihat sehingga status dan kualitasnya perlu divalidasi.
- Pola otentikasi dan manajemen sesi pengguna (JSON Web Token/JWT atau mekanisme lain) tidak eksplisit; header Authorization belum ditangani secara utuh.
- Struktur repository layer, strategi migrations, dan pola transaksi database belum terlihat; saat ini akses DB melalui helper atdb.
- Strategi error handling yang seragam dan korelasi request-id/trace-id belum terlihat; kesehatan logging dan metrik belum nampak.
- Detail deployment pipeline (CI/CD), strategi blue/green, dan konfigurasi Cloud Run spesifik perlu dilengkapi.
- Kebutuhan audit keamanan dan kepatuhan belum teridentifikasi secara eksplisit.
- Kinerja endpoint, beban kerja, dan scaling strategy di Cloud Run belum dijabarkan.

---

Dengan blueprint ini, tim memiliki peta jalan jelas untuk mentransformasi backend Research Data Analysis menuju praktik GoCroot yang benar: modular, teruji, dan siap produksi. Kunci keberhasilannya adalah eksekusi bertahap yang disiplin, menjaga kontrak HTTP tetap stabil, serta menata lapisan dengan tegas sehingga infrastruktur dapat berevolusi tanpa menganggu presentasi atau bisnis logic.