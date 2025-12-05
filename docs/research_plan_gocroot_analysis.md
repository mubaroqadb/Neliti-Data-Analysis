# Blueprint Analisis Implementasi GoCroot di Backend Research Data Analysis

## Ringkasan Eksekutif

Analisis ini mengevaluasi kesiapan backend Research Data Analysis untuk mengadopsi kerangka GoCroot secara penuh. Fokus utama adalah pada struktur folder, praktik routing, implementasi controller, kesiapan migrasi ke injeção de dependência (DI) berbasis wire, penanganan error dan logging, konfigurasi, serta rencana refactoring yang dapat dijalankan tanpa menimbulkan downtime. Tujuan utamanya adalah menyusun peta jalan yang aman, teruji, dan terukur untuk penerapan praktik GoCroot terbaik, termasuk di lingkungan Cloud Run.

Berdasarkan konteks yang tersedia, struktur saat ini dipercaya telah memisahkan concerns secara garis besar (main entry, route, controller, config, helper). Namun, beberapa area memerlukan perbaikan agar konsisten dengan GoCroot: routing manual yang rentan duplikasi dan inkonsistensi penamaan; controller yang dirasa belum terdistribusi dengan jelas per domain; konfigurasi yang belum terlihat pola loading berbasis environment; serta belum terlihatnya penggunaan wire untuk DI. Dampak yang mungkin timbul adalah reliabilitas yang belum optimal, kompleksitas pemeliharaan meningkat, serta sulitnya pengembangan horizontal untuk domain baru. Dengan refactor bertahap, kami memperkirakan peningkatan signifikan terhadap kejelasan arsitektur, kemampuan tes, dan stabilitas release.

Rekomendasi tingkat tinggi yang diusulkan meliputi: standardisasi routing dengan grouping yang konsisten dan penamaan path yang jelas;重构 controller agar focus tetap pada orchestrasi use-case;引入 DI melalui wire untuk memisahkan构造 dependency dari bisnis; mengadopsi centralized error wrapper dan korelasi ID untuk observabilitas; serta menyusun dokumen migrasi fase per fase dengan checklist uji regresi dan rollout terkontrol.

Untuk mengilustrasikan tingkat kesiapan domain/domain-layer saat ini, tabel berikut merangkum hasil evaluasi kualitatif. Nilai readiness adalah indikasi relatif untuk memandu prioritas, bukan audit kode final.

Tabel 1 — Ringkasan Readiness per Domain (kualitatif)

| Domain            | Readiness | Catatan Utama                                                                 |
|-------------------|-----------|--------------------------------------------------------------------------------|
| Auth              | Rendah    | Pola handler-controller tidak terdokumentasi; dependency belum di-DI kan.     |
| Dataset           | Sedang    | Kemungkinan adanya logic di controller; routing manual perlu distandardisasi. |
| Analysis          | Rendah    | Batas controller-service belum jelas; error handling tidak konsisten.          |
| Visualization     | Rendah    | Logging dan context propagation belum terlihat; DI belum diterapkan.          |
| Reporting         | Rendah    | Banyak helper di layer tidak jelas; perlu facades/abstractions.                |
| User Management   | Sedang    | Ada pemisahan folder, namun kontrak Request/Response perlu diperjelas.        |

Langkah mitigasi risiko selama refactor mencakup: aktivasi feature flags, blue/green atau canary release, kontrak OpenAPI yang dikunci, serta suite uji end-to-end (E2E) untuk validasi fungsional sebelum dan sesudah migrasi. Pendekatan bertahap ini bertujuan meminimalkan gangguan operasional, khususnya di Cloud Run.

### Tujuan & Nilai Bisnis

GoCroot旨在提高研发效能和质量：在结构上实现清晰的分层与依赖倒置，在运维上实现可观测、可测试、可维护的服务。 manfaat langsung bagi tim dan bisnis: time-to-market lebih cepat karena pengembangan fitur baru dapat dilakukan secara paralel dengan kontrak yang stabil; menurunnya defect akibat konsistensi pola dan modularisasi yang tegas; serta biaya operasional lebih rendah karena pengurangan troubleshooting insidental dan reliabilitas yang lebih tinggi. Di jangka panjang, konsistensi ini juga mengurangi onboarding time bagi engineer baru, karena setiap domain mengikuti pola yang seragam.

### RekomendasiPrioritas Tinggi

Prioritas implementasi dimulai dari area yang paling memengaruhi stabilitas rilis dan biaya jangka panjang:

- Routing: standardisasi pola grouping, penamaan path, dan deklarasi method agar konsisten di seluruh domain.
- Controller: menegakkan kontrak Request/Response, memindahkan business logic ke service/usecase, serta menambahkan pengujian handler.
- Konfigurasi & DI: mengadopsi wire untuk memisahkan konstruksi dependency, mengaktifkan konfigurasi berbasis environment, dan meminimalkan hardcoding.
- Error Handling: menerapkan error wrapper terpusat, korelasi ID, serta kebijakan respon yang ramah klien.
- Logging/Observability: menerapkan struktur log terstandar, correlation ID, metrik, dan health checks.
- Testing: menambah unit test untuk handler, mock untuk service, serta E2E smoke test untuk rute-rute kritikal.

## Cakupan, Sumber Informasi, dan Metodologi

Analisis ini mencakup keseluruhan backend dengan fokus pada bagian-bagian yang diketahui: entry point aplikasi, router, controller, konfigurasi, dan helper. Metode yang digunakan adalah audit struktur folder dan peninjauan pola penggunaan di setiap lapisan, сопоставление terhadap praktik GoCroot, identifikasi gap dan risiko, serta penyusunan roadmap refactor yang dapat dieksekusi fase demi fase. Evaluasi dilakukan dengan criterios: konsistensi arsitektur, kejelasan kontrak, testabilitas, dan operasional di Cloud Run.

Keterbatasan yang kami catat sejak awal adalah ketiadaan akses terhadap isi berkas aktual. Karena itu, temuan dan rekomendasi diarahkan sebagai hipotesis terstruktur yang perlu divalidasi melalui review kode dan diskusi tim. Hal ini memungkinkan penajaman plan sebelum eksekusi di lingkungan produksi.

Tabel 2 — Inventaris Artefak Bekend (apa yang diharapkan, masih perlu divalidasi)

| Artefak                        | Peran                                                                                 | Status Ditemukan |
|-------------------------------|----------------------------------------------------------------------------------------|------------------|
| main-cloudrun.go              | Entry point; bootstrap server, inisialisasi config, register routes                    | Diharapkan       |
| route/route.go                | Agregator routing; definisi path per domain                                            | Diharapkan       |
| controller/                   | Handler per fitur; memproses HTTP, memanggil use-case                                  | Diharapkan       |
| config/                       | Pemuatan variabel environment, secrets, opsi runtime                                   | Diharapkan       |
| helper/                       | Utilitas/alat bantu lintas lapisan                                                     | Diharapkan       |

### Artefak yang Dianalisis

- main-cloudrun.go: titik awal eksekusi, memuat konfigurasi dan menyiapkan router server.
- route/route.go: pusat definisi routing; ideal-nya melakukan grouping per domain dan mendelegasikan handler ke controller.
- controller/: kumpulan handler per fitur yang meneruskan permintaan ke lapisan use-case/service.
- config/: memuat opsi runtime dan lingkungan; targets-adopsi pola berbasis environment.
- helper/: utilitas umum yang membantu pekerjaan lintas lapisan.

### Kriteria Evaluasi

- Struktur folder: kesesuaian dengan pola GoCroot dan pemisahan concerns.
- Routing: konsistensi penamaan, method, dan grouping; minimisasi duplikasi.
- Controller: kontrak Request/Response yang jelas; hanya orchestrasi, tidak ada business logic.
- DI: ketergantungan dikelola via wire, bukan construção explícita di handler.
- Konfigurasi: decoupled dari kode; fleksibel antar environment.
- Observabilitas: logging terstruktur, error wrapping, health checks.
- Testabilitas: handler dan service mudah ditest dengan mock.
- Operasional (Cloud Run): port binding, graceful shutdown, konfigurasi runtime.

## Gambaran Umum Arsitektur Bekend Saat Ini

Berdasarkan konteks, backend diharapkan memiliki struktur folder dengan pemisahan relatif antara entry point, router, controller, konfigurasi, dan helper. Entry point memulai aplikasi, router mengatur akses ke handler, controller berperan sebagai lapisan presentation yang tipis, config mengatur lingkungan, dan helper memuat utilitas yang dapat digunakan lintas lapisan. Alur permintaan umum adalah: client → router → controller → service/usecase → repository → response.

Sebelum memberikan tabel ringkas, perlu ditekankan bahwa ini adalah gambaran konseptual untuk memandu diskusi dan validasi工程师 tim.

Tabel 3 — Pemetaan Komponen → Peran → Dependensi (konseptual)

| Komponen          | Peran                                                     | Dependensi Utama                                   |
|-------------------|-----------------------------------------------------------|----------------------------------------------------|
| main-cloudrun.go  | Bootstrap aplikasi; init config, router, server           | config, route, controller, middleware, server      |
| route/route.go    | Mendefinisikan path, method, grouping, handler            | controller, middleware                             |
| controller/       | Menerima request, validasi input, panggil use-case        | service/usecase, DTO, error wrapper                |
| service/usecase   | Bisnis logic; orkestrasi data dari repository             | repository, external service, cache                |
| repository        | Abstraksi akses data; persistensi                         | database/client eksternal                          |
| config/           | Pemuatan environment dan opsi runtime                     | env vars, secrets, flag konfigurasi                |
| helper/           | Utilitas; reusable functions                              | Dapat digunakan lintas lapisan                     |

Tabel 4 — Matriks Lapisan vs Tanggung Jawab

| Lapisan                 | Tanggung Jawab Utama                                        | Apakah Sudah Terpisah (hipotesis) |
|-------------------------|--------------------------------------------------------------|-----------------------------------|
| Presentation (HTTP)     | Parsing request, validasi ringan, delegasi ke use-case       | Parsial                           |
| Application/UseCase     | Bisnis logic; orkestrasi step; kebijakan case-by-case        | Perlu diperkuat                   |
| Domain/Model            | Entitas dan aturan domain tingkat tinggi                     | Perlu tinjau ulang                |
| Infrastructure          | Akses DB, call eksternal, cache, logging                     | Parsial                           |

### Struktur Folder dan Konvensi

Penempatan entry point di root memudahkan run di CI/CD. Router sebaiknya tidak mencampur definisi handler; ia hanya mendefinisikan route dan menargetkan controller yang relevan. Konvensi penamaan path harus konsisten antar domain: gunakan bentuk jamak untuk resource, lowercase, dan penamaan yang deskriptif. Untuk versioning, gunakan prefix versi mayor (misalnya /api/v1) guna memberi ruang migrasi di masa depan.

### Alur Permintaan

Alur standar: router mengarahkan permintaan ke controller, controller men-transform request ke DTO (Data Transfer Object) yang ditunggu use-case, lalu use-case melakukan orkestrasi dan mengembalikan response DTO. Prinsip ini menjamin bahwa controller tetap tipis dan dapat ditest dengan mudah tanpa ketergantungan ke detail infrastruktur.

## Analisis Mendalam per Area

### Struktur Bekend Saat Ini

Berdasarkan konteks, file-file inti sudah ada dengan pemisahan awal. Namun, praktik GoCroot mengharuskan batas lapisan lebih tegas: presentation hanya menerima dan merespons, application berisi kasus penggunaan, domain memuat model dan invariant, serta infrastruktur menangani teknologi tertentu. Perlu penambahan layer repository untuk memperjelas abstração akses data. Kontrak antar lapisan harus eksplisit (DTO, interface), sehingga perubahan infrastruktur tidak menusuk ke lapisan atas.

### Routing Issues (Masalah Routing)

Routing manual kerap menimbulkan duplikasi, pola penamaan tidak konsisten, serta tidak adanya pengelompokan logis per domain. Method handler yang tidak seragam menyulitkan testing dan meningkatkan beban kognitif. Middleware juga mungkin tidak terdistribusi dengan benar, sehingga复用 tidak maksimal dan penempatan logic lintas concerns tercampur. Di Cloud Run, port binding dan graceful shutdown harus disetel dengan benar agar layanan tetap responsif saat scaling.

Tabel 5 — Checklist Standardisasi Routing

| Aspek               | Standar yang Disarankan                                                    | Status (hipotesis) |
|---------------------|----------------------------------------------------------------------------|--------------------|
| Method              | GET, POST, PUT, PATCH, DELETE sesuai RFC dan semantik                      | Parsial            |
| Path                | Lowercase, jamak, tanpa spasi; /api/v1 prefix untuk versi mayor            | Perlu perbaikan    |
| Penamaan Handler    | Nama fungsi jelas: GetByID, Create, Update, Delete, List                   | Perlu perbaikan    |
| Grouping            | Group per domain; router mendelegasikan ke controller                      | Parsial            |
| Middleware          | CORS, auth, rate limit terdistribusi secara konsisten                      | Perlu perbaikan    |
| Binding             | Port dan shutdown dikonfigurasi dengan benar untuk Cloud Run               | Perlu perbaikan    |

Tabel 6 — Risiko Inkonsistensi Routing vs Dampak Operasional

| Risiko                         | Dampak                                                         | Mitigasi                                        |
|-------------------------------|----------------------------------------------------------------|-------------------------------------------------|
| Duplikasi route               | Kebingungan developer; bug lintas fitur                       | Agregator route; review rutin                   |
| Penamaan tidak konsisten      | Onboarding lambat; sulit поиск implementasi                   | Pedoman penamaan; linting规则                   |
| Method tidak seragam          | Klien/pengguna bingung; kontrak tidak dapat diprediksi        | Standar method; validasi otomatis               |
| Middleware tersebar           | Kerentanan keamanan; logic berulang                           | Middleware terpusat; dokumentasi yang jelas     |
| Binding/shutdown salah        | Cold start lama; request drop saat deploy                     | health checks, graceful shutdown, uji load      |

### Controller Problems (Masalah Controller)

Controller yang mencampur бизнес logic dengan respons HTTP menyulitkan tes dan pengubahan kontrak. Batas kontrak Request/Response yang kabur mengahalangi evolusi antar versi. Banyaknya dependensi langsung di handler memperberat testability. Tanpa pengujian, regresi mudah terjadi.

Tabel 7 — Anti-Pattern di Controller vs Perbaikan yang Disarankan

| Anti-Pattern                          | Dampak                                    | Perbaikan yang Disarankan                                            |
|--------------------------------------|-------------------------------------------|-----------------------------------------------------------------------|
| Business logic di handler            | Sulit di-test; logic tersebar             | Pindahkan ke use-case; handler hanya orchestrasi                      |
| Dependensi infrastruktur langsung    | Ketergantungan erat; sukar di-mock        | Gunakan interface; injeksi via wire                                   |
| Response manual per case             | Inconsistensi format; error handling kacau| Standarisasi response DTO; error wrapper terpusat                     |
| Validasi ad-hoc                      | Potensi bug; keamanan rapuh               | Schema/validator terpusat; DTO terstruktur                            |
| Coupling ke framework细节            | Sulit migrasi; refactor mahal             | Abstraksi apresentação: adapter HTTP yang tipis                        |

### Missing Patterns (Pattern GoCroot yang Belum)

Beberapa pola kunci yang perlu diperkuat atau dihadirkan:

- Handler → Controller → Service → Repository: struktur bertahap yang tegas.
- DI via wire: memisahkan konstruksi dependency dari bisnis logic.
- Abstractions (interfaces): repositories dan service di decoupling-kan dari конкретная implementasi.
- Configuration loading: berbasis environment; secretos di luar код; flags untuk feature gating.
- Error standardization: error wrapper dan kebijakan respon terpusat.
- Observability: logging terstruktur, trace/correlation ID, health checks, metrik.
- Testing pyramid: unit → integration → E2E.

Tabel 8 — Gap Pattern vs Aksi Penutup

| Pattern                          | Status (hipotesis) | Aksi Penutup                                                                 |
|----------------------------------|--------------------|-------------------------------------------------------------------------------|
| Handler → Controller → Service   | Parsial            | Refactor handler ke thin layer; pindahkan logic ke service                    |
| Repository abstraction           | Kurang             | Introduce interfaces; implementasi di infrastructure                          |
| DI via wire                      | Kurang             | Buat provider wire; init dependencies per domain                              |
| Config berbasis env              | Kurang             | Standarisasi loading config; mendukung development/staging/production         |
| Error wrapper                    | Kurang             | Wrapper terpusat; map error domain ke HTTP                                    |
| Logging terstruktur              | Kurang             | Structured logs; correlation ID; kebijakan level dan fields                   |
| Testing pyramid                  | Kurang             | Tambah unit & integration tests; E2E smoke tests                              |

### Integration Points (Poin Integrasi)

Registrasi route ideal-nya berasal dari modul domain, bukan dari satu file monolitis. Middleware harus di-compose dengan benar: CORS, autentikasi, rate limiting, serta logging bercorelasi. Kontrak untuk DB, cache, dan service eksternal harus diekspos melalui interfaces, sehingga penggantian implementasi (misalnya untuk test) mudah dilakukan. Di Cloud Run, health checks, readiness, dan liveness endpoints harus disediakan; port binding mengikuti variabel ambiente; serta logging diarahkan ke output standar agar mudah consumed oleh platform.

Tabel 9 — Matriks Integrasi → Perubahan yang Diperlukan → Risiko → Mitigasi

| Integrasi               | Perubahan Diperlukan                                         | Risiko Utama                            | Mitigasi                                          |
|-------------------------|--------------------------------------------------------------|-----------------------------------------|---------------------------------------------------|
| Router → Controller     | Grouping per domain; kontrak handler jelas                   | Downtime karena perubahan route         | Canary release; lock kontrak via OpenAPI          |
| Middleware              | Distribusi konsisten; compose terpusat                       | Kebocoran auth; duplikasi logic         | Library middleware; pengujian integrasi           |
| DB/Repo                 | Abstraksi interface; konfigurasi via env                     | Coupling ke конкретная implementasi     | Mock di test; repository implementasi terpisah    |
| Cache                   | Interface jelas; invalidasi terstruktur                      | Data stale; kompleksitas invalidasi     | Kebijakan TTL; logging cache hit/miss             |
| External services       | Retry, timeout, circuit breaker                              | Cascade failure                         | Kebijakan retry; backoff; observabilitas          |
| Cloud Run runtime       | Port binding, health checks, graceful shutdown               | Cold start lama; request drop           | Warm-up, readiness, uji load pra-rilis            |

## Perbandingan dengan Implementasi GoCroot yang Benar

Praktik GoCroot menekankan tiga hal: struktur folder yang eksplisit per domain (bukan hanya file-type分层), DI berbasis wire yang memisahkan konstruksi dependency dari bisnis, dan standarisasi error handling, konfigurasi, serta observabilitas. Perbandingan terhadap kondisi saat ini (berdasarkan konteks) menunjukkan kesenjangan di ketiga aspek tersebut.

Tabel 10 — Kondisi Saat Ini vs Praktik GoCroot → Gap → Aksi Perbaikan

| Kondisi Saat Ini (hipotesis)                     | Praktik GoCroot                          | Gap                                   | Aksi Perbaikan                                           |
|--------------------------------------------------|------------------------------------------|---------------------------------------|----------------------------------------------------------|
| Routing manual per file                          | Grouping, konsistensi, agregator route   | Duplikasi, inkonsistensi              | Standarisasi route; validasi otomatis                    |
| Controller tebal                                 | Thin handler; service/usecase kuat       | Logic tersebar; sulit ditest          | Refactor ke use-case; DTO & kontrak eksplisit            |
| Dependensi konkret di handler                    | DI via wire                              | Testability rendah                    | Wire providers; interfaces untuk repo/service            |
| Konfigurasi hardcoding                           | Config via env dan flags                 | Kurang fleksibilitas                  | Env-based config; secretos terpisah                      |
| Error handling lokal                             | Error wrapper terpusat                   | Respons tidak konsisten               | Wrapper dan kebijakan mapping ke HTTP                    |
| Logging minimal                                  | Logging terstruktur, correlation ID      | Observabilitas rendah                 | Structured logs; metrik; health checks                   |
| Tes fokus pada level terbatas                    | Pyramid: unit→integration→E2E           | Resiko regresi tinggi                 | Tambah unit/integration/E2E; coverage target             |

### Struktur Folder

Pisahkan per domain, bukan hanya file type. Setiap domain memiliki router, controller, service, repository, dan model. shared kernel untuk utilitas lintas domain boleh ada, namun dibatasi agar tidak mendistorsi batas lapisan. Kontrak di экспортируем melalui interfaces dan DTO, bukan melalui implementasi konkret.

### Routing & Middleware

Grouping per domain dengan path dan method yang seragam. Middleware di-compose secara eksplisit dan reusable. Tidak ada duplikasi definisi route; validator dan serializer digunakan secara konsisten untuk menjaga kontrak tetap bersih.

### Controller & UseCase

Handler tipis yang menerima request, memvalidasi DTO sederhana, lalu memanggil use-case. Use-case menanggung бизнес logic, orkestrasi, dan kebijakan error domain. Response dikembalikan dalam bentuk DTO standar, sehingga perubahan infrastruktur tidak memengaruhi kontrak HTTP.

### DI & Config

Wire digunakan untuk memisahkan konstruksi dependency dari bisnis. Konfigurasi melalui environment variables dan secrets manager; flags untuk rollout fitur. Hal ini memampukan pengujian yang deterministik dan migrasi teknologi tanpa dampak luas ke lapisan atas.

### Observabilitas & Error Handling

Logging terstruktur dengan correlation ID memudahkan pelacakan transaksi lintas layanan. Error wrapper memetakan error domain ke HTTP status dan kode kesalahan yang konsisten. Health checks menampilkan status readiness dan liveness; metrik latency dan error rate memungkinkan tindakan proaktif.

## Rencana Refactoring Bertahap

Refactor dilakukan dalam lima fase berurutan, masing-masing with definition of done (DoD) dan criteria masuk/keluar yang jelas. Setiap fase menyertakan uji regresi dan opsi rollback.

Tabel 11 — Roadmap Fase per Fase: Aktivitas → Output → DoD → Risiko → Rollback Plan

| Fase | Aktivitas Utama                                                                 | Output                               | DoD                                               | Risiko                                     | Rollback Plan                                |
|------|----------------------------------------------------------------------------------|--------------------------------------|---------------------------------------------------|--------------------------------------------|----------------------------------------------|
| 0    | Dokumentasi & baseline assessment                                               | Dokumen baseline & target arsitektur | Baseline disetujui; backlog disepakati            | Scope creep                                 | Freeze scope; revisi mingguan                |
| 1    | Routing refactor & validator                                                    | Router konsisten; validator terpusat | Semua rute baru melewati review; OpenAPI update   | Rute patah sementara                        | Canary; revert cepat via tags                |
| 2    | Controller → UseCase split; DTO; error wrapper                                  | Handler tipis; use-case terstruktur  | 80% handler refactor; unit test memadai           | Regressi logika bisnis                      | Feature flags; rollback ke versi sebelumnya  |
| 3    | Wire DI; config berbasis env                                                    | Provider wire; config/env stabil     | DI lewat wire; tidak ada构造 explícita di handler | Build复杂度 meningkat                       | Build pipeline diperkuat; fallback manual    |
| 4    | Observability & testing                                                         | Logging terstruktur; pyramid tes     | Metrik & health ok; E2E smoke pass                | Data sensitive dalam log                    | Redaksi fields; audit logging                |
| 5    | Cloud Run readiness                                                             | Port, health, readiness siap         | Cold start optimal; graceful shutdown             | Deploy belum stabil                         | Blue/green; canary; rollback otomatis        |

### Fase 0 — Persiapan

Lakukan baseline audit struktur dan praktik saat ini; definisikan arsitektur target dan acceptance criteria. Susun backlog refactor yang terukur dan peta dependensi antar perubahan. Komunikasikan perubahan ini kepada seluruh pemangku kepentingan agar ekspektasi rilis terjaga.

### Fase 1 — Routing & Middleware

Standardisasi definisi route dan middlewares. Validasi input di edge. Perbaiki binding port dan shutdown handler agar bekerja optimal di Cloud Run. Lakukan review lintas domain untuk memastikan konsistensi.

### Fase 2 — Controller & UseCase

Pindahkan бизнес logic ke use-case. Standarisasi Request/Response DTO dan error handling via wrapper. Tambah unit test untuk use-case dan handler, serta pengujian integrasi untuk memastikan interaksi dengan repository sesuai.

### Fase 3 — Config & DI

Refactor konfigurasi ke environment-based. Implementasi wire untuk DI. Pisahkan secrets dari kode. Pastikan pengujian yang melakukan mock dapat berjalan tanpa ketergantungan ke infrastruktur konkret.

### Fase 4 — Observabilitas & Testing

Implementasi logging terstruktur dengan correlation ID. Tambah metrics dan pelaporan health. Lengkapi pyramid tes: unit untuk logic, integration untuk kontrak infrastruktur, dan E2E untuk skenario kritikal.

### Fase 5 — Cloud Run Readiness

Pastikan endpoint health, readiness, dan liveness tersedia. Optimasi cold start dan shutdown. Siapkan strategi rollout: blue/green atau canary dengan feature flags agar rilis aman.

## Risiko, Ketergantungan, dan Mitigasi

Perubahan arsitektur pasti membawa risiko. Tabel berikut merangkum risiko utama, dampaknya, dan mitigasi yang disarankan.

Tabel 12 — Risiko vs Dampak vs Mitigasi

| Risiko                             | Dampak                                        | Mitigasi                                                   |
|-----------------------------------|-----------------------------------------------|------------------------------------------------------------|
| Regressi fungsional               | Fitur现有 rusak; kepercayaan turun            | Suite uji otomatis; canary; rollback cepat                 |
| Ketidakstabilan DI                | Build gagal; runtime error                    | Wire providers jelas; build pipeline diperkuat             |
| Kebocoran data via log            | Risiko keamanan; compliance terganggu         | Redaksi fields; kebijakan akses log                        |
| Ketidakseragaman route            | Bug antar domain; pemahaman tim memburuk      | Pedoman tertulis; linting; review rute                     |
| Ketergantungan eksternal          | Cascade failure; timeout                      | Retry dengan backoff; circuit breaker; observabilitas      |
| Cold start Cloud Run              | Latensi tinggi di awal                        | Warm-up; readiness; optimasi image                         |
| Downtime saat deploy              | Service unavailable                           | Blue/green; canary; orchestrasi CI/CD aman                 |

### Risiko Teknis

Masalah DI dan konfigurasi dapat menimbulkan error pada saat runtime. Perubahan route yang tidak teruji menimbulkan potensi regressi pada jalur kritis. Mitigasi prioritas adalah memperkuat suite uji otomatis, terutama untuk handler, use-case, dan integrasi repository, sehingga setiap perubahan dapat divalidasi sebelum menyentuh produksi.

### Risiko Operasional

Keamanan logging perlu mendapat perhatian: jangan menyimpan data sensitif. Di sisi operasional, latensi cold start dan drain connections saat scaling harus diuji dan dioptimalkan. Di Cloud Run, pengaturan readiness/liveness, port binding, dan graceful shutdown adalah komponen kunci.

### Risiko Organisasi

Koordinasi antar tim dan perubahan proses perlu dikelola. Komunikasi rilis, dokumentasi, dan sesi onboarding membantu menyamakan pemahaman. Dengan definisiDoD yang jelas per fase, perubahan tetap fokus dan dapat diaudit.

## Metrik Keberhasilan & Penerimaan

Keberhasilan refactor diukur melalui kombinasi metrik kualitas, performa, dan reliabilitas. Target yang diusulkan harus realistis dan terukur; definisi penerimaan berfokus pada tren yang konsisten，而不是 angka absolut.

Tabel 13 — KPI & Target (pra vs pasca refactor)

| KPI                         | Deskripsi                                        | Target (indikatif)                         |
|----------------------------|--------------------------------------------------|--------------------------------------------|
| Kualitas kode              | Skor lint/standar; konsistensi arsitektur        | Peningkatan jelas pasca refactor            |
| Cakupan tes                | Unit/integration/E2E coverage                     | Peningkatan yang signifikan                |
| Defect density             | Jumlah defect per modul                           | Tren menurun konsisten                     |
| Lead time for change       | Waktu dari commit ke deploy stabil                | Lebih singkat setelah standardisasi         |
| Deployment frequency       | Frekuensi rilis tanpa gangguan                    | Meningkat dengan kontrol risiko             |
| Mean time to recovery      | Waktu pemulihan insiden                           | Lebih cepat grâce observabilitas           |
| Latency p95/p99            | Waktu respons pada persentil tinggi               | Stabil/ber手下 setelah optimasi cold start  |
| Error rate                 | Persentase error per request                      | Turun dan terkendali                        |
| Cold start time            | Waktu inisialisasi di Cloud Run                   | Menurun setelah optimasi dan warm-up        |

### Metrik Teknis

Cakupan tesunit/integration/E2E yang meningkat, coupled dengan jumlah anti-pattern yang menurun, adalah indikator perbaikan kualitas. Performa endpoint, terutama di persentil tinggi, harus stabil atau membaik setelah optimasi cold start dan refactor handler.

### Metrik Operasional

Kejadian incident berkurang, sementara meantime to recovery juga menurun grâce observabilitas. Frekuensi rilis meningkat tanpa mengorbankan stabilitas, yang menandakan proses rilis yang lebih halus dan terkontrol.

## Lampiran

### Checklist Migrasi per Fase

Tabel 14 — Checklist per Fase dengan Status, Penanggung Jawab, dan Batas Waktu

| Fase | Item Checklist                                                     | Status | Penanggung Jawab | Batas Waktu |
|------|--------------------------------------------------------------------|--------|------------------|-------------|
| 0    | Baseline audit & target arsitektur disetujui                       |        | Tech Lead        | TBC         |
| 1    | Standar route & validator aktif                                    |        | Backend Lead     | TBC         |
| 2    | 80% handler refactor ke use-case                                   |        | Domain Owner     | TBC         |
| 3    | DI via wire & config/env jadi standar                              |        | DevOps/Backend   | TBC         |
| 4    | Logging terstruktur & pyramid tes tersedia                         |        | QA/Backend       | TBC         |
| 5    | Health/readiness & rollout canary siap                             |        | DevOps           | TBC         |

### Daftar Artefak yang akan Dibuat/Diubah (Target Saja)

- Provider wire (untuk DI).
- DTO dan kontrak Request/Response per domain.
- Middleware yang terstandarisasi (CORS, auth, rate limit, logging).
- Router per domain dengan grouping konsisten.
- Health/readiness endpoints.
- Dokumentasipedoman naming & routing.

### Glosarium Pattern

- Handler: lapisan paling tipis yang menerima request HTTP dan mendelegasikan ke use-case.
- Controller (dalam konteks ini): padanan handler; istilah ini digunakan untuk menyebut lapisan presentation.
- UseCase: komponen yang menanggung бизнес logic dan orkestrasi.
- Repository: abstraksi akses data; memisahkan细节DB dari aplikasi.
- DTO: objek untuk Transfer Data; mengisolasi kontrak HTTP dari model domain.
- Error Wrapper: mekanisme standarisasi error agar respons konsisten.
- DI (Dependency Injection) via Wire: kerangka untuk menyuntik dependensi secara deklaratif.
- Health Checks: endpoint yang menunjukkan kesiapan layanan (readiness/liveness).

## Informasi yang Belum Tersedia (untuk Validasi Lanjutan)

- Isi aktual berkas backend (main-cloudrun.go, route/route.go, controller/, config/, helper/) belum dibaca sehingga temuan bersifat hipotetis dan perlu divalidasi.
- Diagram alur request end-to-end belum tersedia.
- Daftar endpoint, skema data, kontrak Request/Response, dan status kode belum terdokumentasi.
- Detail pola error handling dan logging yang diterapkan belum terlihat.
- Struktur basis data, konfigurasi secrets, dan variabel environment belum diketahui.
- Kebijakan pengujian dan cakupan tes saat ini belum tersedia.
- Rincian deployment (CI/CD, strategi migrasi, batas waktu) belum tersedia.
- Kebutuhan keamanan dan kepatuhan spesifik belum teridentifikasi.

---

Dengan dokumen ini, tim memiliki peta jalan menyeluruh untuk menata ulang backend agar selaras dengan praktik GoCroot yang benar. Hal penting berikutnya adalah memvalidasi asumsi di atas melalui pembacaan kode aktual, menyepakati arsitektur target dan standar yang akan diadopsi, lalu mengeksekusi refactor bertahap sembari menjaga stabilitas produksi. Pendekatan ini memaksimalkan manfaat—struktur lebih bersih, kualitas lebih baik, operasional lebih handal—tanpa mengorbankan waktu-to-market.# Analisis Implementasi GoCroot pada Backend Research Data Analysis: Current Structure, Routing Issues, Controller Problems, Missing Patterns, dan Integration Points

## Ringkasan Eksekutif

Penilaian menyeluruh terhadap backend Research Data Analysis menunjukkan fondasi yang solid dari segi organisasi folder dan ketersediaan modul inti, namun menemukan sejumlah ketidaksesuaian terhadap praktik GoCroot yang benar. Tiga masalah inti muncul: sistem routing yang sepenuhnya manual dan tidak terstruktur (fungsi URL dengan switch-case dan parser path kustom), controller yang menjalankan bisnis logika, integrasi database dan AI secara langsung tanpa lapisan service/usecase, serta ausencia lapisan repository untuk abstraksi akses data. Selain itu, injeksi dependensi (DI) belum diimplementasikan, error handling tidak standar, observabilitas minim, dan beberapa praktik keamanan (khususnya CORS dan manajemen secrets) memerlukan penyesuaian agar sesuai untuk produksi Cloud Run.

Dampak langsung dari kondisi ini adalah meningkatnya biaya pemeliharaan, perlambatan pengembangan, risiko regresi yang lebih tinggi, serta ketahanan operasional yang belum optimal. Jika dibiarkan, akumulasi deuda teknis akan menambah friksi saat penambahan domain baru, memperberat esforço onboarding engineer, dan meningkatkan potensi insiden di produksi. Untungnya, kelemahan-kelemahan tersebut dapat ditargetkan melalui refactoring bertahap yang terstruktur.

Rekomendasi prioritas meliputi: migrasi ke router berbasis library standar (gorilla/mux atau chi) dengan grouping per domain dan versioning path; refactor controller menjadi thin handlers yang terikat kontrak DTO dan mendelegasikan bisnis logika ke service/usecase; introduce lapisan repository dan facade untuk integrasi eksternal (database MongoDB dan Vertex AI); menerapkan DI dengan google wire; menyeragamkan error wrapper dan logging terstruktur; serta memperkuat observability (health checks, readiness, liveness) dan pipeline middleware. Perubahan akan disusun dalam roadmap lima fase yang menjaga stabilitas rilis dan meminimalkan downtime melalui strategi canary/blue-green, lock kontrak OpenAPI, dan pengujian regresi end-to-end (E2E).

Sebagai ilustrasi, Tabel 1 menyajikan ringkasan readiness per domain berdasarkan bukti yang terlihat dan indikasi yang wajar dari modul yang belum diamati.

Tabel 1 — Ringkasan Readiness per Domain (indikatif)

| Domain         | Readiness | Catatan                                                                                   |
|----------------|-----------|--------------------------------------------------------------------------------------------|
| Auth           | Rendah    | Handler belum terlihat; header non-standar (Login/Secret); belum DI; belum validator.     |
| Dataset/Upload | Rendah    | UploadData/GetDataPreview/GetDataStats belum terlihat; storage facade belum terlihat.     |
| Analysis       | Rendah    | Controller tebal; akses DB/AI langsung; belum service/usecase; error handling ad-hoc.     |
| Visualization  | Rendah    | Stats/preview indikatif; belum jelas pemisahan concerns dan facade.                       |
| Reporting      | Rendah    | Export Resultados ditemukan, tetapi masih procedural; belum DI dan strategi export formal.|
| User Mgmt      | Sedang    | Project CRUD indikatif; belum terlihat implementasi controller; konsistensi DTO needed.   |

Nilai readiness bersifat indikatif untuk memandu prioritas refactor. Temuan kunci di tingkat struktur dan routing dikonfirmasi oleh artefak yang tersedia; untuk handler yang belum terlihat, penilaian konservatif diberikan untuk menghindari asumsi berlebih.

## Ruang Lingkup, Sumber Data, dan Metodologi

Analisis ini mencakup backend Research Data Analysis dengan fokus pada entry point, router, controller, konfigurasi, dan helper. Metode yang digunakan adalah pembacaan sistematis terhadap kode yang tersedia, inventarisasi endpoint, pemetaan dependensi langsung, dan penyusunan rencana refactoring bertahap. Keterbatasan utama adalah beberapa handler belum terlihat, sehingga detail fungsionalnya memerlukan validasi lanjutan sebelum pelaksanaan perubahan skala besar.

Tabel 2 — Inventaris Artefak Backend

| Artefak                   | Peran                                                     | Status Ditemukan |
|---------------------------|-----------------------------------------------------------|------------------|
| main-cloudrun.go         | Entry point; PORT dari env; registrasi route              | Ditemukan        |
| route/route.go           | Router manual; CORS; SetEnv di setiap request             | Ditemukan        |
| controller/base.go       | GetHome/NotFound/MethodNotAllowed                         | Ditemukan        |
| controller/analysis.go   | Analysis domain handlers                                  | Ditemukan        |
| controller/export.go     | ExportResults handler                                     | Ditemukan        |
| config/config.go         | Vars; SetEnv; Mongo init; CORS header                     | Ditemukan        |
| helper/at/at.go          | HTTP util; URL matcher kustom                             | Ditemukan        |
| helper/atdb/atdb.go      | DB helper generic (CRUD, count, sort)                     | Ditemukan        |
| helper/vertexai/vertexai.go | Integrasi Vertex AI (Gemini)                           | Ditemukan        |
| model/model.go           | Domain models & request DTO                               | Ditemukan        |

Keterbatasan dan Informasi yang Belum Tersedia:
- Implementasi handler untuk auth, project, dan upload belum terlihat secara lengkap.
- Pola otentikasi dan manajemen sesi (JWT/header kustom) belum dapat dikonfirmasi.
- Lapisan repository belum tersedia; saat ini akses data via helper atdb.
- Desain error handling terpusat, trace ID/correlation ID, dan logging terstruktur belum nampak.
- Rincian deployment pipeline (CI/CD), strategi blue/green, dan konfigurasi Cloud Run spesifik belum terlihat.
- Kebutuhan audit keamanan dan kepatuhan belum terdefinisi.
- Kinerja endpoint dan scaling strategy belum didokumentasikan.

### Artefak yang Dianalisis

- Entry point memetakan PORT dari environment dan menyiapkan HTTP server, kemudian registrasi router utama.
- Router berfungsi sebagai satu-satunya pintu masuk untuk seluruh request, menangani CORS dan inisialisasi environment per permintaan.
- Controller mendefinisikan responses dasar dan memusatkan logika domain analisis serta export hasil.
- Config memuat variabel lingkungan, CORS policy, dan koneksi MongoDB.
- Helper menyediakan utilitas HTTP, parser URL kustom, DB helper generic, dan integrasi AI Vertex (Gemini).
- Model mendefinisikan entitas domain dan DTO request/response.

### Kriteria Evaluasi

- Struktur folder dan pemisahan concerns.
- Routing: konsistensi, modularitas, versioning, dan middleware.
- Controller: ketipisan, kontrak DTO, validasi terpusat, dan pengujian.
- Konfigurasi & DI: keamanan, decoupled, idempotensi, dan testability.
- Observabilitas: logging, error handling, health checks, metrik.
- Operasional Cloud Run: port binding, graceful shutdown, dan rollout aman.

## Gambaran Umum Arsitektur Backend Saat Ini

Kondisi saat ini memisahkan beberapa concerns secara garis besar: main entry, router, controller, config, helper, dan model. Namun, batas lapisan belum tegas: controller menangani akses DB dan AI langsung, router mencampur CORS dan setting environment, helper digunakan inline tanpa kontrak interface, dan tidak ada DI untuk memudahkan pengujian.

Tabel 3 — Komponen → Peran → Ketergantungan (indikatif)

| Komponen           | Peran                                         | Ketergantungan Utama                         |
|--------------------|-----------------------------------------------|----------------------------------------------|
| main-cloudrun.go   | Bootstrapping server                           | route, http server                            |
| route/route.go     | Routing + CORS + env loading                   | controller, config, helper/at                 |
| controller/*       | Presentation + bisnis logika                   | helper/atdb, helper/vertexai, config          |
| service/usecase    | Belum ada (target refactor)                    | repository (target), external services        |
| repository         | Belum ada (target refactor)                    | DB client (Mongo), cache, external API        |
| config/*           | Env loading, DB init, CORS policy              | env vars, Mongo client                        |
| helper/*           | Infrastruktur/utilities                        | HTTP, DB, AI client                           |
| model/*            | Domain models & DTO                            | —                                            |

Tabel 4 — Lapisan → Tanggung Jawab → Status Kesesuaian (indikatif)

| Lapisan            | Tanggung Jawab Utama                               | Status Kesesuaian      |
|--------------------|-----------------------------------------------------|------------------------|
| Presentation (HTTP)| Parsing request, validasi ringan, delegasi ke use-case| Parsial (controller tebal) |
| Application        | Bisnis logika, orkestrasi, kebijakan                | Kurang (belum ada)     |
| Domain             | Entitas & aturan domain tingkat tinggi              | Parsial (DTO tersedia) |
| Infrastructure     | Akses DB, AI eksternal, cache, logging              | Parsial (helper inline)|

### Struktur Folder & Konvensi

Penempatan entry point di root memudahkan eksekusi di CI/CD. Konvensi path yang baik untuk GoCroot adalah: lowercase, jamak untuk resource, versioning dengan prefix mayor (misal /api/v1), dan penamaan handler yang jelas (List, GetByID, Create, Update, Delete). Route aggregator harus mendelegasikan ke controller per domain, bukan mencampur definisi handler dalam satu fungsi.

### Alur Permintaan

Alur yang benar: client → router → controller → service/usecase → repository → response. Prinsip ini memastikan controller tetap tipis, mudah ditest, dan tidak acoplado dengan detail infrastruktur. Perubahan infrastruktur tidak akan memengaruhi kontrak HTTP dan sebaliknya.

## Analisis Mendalam per Area

### Struktur Backend Saat Ini

Fungsi router pusat menangani CORS dan memuat environment variables pada setiap request, sementara SetEnv di config memicu inisialisasi koneksi MongoDB saat diperlukan. Pendekatan ini berisiko menimbulkan duplikasi effort, overhead, dan potensi race condition.oltre itu, tidak ada DI yang memisahkan konstruksi dependency dari bisnis logic.

### Routing Issues

Routing sepenuhnya manual menggunakan switch-case pada satu fungsi pusat, tanpa library router. Custom URL matcher hanya memeriksa kesesuaian segmen path tanpa mendukung regex atau pola kompleks. CORS di-router berpotensi mencampur logika routing dengan kebijakan keamanan, dan jika origin tidak ditemukan, fallback ke “*” meningkatkan risiko di lingkungan non-dev. Middleware pipeline tidak terstruktur: auth, rate limit, dan logging tidak memiliki tempat yang jelas.

Tabel 5 — Ringkasan Endpoint (indikatif, perlu validasi)

| Method | Path                               | Handler            | Catatan                                     |
|--------|------------------------------------|--------------------|---------------------------------------------|
| GET    | /                                  | GetHome            | Root endpoint                               |
| POST   | /auth/register                     | Register           | Handler belum terlihat                      |
| POST   | /auth/login                        | Login              | Handler belum terlihat                      |
| GET    | /auth/profile                      | GetProfile         | Handler belum terlihat                      |
| POST   | /api/project                       | CreateProject      | Handler belum terlihat                      |
| GET    | /api/project                       | GetProjects        | Handler belum terlihat                      |
| GET    | /api/project/:id                   | GetProjectByID     | Parser path kustom                          |
| PUT    | /api/project                       | UpdateProject      | Handler belum terlihat                      |
| DELETE | /api/project                       | DeleteProject      | Handler belum terlihat                      |
| POST   | /api/upload/:projectId             | UploadData         | Handler belum terlihat                      |
| GET    | /api/preview/:uploadId             | GetDataPreview     | Handler belum terlihat                      |
| GET    | /api/stats/:uploadId               | GetDataStats       | Handler belum terlihat                      |
| POST   | /api/recommend/:projectId          | GetRecommendations | Controller tebal; akses DB & AI langsung    |
| POST   | /api/process                       | ProcessAnalysis    | Controller tebal; orkestrasi AI             |
| GET    | /api/results/:analysisId           | GetAnalysisResults | Controller tebal                            |
| POST   | /api/refine/:analysisId            | RefineAnalysis     | Controller tebal                            |
| GET    | /api/export/:analysisId            | ExportResults      | Export procedural                           |

Tabel 6 — Risiko Inkonsistensi Routing vs Dampak Operasional

| Risiko                               | Dampak                                         | Mitigasi                                      |
|--------------------------------------|------------------------------------------------|-----------------------------------------------|
| Router manual switch-case            | Sulit diskalakan; rawan duplikasi              | Library router; grouping; linter              |
| URL matcher tanpa regex              | Sukar support pola kompleks                    | Router standar; variabel path terstruktur     |
| CORS di router                       | Kebijakan keamanan tercampur                   | Middleware CORS; validasi strict origin       |
| Fallback CORS ke “*” (origin unknown)| Risiko di environment non-dev                  | Environment-based policy; whitelist yang tegas|
| Middleware pipeline tidak jelas      | Auth, rate limit, logging tidak konsisten      | Pipeline eksplisit; dokumentasi               |
| SetEnv per request                   | Overhead; potensi race inisialisasi DB         | Env loading saat bootstrap; idempotensi       |

### Controller Problems

Controller memadukan terlalu banyak tanggung jawab: autentikasi, akses DB, compose konteks AI, orkestrasi analisis, serta pembuatan response. Tanpa lapisan service/usecase, pengujian unit sulit dilakukan. Dependensi konkret ke helper atdb dan vertexai mempererat coupling dan menghalangi penggunaan mock. Response dibuat manual per case tanpa error wrapper terpusat.

Tabel 7 — Anti-Pattern Controller → Dampak → Perbaikan

| Anti-Pattern                                | Dampak                                 | Perbaikan                                           |
|---------------------------------------------|----------------------------------------|-----------------------------------------------------|
| Bisnis logika di handler                    | Sulit test; regresi mudah terjadi      | Pindahkan ke use-case/service                        |
| Akses DB/AI langsung via helper             | Coupling tinggi; testability rendah    | Repository & facade interface; DI via wire           |
| Response/error handling manual              | Respons tidak konsisten                | Error wrapper terpusat; response DTO standar         |
| Validasi ad-hoc                             | Potensi bug; keamanan rapuh            | Validator terpusat; DTO berbasis schema              |
| SetEnv di router/inline                     | Overhead; inisialisasi DB berulang     | Bootstrap di main; validasi env terpusat             |

### Missing Patterns (Pola GoCroot yang Belum)

- Handler → Controller → Service → Repository: struktur bertahap belum diterapkan.
- DI via wire: belum ada; konstruktor dependency eksplisit di handler.
- Configuration loading: env-based config tanpa struktur default/validasi tegas.
- Error standardization: belum ada wrapper terpusat.
- Logging/Observability: belum tampak structured logging, correlation ID, metrik.
- Testing pyramid: belum terlihat bukti unit/integration/E2E.

Tabel 8 — Gap Pattern → Status → Aksi

| Pattern                      | Status (indikatif) | Aksi Perbaikan                                 |
|------------------------------|--------------------|-----------------------------------------------|
| Routing standar + grouping   | Kurang             | gorilla/chi; grouping; versioning             |
| Thin handler + DTO           | Kurang             | Refactor; DTO/resolver; kontrak eksplisit     |
| Service/Usecase layer        | Kurang             | Buat service; pindahkan bisnis logika         |
| Repository abstraction       | Kurang             | Interface repo; implementasi infra            |
| DI via wire                  | Kurang             | Wire provider; graph dependency                |
| Error wrapper                | Kurang             | Standarisasi kode & pesan error               |
| Logging terstruktur          | Kurang             | Structured logs; correlation ID; metrik        |
| Health checks                | Kurang             | Readiness/liveness endpoints                   |
| Testing pyramid              | Kurang             | Unit/service/integration/E2E; coverage target  |

### Integration Points

Registrasi route saat ini tersentralisasi di satu fungsi; ideal-nya router modular per domain. Middleware pipeline belum terstruktur. DB diakses langsung via helper; belum facade resmi. Vertex AI dipanggil langsung dari controller; tanpa strategi error yang kompleks. Export dilakukan secara procedural.

Tabel 9 — Integrasi → Perubahan → Risiko → Mitigasi

| Integrasi          | Perubahan Diperlukan                              | Risiko                    | Mitigasi                              |
|--------------------|---------------------------------------------------|---------------------------|---------------------------------------|
| Router             | Modul per domain; middleware pipeline             | Rute patah                | Canary; OpenAPI; E2E                  |
| DB/Repo            | Interface + provider wire                         | Coupling infra            | Mock di test; implementasi terpisah   |
| Vertex AI          | Facade + retry/backoff/circuit breaker            | Cascade failure           | Observabilitas; strategi ulang        |
| Export             | Service export; format strategy                   | Perubahan affect kontrak  | DTO stabil; test regresi              |

## Perbandingan dengan GoCroot yang Benar

GoCroot menuntut pemisahan tegas: domain-oriented folder, router library dengan grouping/versioning, controller tipis dengan DTO kontrak, service/usecase untuk bisnis logika, repository untuk akses data, DI via wire, error wrapper terpusat, dan logging terstruktur.

Tabel 10 — Kondisi Saat Ini vs GoCroot → Gap → Aksi

| Kondisi Saat Ini                         | Praktik GoCroot                       | Gap                        | Aksi Perbaikan                                  |
|-----------------------------------------|---------------------------------------|----------------------------|-------------------------------------------------|
| Router switch-case                      | Router library + grouping              | Modularitas                | gorilla/chi; versioning                         |
| Controller tebal                        | Thin handler + service/usecase         | Bisnis logika berpindah    | Refactor ke service                             |
| Helper DB/AI inline                     | Repository/facade + DI wire            | Decoupling & testability   | Interface + wire provider                       |
| SetEnv per request                      | Env loading saat bootstrap             | Idempotensi & keamanan     | Bootstrap; validasi env                         |
| Error handling manual                   | Error wrapper terpusat                 | Konsistensi respon         | Wrapper; mapping domain error                   |
| Logging minimal                         | Logging terstruktur & correlation ID   | Observabilitas             | Structured logs; metrik; health endpoints       |
| Tes tidak jelas                         | Pyramid tes                            | Kualitas & regresi         | Unit/integration/E2E                            |

### Struktur Folder

Domain-oriented folder (auth, project, upload, analysis, export) memusatkan router, controller, service, repository, dan model per domain. Shared kernel dibatasi untuk utilitas umum lintas domain.

### Routing & Middleware

Penggunaan library standar memfasilitasi grouping, middleware pipeline eksplisit (CORS, auth, rate limit, logging), dan konsistensi route naming. Tidak ada duplikasi definisi; validator dan serializer digunakan seragam.

### Controller & UseCase

Handler hanya menerima request, memvalidasi input ringan, dan memanggil use-case. Response menggunakan DTO standar, sehingga perubahan infrastruktur tidak afectará kontrak HTTP.

### DI & Config

google wire memisahkan konstruksi dependency dari bisnis. Konfigurasi melalui environment variables dan secrets manager; flags untuk rollout fitur. Pengujian dapat deterministik dengan mock.

### Observabilitas & Error Handling

Logging terstruktur dengan correlation ID memudahkan pelacakan. Error wrapper memetakan error domain ke HTTP status. Health checks dan metrik memungkinkan observabilitas operasional yang memadai.

## Rencana Refactoring Bertahap

Refactor bertahap disusun untuk menjaga stabilitas rilis, meminimalkan downtime, dan mencegah regresi.

Tabel 11 — Roadmap Fase → Aktivitas → DoD → Risiko → Rollback

| Fase | Aktivitas Utama                                  | DoD                                            | Risiko                    | Rollback Plan                           |
|------|--------------------------------------------------|------------------------------------------------|---------------------------|------------------------------------------|
| 0    | Baseline assessment & arsitektur target          | Baseline & backlog disetujui                   | Scope creep               | Freeze scope; revisi mingguan            |
| 1    | Routing & middleware                             | Router konsisten; OpenAPI update               | Rute patah                | Canary; revert cepat via tags            |
| 2    | Handler → UseCase; DTO; error wrapper            | ≥80% handler refactor; unit test memadai       | Regressi logika           | Feature flags; rollback                  |
| 3    | Wire DI; config/env                              | DI via wire; tidak ada konstruktor eksplisit   | Build kompleks            | Pipeline diperkuat; fallback manual      |
| 4    | Observability & testing                          | Logging & health ok; E2E smoke pass            | Data sensitif dalam log   | Redaksi fields; audit logging            |
| 5    | Cloud Run readiness                              | Cold start optimal; graceful shutdown          | Deploy belum stabil       | Blue/green; canary; rollback otomatis    |

### Fase 0 — Persiapan

Audit baseline struktur dan praktik; definisikan arsitektur target dan acceptance criteria; susun backlog refactor dan peta dependensi antar perubahan. Koordinasi lintas tim untuk menyamakan pemahaman dan ekspektasi rilis.

### Fase 1 — Routing & Middleware

Standardisasi definisi route, grouping per domain, dan middleware pipeline eksplisit (CORS, auth, rate limit, logging). Perbaiki binding port dan shutdown handler untuk Cloud Run. Review lintas domain untuk konsistensi.

### Fase 2 — Controller & UseCase

Pindahkan bisnis logika ke use-case; tegakkan Request/Response DTO dan error wrapper; tambah unit test untuk use-case dan handler; pastikan integrasi repository sesuai.

### Fase 3 — Config & DI

Refactor konfigurasi ke environment-based; implementasi wire untuk DI; pisahkan secrets dari kode; pastikan pengujian mock dapat berjalan tanpa ketergantungan infrastruktur konkret.

### Fase 4 — Observabilitas & Testing

Implementasi logging terstruktur dengan correlation ID; tambahkan metrics dan health endpoints; lengkapi pyramid tes untuk menjaga kualitas.

### Fase 5 — Cloud Run Readiness

Siapkan health/readiness/liveness; optimasi cold start dan shutdown; siapkan strategi rollout (blue/green/canary) dengan feature flags.

## Risiko, Ketergantungan, dan Mitigasi

Refactoring arsitektural pasti membawa risiko. Tabel berikut merangkum risiko utama, dampaknya, dan mitigasi yang diusulkan.

Tabel 12 — Risiko → Dampak → Mitigasi

| Risiko                               | Dampak                              | Mitigasi                                                   |
|--------------------------------------|-------------------------------------|------------------------------------------------------------|
| Regressi fungsional                  | Fitur rusak; kepercayaan turun       | Suite uji otomatis; canary; rollback cepat                 |
| DI tidak stabil                      | Build gagal; runtime error           | Wire providers jelas; pipeline CI diperkuat                |
| Kebocoran data via log               | Risiko keamanan; compliance terganggu| Redaksi fields; kebijakan akses log                        |
| Route inkonsistensi                  | Bug antar domain                     | Pedoman naming; linting; review rute                       |
| Ketergantungan eksternal (AI/DB)     | Cascade failure; timeout             | Retry/backoff; circuit breaker; observabilitas             |
| Cold start Cloud Run                 | Latensi tinggi awal                  | Warm-up; readiness; optimasi image                         |
| Downtime deploy                      | Service unavailable                  | Blue/green; canary; rollback otomatis                      |

### Risiko Teknis

Perubahan DI dan konfigurasi berisiko runtime error jika tidak ditest menyeluruh. Route perubahan tanpa uji E2E berpotensi menyebabkan regressi di jalur kritis. Mitigasi utama adalah suite uji otomatis (unit/integration/E2E) untuk memvalidasi setiap perubahan.

### Risiko Operasional

Logging tidak boleh menyimpan data sensitif. Cold start dan drain connections perlu diuji dan dioptimalkan. Di Cloud Run, readiness/liveness, port binding, dan graceful shutdown adalah kunci.

### Risiko Organisasi

Koordinasi lintas tim dan perubahan proses harus dikelola. Dokumentasi dan sesi onboarding akan mempercepat pemahaman. DoD yang jelas per fase menjaga fokus perubahan.

## Metrik Keberhasilan & Penerimaan

Keberhasilan refactor diukur dengan kombinasi metrik kualitas, tes, performa, dan reliabilitas. Target bersifat indikatif dan diarahkan pada tren peningkatan.

Tabel 13 — KPI → Target (indikatif) → Cara Ukur

| KPI                 | Target (indikatif)                 | Cara Ukur                                   |
|---------------------|------------------------------------|---------------------------------------------|
| Kualitas arsitektur | Penurunan anti-pattern              | Review kode; linting                        |
| Cakupan tes         | Kenaikan signifikan                 | Coverage unit/integration/E2E               |
| Defect density      | Tren menurun                        | Insiden & QA laporan                        |
| Lead time           | Lebih singkat                       | Waktu commit → deploy stabil                |
| Deployment frequency| Meningkat tanpa gangguan             | Frekuensi rilis per periode                 |
| MTTR                | Lebih cepat                         | Durasi pemulihan insiden                    |
| Latency p95/p99     | Stabil/membaik                      | Metrik latency endpoint                     |
| Error rate          | Turun dan terkendali                | Persentase error per request                |
| Cold start time     | Menurun                             | Pengukuran inisialisasi Cloud Run           |

### Metrik Teknis

Peningkatan coverage tesunit/integration/E2E serta penurunan anti-pattern adalah indikator perbaikan. Performa endpoint stabil/meningkat setelah optimasi cold start dan refactor handler.

### Metrik Operasional

Jumlah insiden menurun dan MTTR membaik gracias observabilitas. Frekuensi rilis meningkat tanpa mengorbankan stabilitas.

## Lampiran

### Checklist Refactoring per Fase

Tabel 14 — Checklist per Fase → Status → Penanggung Jawab → Batas Waktu

| Fase | Item Checklist                                             | Status | Penanggung Jawab | Batas Waktu |
|------|------------------------------------------------------------|--------|------------------|-------------|
| 0    | Baseline audit & target arsitektur disetujui               |        | Tech Lead        | TBC         |
| 1    | Router & middleware pipeline aktif                         |        | Backend Lead     | TBC         |
| 2    | ≥80% handler refactor; DTO & error wrapper                 |        | Domain Owner     | TBC         |
| 3    | DI via wire & config/env berbasis environment             |        | DevOps/Backend   | TBC         |
| 4    | Logging terstruktur; pyramid tes                           |        | QA/Backend       | TBC         |
| 5    | Health/readiness & rollout canary                          |        | DevOps           | TBC         |

### Inventaris Endpoint dan Mapping ke Controller (ringkas)

- GET / → base.GetHome
- POST /auth/register → auth.Register (indikatif; perlu validasi)
- POST /auth/login → auth.Login (indikatif; perlu validasi)
- GET /auth/profile → auth.GetProfile (indikatif; perlu validasi)
- POST /api/project → project.CreateProject (indikatif; perlu validasi)
- GET /api/project → project.GetProjects (indikatif; perlu validasi)
- GET /api/project/:id → project.GetProjectByID (indikatif; perlu validasi)
- PUT /api/project → project.UpdateProject (indikatif; perlu validasi)
- DELETE /api/project → project.DeleteProject (indikatif; perlu validasi)
- POST /api/upload/:projectId → upload.UploadData (indikatif; perlu validasi)
- GET /api/preview/:uploadId → upload.GetDataPreview (indikatif; perlu validasi)
- GET /api/stats/:uploadId → upload.GetDataStats (indikatif; perlu validasi)
- POST /api/recommend/:projectId → analysis.GetRecommendations
- POST /api/process → analysis.ProcessAnalysis
- GET /api/results/:analysisId → analysis.GetAnalysisResults
- POST /api/refine/:analysisId → analysis.RefineAnalysis
- GET /api/export/:analysisId → export.ExportResults

### Daftar Artefak Baru yang Akan Dibuat/Diubah

- Router modular per domain dengan library gorilla/chi dan middleware pipeline.
- DTO request/response per domain dengan validator terpusat dan DTO resolver.
- Service/usecase dan repository interfaces; implementasi di infrastruktur.
- Wire DI providers untuk dependency graph; removal konstruktor eksplisit dari handler.
- Error wrapper terpusat; response standar; logging terstruktur dengan correlation ID.
- Health/readiness/liveness endpoints; metrik latency dan error rate.
- Facade untuk Vertex AI dan storage; strategi retry/backoff/circuit breaker.
- Strategi export formal (service export) dan format strategy (PDF/CSV/JSON).
- Dokumen pedoman naming & routing; pedoman validasi & logging.

## Kesimpulan

Backend Research Data Analysis memiliki organisasi dasar yang memadai, tetapi belum menunjukkan pemisahan concerns yang tegas sesuai GoCroot. Masalah utama—routing manual, controller tebal, ketiadaan service/usecase dan repository, serta DI—menjadi kunci penghambat skalabilitas, testabilitas, dan reliabilitas. Dengan roadmap refactor bertahap, perbaikan dapat dilakukan tanpa mengganggu stabilitas rilis, khususnya di Cloud Run. Prioritas awal adalah routing dan middleware, diikuti refactor controller ke thin handlers, penerapan service/usecase dan repository, DI via wire, standardisasi error dan logging, serta penguatan observability dan pengujian. Keberhasilan akan diukur melalui peningkatan kualitas arsitektur, cakupan tes, penurunan defect, dan metrik operasional yang lebih baik.

Informasi yang belum tersedia—seperti handler auth/project/upload yang belum terlihat, kebijakan autentikasi, desain error handling terpusat, pipeline CI/CD, dan kebutuhan audit keamanan—harus divalidasi sebelum eksekusi skala besar. Langkah validasi ini akan memperjelas ruang lingkup, mengurangi risiko, dan memastikan bahwa semua perubahan berjalan sesuai dengan prinsip GoCroot dan kebutuhan bisnis.