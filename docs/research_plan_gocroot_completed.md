# Blueprint Laporan: GoCroot Framework Resmi (Struktur, Routing, Middleware, Controller, Konfigurasi, Best Practices, Contoh, dan Integrasi)

## 1. Pendahuluan: Apa, Mengapa, dan Sumber Resmi GoCroot

GoCroot adalah boilerplate resmi berbasis Go yang memanfaatkan Go Fiber sebagai kerangka web dan MongoDB sebagai database. Proyek ini menyediakan struktur folder yang siap pakai untuk mempercepat pengembangan layanan HTTP, sekaligus mengadopsi praktik integrasi berkelanjutan (continuous integration/continuous delivery atau CI/CD) dan ekosistem pendukung seperti WhatsAuth untuk kebutuhan autentikasi berbasis pesan WhatsApp. Dalam praktik produksi, GoCroot berfungsi sebagai “starter” yang memisahkan konfigurasi, rute, controller, model, dan helper secara terorganisir, sehingga tim bisa fokus pada lógica bisnis tanpa harus membangun fondasi struktur dari nol[^1].

Sebagai kerangka kerja turunannya, GoCroot mengadopsi filosofi desain Go Fiber: ringan, terinspirasi Express, dan dioptimalkan di atas Fasthttp. Kombinasi ini memberikan jalur pembelajaran yang familier bagi developer yang berpengalaman dengan framework Express di JavaScript, sambil mempertahankan karakteristik performa Go yang efisien. Pada saat yang sama, GoCroot memperluas kemampuan dasar Fiber dengan templat project yang konsisten, variabel lingkungan, dan contoh integrasi ke MongoDB serta layanan WhatsAuth[^4].

Di ekosistem Go, terdapat juga modul “gocroot” dari maintainer lain yang secara eksplisit menyediakan antarmuka dan contoh integrasi Fiber–WhatsAuth serta pengaturan CORS dan koneksi database MariaDB/MongoDB. Kedua sumber ini tidak saling menggantikan; GoCroot gocroot/gocroot adalah boilerplate projekt yang menitikberatkan pada struktur dan deployment, sedangkan gocroot dari jasminemutiara03 berfungsi sebagai paket utilitas/SDK untuk mengaktifkan WebSocket QR, endpoint verifikasi, dan konfigurasi pendukung. Keduanya relevan, saling melengkapi, dan layak dijadikan rujukan implementasi resmi[^1][^3].

Untuk memudahkan orientasi, Tabel 1 merangkum sumber resmi yang menjadi landasan laporan ini. Tabel ini memetakan sumber ke peranan dan catatan relevansi, sehingga pembaca dapat memastikan rujukan “apa untuk apa” sejak awal.

Tabel 1. Peta sumber resmi GoCroot dan relevansi

| Sumber (ID) | Jenis | Peranan | Catatan relevansi |
|---|---|---|---|
| GoCroot boilerplate (1) | Repositori GitHub | Struktur proyek, organisasi folder, CI/CD | Template operasional yang siap pakai, baseline struktur folder |
| Situs dokumentasi GoCroot (2) | Situs proyek | Rujukan dokumentasi tambahan | Konteks dan referensi lintas tautan proyek |
| Paket gocroot (jasminemutiara03) (3) | Go package (pkg.go.dev) | API SDK, konfigurasi CORS, WhatsAuth | Contoh kode resmi untuk endpoint dan WebSocket |
| Dokumentasi Go Fiber (4) | Dok. resmi | Fondasi framework (routing, middleware, config) | Acuan utama desain dan kapabilitas teknis |
| Middleware Fiber (5) | Dok. resmi | Katalog middleware | Daftar middleware resmi dan deskripsinya |
| Error Handling Fiber (6) | Dok. resmi | Panduan error | Pola penanganan error yang idiomatik |
| Routing Fiber (7) | Dok. resmi | Pola routing, constraints, grouping | Praktik ordre deklarasi, prefix, middleware grup |
| Grouping Fiber (8) | Dok. resmi | Group/subrouting | Perilaku deklarasi virtual dan flattening |
| Request/Response Fiber (10) | Artikel teknis | Pola parsing/binding/response | Referensi praktis untuk implementasi handler |

Catatan penting: beberapa informasi tetap menjadi “information gaps” karena tidak tersedia di sumber resmi. Termasuk contoh resmi implementasi controller yang terpisah dari WhatsAuth, detail validasi input menyeluruh, contoh uji otomatis (unit/integration), dokumentasi dependency injection (DI) yang komprehensif, dan praktik observability (logging terstruktur, tracing, metrik) besides middleware Logger/Monitor. Bagian-bagian yang bergantung pada aspek tersebut akan disertai peringatan atauolahtf agar pembaca dapat melengkapi dengan praktik internal tim[^1][^3][^6].

### 1.1 Ringkasan Eksekutif

GoCroot adalah boilerplate Golang yang mengadopsi Go Fiber sebagai framework web, MongoDB sebagai penyimpanan data, dan WhatsAuth untuk autentikasi lewat WhatsApp. Repositori resmi menekankan struktur folder yang bersih, kompatibilitas CI/CD (Alwaysdata dan fly.io), dan integrasi yang mudah dengan layanan-layanan tersebut. Dengan demikian, GoCroot cocok digunakan sebagai starting point untuk layanan API modern di mana performa, struktur kode yang rapi, dan integrasi pihak ketiga (misalnya WhatsApp) menjadi kebutuhan utama[^1][^2].

## 2. Gambaran Umum Arsitektur GoCroot (What)

GoCroot memposisikan Go Fiber sebagai “mesin” aplikasi HTTP. Fiber adalah kerangka web yang terinspirasi Express dan dibangun di atas Fasthttp, mesin HTTP tercepat untuk Go. Kombinasi ini memberi trade-off yang perlu dipahami: Fiber menggunakan strategi zero-allocation pada konteks permintaan untuk throughput tinggi, sementara beberapa kebutuhan seperti immutable context mengharuskan konfigurasi eksplisit yang berdampak pada performa. Di atas fondasi ini, GoCroot menyusun struktur folder yang memisahkan perhatian (separation of concerns) secara jelas[^1][^4].

Untuk memperjelas peranan masing-masing direktori, Tabel 2 memetakan folder inti GoCroot terhadap tanggung jawab dan jenis konten yang biasanya menempatkannya. Tabel ini dapat dianggap sebagai “pemandangan birds-eye” untuk orientasi struktur proyek.

Tabel 2. Folder GoCroot dan tanggung jawab

| Folder | Tanggung jawab utama | Jenis konten |
|---|---|---|
| .github | Konfigurasi CI/CD | Workflow GitHub Actions |
| config | Semua konfigurasi aplikasi | Variabel lingkungan, CORS, koneksi DB |
| controller | Fungsi handler endpoint | HTTP handlers (termasuk WebSocket) |
| helper | Fungsi pendukung | Utility yang dipanggil lintas paket |
| model | Struktur data | Tipe/domain model |
| url | Rute dan grouping | Pendekatan “URL routing” map ke controller |

Struktur di atas memfasilitasi pemisahan yang jelas: konfigurasi tidak tercampur dengan logika rute, controller tidak dibebani detail infrastruktur, dan model berdiri sebagai representasi domain yang bersih. Dengan pemisahan seperti ini, refactoring, pengujian, dan escalasi ke produksi menjadi lebih terukur[^1].

### 2.1 Diagram Arsitektur/logical view

Secara logika, alur permintaan berjalan sebagai berikut: klien → router Fiber → middleware → controller → layanan/domain → repositori → database. Komponen “config” membaca variabel lingkungan (misalnya MONGOSTRING, kunci publik/privat untuk WhatsAuth) dan menyediakan objek koneksi yang dibutuhkan di lapisan repositori atau layanan. Pendekatan ini menjaga controller tetap ringkas: mereka hanya mengubah input HTTP menjadi panggilan ke layanan/domain, serta memformat respons HTTP yang konsisten[^1].

## 3. Struktur Framework dan Organisasi Kode (How — Structure)

Struktur paket GoCroot secara eksplisit memisahkan empat lapisan inti: konfigurasi (config), model (model), rute (url), dan controller (controller). “helper” menyediakan fungsi-fungsi通用 (utility) yang dapat digunakan lintas paket tanpa menimbulkan ketergantungan yang berantai. Dalam praktik, pola ini mengikuti kaidah idiomatik Go: gunakan paket kecil yang jelas tujuan dan batasannya, serta minimkan ketergantungan silang. Hal ini sejalan dengan panduan umum cara menulis kode Go yang baik, seperti yang dirangkum dalam dokumentasi resmi Go dan Effective Go[^13][^14].

Untuk memperjelas kepemilikan file dan relasi antarpaket, Tabel 3 merangkum beberapa file kunci beserta paket dan deskripsi singkatnya. Tabel ini berguna untuk reviewer kode agar memahami “apa yang sebaiknya hidup di mana”.

Tabel 3. File kunci, paket, dan deskripsi

| File | Paket | Deskripsi ringkas |
|---|---|---|
| main.go | main | Titik masuk aplikasi, inisialisasi Fiber, CORS, rute, dan server |
| url/url.go | url | Mapp route ke handler controller (termasuk WebSocket) |
| controller/controller.go | controller | Handler endpoint dan WebSocket WhatsAuth |
| config/cors.go | config | Konfigurasi CORS (allow origins/headers/credentials) |
| config/token.go | config | Pembacaan kunci publik/privat dari lingkungan |
| config/db.go | config | Informasi koneksi MariaDB/MongoDB |

### 3.1 Pola Penamaan dan Batas Paket

- Paket config: mengelola variabel lingkungan (misalnya MONGOSTRING, PUBLICKEY, PRIVATEKEY) dan objek koneksi (MariaConnection) agar tersentralisasi. Tidak menulis output HTTP, tidak melakukan validasi lintas-layer secara prosedur.
- Paket controller: hanya berisi handler HTTP (termasuk WebSocket). Tidak mengakses detail infrastruktur di luar yang disediakan config; mendelegasikan ke layanan/domain.
- Paket model: menggambarkan tipe/domain. Bebas dari detail transport HTTP dan infrastruktur database.
- Paket helper: berisi utility yang “dipinjam” controller/layanan. Tidak membawa state implisit, mudah di-test, dan tidak menciptakan ketergantungan melingkar.

Dengan batas paket yang jelas, reviewer bisa menilai kelayakan perubahan tanpa担心 “efek samping” ke lapisan lain. Pola ini juga selaras dengan praktik umum “How to Write Go Code” yang mendorong organisasi paket/module yang eksplisit[^14].

## 4. Routing System (How — Routing di Fiber/GoCroot)

Fiber menyediakan API routing yang ekspresif: metode HTTP (Get, Post, Put, Delete, dll.), pola path (termasuk named, greedy, optional), constraints untuk type/value, serta grouping dan middleware per grup. Secara performa, Fiber mencatat rute dalam urutan deklarasi; oleh karena itu, menempatkan rute statis sebelum rute dinamis adalah praktik penting untuk mencegah shadowing yang tidak diharapkan[^7][^8].

Dalam GoCroot, pendefinisian route terpusat di paket url (misalnya Web(page *fiber.App)), yang kemudian memetakan endpoint ke fungsi controller terkait. Pola ini menjaga main.go tetap minimal, sementara url berfungsi sebagai “peta navigasi” aplikasi[^1][^3].

Untuk membantu pemahaman, Tabel 4 merangkum Jenis Parameter Rute di Fiber beserta contoh path dan kecocokannya. Tabel ini sekaligus menegaskan perbedaan antara greedy dan optional agar route tester dan reviewer selaras dalam ekspektasi.

Tabel 4. Jenis parameter rute Fiber

| Jenis | Introducer | Contoh path | Nilai c.Params(...) | Catatan |
|---|---|---|---|---|
| Named | : | /users/:id | id | Segment dinamis dengan nama |
| Greedy + | + | /files/+ | + | Menangkap segmen sisa, tidak optional |
| Greedy * | * | /api/* | * | Menangkap segmen sisa, optional |
| Optional ? | :param? | /search?q? | q | Membuat named param opsional |
| Wildcard opsional | * | /user/* | * | Cocok dengan apa pun di bawah /user |
| Literal dot/hyphen | . / - | /plantae/:genus.:species | genus, species | Titik dan hyphen interpreted literally |

Selain parameter, constraints di Fiber (v2.37.0+) memungkinkan pembatas pada nilai param tanpa validasi manual di handler. Tabel 5 merangkum constraints yang tersedia beserta contoh kecocokan. Menggunakan constraints membantu mengurangi boilerplate validasi dan menyederhanakan controller.

Tabel 5. Constraints Fiber dan contoh kecocokan

| Constraint | Contoh param | Kecocokan | Kegunaan |
|---|---|---|---|
| int | :id<int> | 123, -10 | Numerik bulat |
| bool | :active<bool> | true, false | Nilai boolean |
| guid | :id<guid> | CD2C1638-... | Format GUID |
| float | :weight<float> | 1.234, -1.001e8 | Bilangan desimal |
| minLen(n) | :username<minLen(4)> | Test | Panjang minimum |
| maxLen(n) | :filename<maxLen(8)> | MyFile | Panjang maksimum |
| len(n) | :code<len(12)> | somefile.txt | Panjang tetap |
| min(v) | :age<min(18)> | 19 | Batas bawah |
| max(v) | :age<max(120)> | 91 | Batas atas |
| range(min,max) | :age<range(18,120)> | 91 | Rentang nilai |
| alpha | :name<alpha> | Rick | Alfabet case-insensitive |
| datetime | :dob<datetime> | 2005-11-01 | Format tanggal |
| regex(expr) | :date<regex(\d{4}-\d{2}-\d{2})> | 2022-08-27 | Pola regex |

Untuk mengilustrasikan struktur routing di GoCroot, Tabel 6 memetakan endpoint contoh dari url ke handler controller. Contoh ini juga menunjukkan pemisahan concerns: url memetakan, controller menangani.

Tabel 6. Pemetaan endpoint contoh dari url ke controller

| Endpoint | Method | Handler (paket/controller) | Deskripsi |
|---|---|---|---|
| /api/whatsauth/request | POST | controller/PostWhatsAuthRequest | Memproses permintaan autentikasi WhatsApp |
| /ws/whatsauth/qr | GET (WebSocket) | controller/WsWhatsAuthQR | Menyajikan QR ke klien melalui WebSocket |

### 4.1 Konvensi Route dan Urutan Deklarasi

Urutan deklarasi matters. Selalu tempatkan rute statis (misalnya /users/profile) sebelum rute dinamis (misalnya /users/:id). Dengan demikian, router tidak “menelan” rute statis melalui pencocokan parameter yang terlalu greedy. Fiber mengeksekusi rute secara flatten berdasarkan prefix Group dan urutan pendaftaran; Group tidak mengubah urutan eksekusi, hanya menambahkan prefix dan middleware yang diterapkan ke seluruh rute di dalamnya[^7][^8].

## 5. Request/Response Handling (How — Binds dan Respons)

Request handling di Fiber memanfaatkan c.Params untuk parameter rute, c.Query untuk query string, c.FormValue untuk data formulir, dan c.BodyParser untuk payload JSON. Pada sisi response, Fiber menyediakan c.JSON untuk respons terstruktur, c.Status/c.SendStatus untuk mengatur kode status, serta c.SendString/c.SendFile untuk konten teks dan file. Fiber menetapkan status 200 untuk respons sukses secara default; namun, kebiasaan baik adalah selalu explicit mengatur status ketika ada kondisi error atau payload tidak standar[^7][^10].

Untuk memudahkan reviewer memastikan konsistensi, Tabel 7 memetakan sumber data HTTP ke metode akses yang disarankan dan contoh penggunaannya.

Tabel 7. Mapping sumber data ke metode akses

| Sumber data | Metode akses | Contoh | Catatan |
|---|---|---|---|
| Parameter rute | c.Params | c.Params("id") | Ambil segmen dinamis |
| Query | c.Query | c.Query("page") | Default values bisa digunakan |
| Form | c.FormValue | c.FormValue("email") | Untuk form-urlencoded |
| JSON body | c.BodyParser | c.BodyParser(&payload) | Binding ke struct/pointer |

Jenis respons dan status code yang disarankan diringkas pada Tabel 8. Tabel ini membantu menormalisasi kontrak API lintas tim.

Tabel 8. Jenis respons dan status code yang disarankan

| Tipe respons | Metode | Status default | Kapan digunakan |
|---|---|---|---|
| JSON sukses | c.JSON | 200 | Respons data terstruktur |
| Teks | c.SendString | 200 | Pesan ringkas |
| File | c.SendFile | 200 | Serving file statis |
| Error | c.Status/SendStatus | 4xx/5xx | Error validation/identifikasi |

### 5.1 Zero Allocation dan Immutable Context

Secara default, nilai yang dikembalikan oleh fiber.Ctx dirancang untuk digunakan kembali (reuse) di seluruh permintaan sebagai bagian dari optimasi performa. Oleh karena itu, menahan referensi ke nilai tersebut di luar handler berpotensi memicu race atau reuse yang tidak diinginkan. Jika Anda perlu menyimpan nilai lebih lama, gunakan copy pada buffer (misalnya utils.CopyString) atau aktifkan konfigurasi Immutable: true, dengan memahami bahwa konsekuensinya adalah penurunan throughput. Prinsip ini menuntun pada pola aman: gunakan konteks hanya dalam scope handler; untuk menyimpan nilai, salin terlebih dahulu[^4].

## 6. Middleware (How — Komposisi dan Implementasi)

Middleware Fiber adalah fungsi yang dirantai dalam siklus permintaan HTTP dengan akses ke Context untuk melakukan aksi tertentu (logging, CORS, proteksi CSRF, rate limiting, dsb.). Setiap middleware yang dibungkus dengan app.Use(...) akan dieksekusi sebelum handler rute yang relevan. Penggunaan c.Next() memindahkan eksekusi ke tahap berikutnya. Komposisi yang baik biasanya menempatkan middleware lintas (global) di bagian awal, kemudian middleware per grup di prefix Group terkait[^5][^7][^8].

Untuk memudahkan pemilihan, Tabel 9 merangkum middleware Fiber yang relevan untuk produksi dan fungsi utamanya.

Tabel 9. Middleware Fiber yang relevan

| Nama | Fungsi utama | Use-case tipikal |
|---|---|---|
| CORS | Mengontrol akses lintas origin | Mengizinkan origin tepercaya |
| Helmet | Mengatur header keamanan | Melindungi dari serangan umum |
| CSRF | Proteksi token untuk method non-aman | Mencegah forgery permintaan |
| Recover | Menangkap panic | Stabilitas layanan |
| Limiter | Rate limiting | Protect endpoint sensitif |
| Logger | Log permintaan | Observability dasar |
| RequestID | Menyisipkan ID permintaan | Korelasi log |
| Compress | Kompresi respons | Efisiensi bandwidth |
| ETag | Cache-control via ETag | Menghemat transfer data |
| Monitor | Metrik server | Health metrik |
| Pprof | Profiling runtime | Analisis performa |

### 6.1 Pola Error Handler Terpusat

Error handler terpusat dikonfigurasi melalui fiber.Config pada saat inisialisasi aplikasi. Di dalamnya, Anda dapat mengubah semua error menjadi respons JSON terstruktur, melakukan logging terpusat, atau mengirim notifikasi ke sistem eksternal (misalnya email/alerting). Pastikan semua panic dilindungi oleh middleware Recover agar ErrorHandler selalu dapat mengambil alih dan mencegah layanan crash. Selain itu, gunakan fiber.NewError(...) untuk kesalahan yang memerlukan kode status spesifik—misalnya 404 atau 503—agar konsistensi kontrak API terjaga[^6].

## 7. Controller Patterns (How — Pola Implementasi di GoCroot)

Pola yang dianjurkan untuk controller GoCroot adalah “HTTP handler sempit” yang hanya mengubah permintaan HTTP menjadi panggilan ke layanan/domain, serta memformat respons HTTP. Pada boilerplate GoCroot, dua controller inti adalah WsWhatsAuthQR untuk WebSocket QR dan PostWhatsAuthRequest untuk endpoint POST. Keduanya memanfaatkan konfigurasi dari paket config (misalnya kunci publik/privat dan informasi koneksi database) tanpa membuat ketergantungan yang berantai ke infrastruktur di dalam controller[^1][^3].

Tabel 10 memetakan controller ke route untuk memperjelas kontrak endpoint.

Tabel 10. Pemetaan controller GoCroot

| Nama | Endpoint | Method | Deskripsi |
|---|---|---|---|
| WsWhatsAuthQR | /ws/whatsauth/qr | GET (WebSocket) | Menjalankan WebSocket QR menggunakan kunci publik |
| PostWhatsAuthRequest | /api/whatsauth/request | POST | Menjalankan modul WhatsAuth untuk verifikasi |

### 7.1 Struktur Handler yang Sehat

Praktik sehat untuk handler meliputi:
- Parsing dan validasi input melalui binding (misalnya c.BodyParser) sebelum memproses logika bisnis.
- Penggunaan errors berstatus melalui fiber.NewError(...), sehingga error handler terpusat dapat menyatukan bentuk respons.
- Pembatasan pekerjaan berat dalam handler: jika proses computationally expensive atau blocking I/O, pertimbangkan offloading ke goroutine dengan konteks yang tepat atau ke layanan terpisah; pastikan timeouts dan cancellation di-handle dengan benar.

Prinsip ini mengikuti panduan penanganan error idiomatik Fiber dan membantu mencegah kebocoran tanggung jawab dari lapisan domain ke transport[^6].

## 8. Konfigurasi dan Environment Setup (How — Config)

Konfigurasi GoCroot memanfaatkan variabel lingkungan untuk menyimpan rahasia dan parameter koneksi. Varian yang lazim meliputi MONGOSTRING, PUBLICKEY, PRIVATEKEY, dan daftar origin CORS. Semua ini dikelola di paket config, yang kemudian diekspor sebagai variabel bernilai ke lapisan lain (controller, url). Pada produksi, penggunaan .env untuk pengembangan dan secrets pada platform CI/CD (misalnya GitHub Actions) adalah pendekatan yang lazim, dengan pembedaan jelas antara “development” dan “production”[^1][^3][^16].

Sebelum menyajikan checklist, perlu dicatat: dokumentasi GoCroot dan paket jasminemutiara03 menyajikan variabel lingkungan yang relevan. Namun, daftar lengkap danstandarisasi “nama variabel resmi” lintas boilerplate mungkin belum seragam. Tim disarankan mendefinisikan “kontrak lingkungan” internal sebagai bagian dari dokumen operasional.

Tabel 11 merangkum variabel lingkungan kunci yang terdokumentasi dalam sumber resmi.

Tabel 11. Variabel lingkungan kunci

| Nama | Deskripsi | Sumber | Catatan |
|---|---|---|---|
| MONGOSTRING | URI koneksi MongoDB | (1) | Diinisialisasi di config/db.go |
| PUBLICKEY | Kunci publik WhatsAuth | (1)(3) | Digunakan WebSocket/verifikasi |
| PRIVATEKEY | Kunci privat WhatsAuth | (1)(3) | Digunakan modul verifikasi |
| INTERNALHOST | Host internal untuk permintaan | (3) | Digunakan untuk判断 host |
| PORT | Port layanan | (3) | Digunakan saat listen server |
| AllowOrigins | Daftar origin CORS | (3) | Diolah di config CORS |

Untuk operasional, Tabel 12 menyajikan matriks lingkungan deployments berdasarkan penyedia. Tabel ini membantu tim DevOps dan engineering memastikan bahwa pipeline dari development ke production tidak kehilangan langkah penting.

Tabel 12. Lingkungan deployment (Development vs Production)

| Lingkungan | Cara pengaturan | Catatan operasional |
|---|---|---|
| Development | .env + config lokal | Mudah diubah; cocok untuk iterasi cepat |
| Production (Alwaysdata) | Environment di panel, secrets di GitHub Actions | Atur Web > Sites > Modify, SSH, APIKEY; Gunakan scheduled tasks untuk refresh token[^1] |
| Production (fly.io) | fly.toml + GitHub Actions | Pilih workflow template, hapus ekstensi .template[^1] |

### 8.1 Konfigurasi CORS

Konfigurasi CORS di GoCroot disusun di config/cors.go, mencantumkan daftar origin yang diizinkan, headers yang diekspos, dan pengaturan kredensial. Prinsipnya: batasi origin ke domain tepercaya, hindari wildcard luas di produksi, dan selalu uji skenario preflight. CORS middleware di Fiber menyediakan konfigurasi fleksibel untuk kebutuhan ini[^3][^5].

## 9. Integrasi Database dan Services Lainnya (How — Integrations)

Integrasi MongoDB di boilerplate GoCroot mengikuti prinsip koneksi terpusat di config dan penggunaan driver resmi MongoDB Go. Pola umum yang digunakan adalah pembacaan MONGO_URI dari lingkungan, penyimpanan objek klien/database dalam variabel config, dan penggunaan objek tersebut di lapisan repositori atau controller. Praktik CRUD (Create, Read, Update, Delete) mengikuti manual MongoDB, termasuk pagination dan pencarian dengan filter regex di endpoint “GetAll”. Pembacaan ini memastikan operasi database idiomatik dan dapat dipelihara tanpa tercampur dengan logika transport[^1][^11][^12][^15].

Tabel 13 merangkum operasi CRUD umum pada MongoDB dengan parameter paginasi/pencarian.

Tabel 13. CRUD MongoDB dengan contoh parameter

| Operasi | Method | Endpoint | Parameter | Respons umum |
|---|---|---|---|---|
| Create | POST | /items | JSON payload | 201 Created |
| Read all | GET | /items | page, limit, s (regex) | 200 OK + array |
| Read one | GET | /items/:id | id (ObjectId) | 200 OK / 404 Not Found |
| Update | PUT | /items/:id | id + payload | 200 OK / 404 Not Found |
| Delete | DELETE | /items/:id | id | 200 OK / 404 Not Found |

Integrasi WhatsAuth memanfaatkan dua endpoint inti: /api/whatsauth/request (POST) untuk verifikasi permintaan dan /ws/whatsauth/qr (WebSocket) untuk menyajikan QR ke klien. Konfigurasi kunci publik/privat dan informasi tabel pengguna (misalnya mahasiswa, dosen) dikelola di config, sehingga handler tetap fokus pada transport dan delegasi ke modul WhatsAuth. Praktik produksi juga mensyaratkan pembaruan token secara terjadwal; Alwaysdata mendukung scheduled tasks yang dapat memanggil endpoint refresh aplikasi secara berkala[^1][^3].

Tabel 14 menyajikan ringkasan endpoint integrasi WhatsAuth.

Tabel 14. Endpoint integrasi WhatsAuth

| Path | Method | Handler | Fungsi | Catatan |
|---|---|---|---|---|
| /api/whatsauth/request | POST | controller.PostWhatsAuthRequest | Memproses verifikasi | Gunakan kunci privat |
| /ws/whatsauth/qr | GET (WS) | controller.WsWhatsAuthQR | Menampilkan QR | Gunakan kunci publik |

### 9.1 Pola Repository dan Layering

Pemisahan controller–service–repository membantu menjaga kode tetap testable dan decoupling dari transport HTTP. Controller tidak perlu tahu detail koneksi MongoDB; ia cukup memanggil service/domain yang pada gilirannya menggunakan repositori untuk operasi database. Prinsip ini penting untuk memungkinkan penggantian driver atau perpindahan infrastruktur tanpa mengacak-acak lapisan handler[^1].

## 10. Best Practices dan Konvensi (So What — Praktik yang Disarankan)

GoCroot menganjurkan konvensi yang mendukung readability, maintainability, dan performa:

- Organisasi paket: pemisahan config/controller/url/helper/model agar responsibilities jelas dan reviewer mudah menavigasi struktur.
- Routing: gunakan Group untuk prefix dan middleware bersama; jaga urutan deklarasi (statis sebelum dinamis) serta perhatikan perilaku flattening.
- Error handling: terapkan ErrorHandler terpusat, gunakan Recover untuk proteksi panic, dan konsistenkan error responses melalui fiber.NewError.
- Zero allocation: hindari menyimpan referensi fiber.Ctx di luar handler; untuk data yang perlu dipertahankan, gunakan copy atau aktifkan Immutable dengan memahami trade-off performa.
- Observability: gunakan Logger untuk log permintaan dan Monitor untuk metrik dasar; catat bahwa praktik advanced logging/ tracing/ metrics belum terdokumentasi penuh dalam sumber dan perlu dilengkapi sendiri.
- Validasi: manfaatkan constraints di rute untuk mengurangi boilerplate validasi; untuk validasi kompleks, gunakan layer service/domain.
- Security: aktifkan CORS dan Helmet sesuai kebutuhan; gunakan CSRF untuk endpoint sensitive; rate-limit endpoint publik seperti reset password[^4][^5][^6][^7].

Tabel 15 merangkum checklist ringkas best practices untuk reviewer dan QA.

Tabel 15. Checklist best practices

| Area | Praktik utama | Referensi |
|---|---|---|
| Struktur paket | Separation of concerns, batas paket jelas | (13), (14) |
| Routing | Group prefix, urutan statis–dinamis | (7), (8) |
| Error | ErrorHandler terpusat, Recover | (6) |
| Performa | Zero-allocation, copy vs Immutable | (4) |
| Observability | Logger, Monitor | (5) |
| Security | CORS, Helmet, CSRF, Limiter | (5) |
| Validasi | Route constraints, service-level | (7) |

## 11. Contoh Implementasi (How — Kode dan Skema Endpoints)

Untuk memperjelas implementasi, berikut beberapa contoh ringkas. Pertama, main.go menginisialisasi aplikasi Fiber, mengaktifkan CORS, dan mendaftarkan rute via url.Web:

```go
package main

import (
    "log"
    "github.com/aiteung/musik"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "iteung/config"
    "github.com/whatsauth/whatsauth"
    "github.com/gofiber/fiber/v2"
    "iteung/url"
)

func main() {
    go whatsauth.RunHub() // menjalankan hub WebSocket WhatsAuth
    site := fiber.New()
    site.Use(cors.New(config.Cors))
    url.Web(site)
    log.Fatal(site.Listen(musik.Dangdut()))
}
```

Kedua, pendefinisian route di url/url.go memetakan endpoint ke controller:

```go
package url

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "iteung/controller"
)

func Web(page *fiber.App) {
    page.Post("/api/whatsauth/request", controller.PostWhatsAuthRequest)
    page.Get("/ws/whatsauth/qr", websocket.New(controller.WsWhatsAuthQR))
}
```

Ketiga, controller/controller.go mengimplementasikan handler, termasuk binding JSON dan pemanggilan modul WhatsAuth:

```go
package controller

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "github.com/whatsauth/whatsauth"
    "iteung/config"
)

func WsWhatsAuthQR(c *websocket.Conn) {
    whatsauth.RunSocket(c, config.PublicKey, config.Usertables[:], config.Ulbimariaconn)
}

func PostWhatsAuthRequest(c *fiber.Ctx) error {
    if string(c.Request().Host()) == config.Internalhost {
        var req whatsauth.WhatsauthRequest
        err := c.BodyParser(&req)
        if err != nil {
            return err
        }
        ntfbtn := whatsauth.RunModule(req, config.PrivateKey, config.Usertables[:], config.Ulbimariaconn)
        return c.JSON(ntfbtn)
    } else {
        var ws whatsauth.WhatsauthStatus
        ws.Status = string(c.Request().Host())
        return c.JSON(ws)
    }
}
```

Keempat, konfigurasi CORS di config/cors.go:

```go
package config

import (
    "os"
    "strings"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

var origins = []string{
    "https://auth.ulbi.ac.id",
    "https://sip.ulbi.ac.id",
    "https://euis.ulbi.ac.id",
    "https://home.ulbi.ac.id",
    "https://alpha.ulbi.ac.id",
    "https://dias.ulbi.ac.id",
    "https://iteung.ulbi.ac.id",
    "https://whatsauth.github.io",
}

var Internalhost string = os.Getenv("INTERNALHOST") + ":" + os.Getenv("PORT")
var Cors = cors.Config{
    AllowOrigins: strings.Join(origins[:], ","),
    AllowHeaders: "Origin",
    ExposeHeaders: "Content-Length",
    AllowCredentials: true,
}
```

Untuk keperluan operasional, Tabel 16 merangkum skema endpoint inti.

Tabel 16. Skema endpoint contoh

| Path | Method | Handler | Deskripsi |
|---|---|---|---|
| /api/whatsauth/request | POST | controller.PostWhatsAuthRequest | Verifikasi permintaan |
| /ws/whatsauth/qr | GET (WebSocket) | controller.WsWhatsAuthQR | Streaming QR ke klien |

### 11.1 Catatan Implementasi

- Tangani error binding JSON secara eksplisit dan kembalikan fiber.NewError(400, ...) bila parsing gagal untuk memberi sinyal jelas ke klien dan error handler.
- Validasi input sebelum delegasi ke service/domain untuk mengurangi kebocoran validasi lintas-layer.
- Gunakan middleware Logger untuk pelacakan request dan Monitor untuk metrik dasar, terutama pada endpoint publik dan critical path[^6][^5].

## 12. Uji Coba, Deployment, dan Operasional

GoCroot mendukung deployment melalui GitHub Actions ke Alwaysdata dan fly.io. Untuk Alwaysdata, langkah-langkah meliputi konfigurasi environment di panel (Web > Sites > Modify), SSH, secrets (apikey, sshusername, sshpassword, dll.), serta scheduled tasks untuk refresh token WhatsAuth setiap tiga minggu. Untuk fly.io, gunakan fly.toml dan workflow yang disediakan; setelah pipeline siap, hapus ekstensi .template sesuai instruksi[^1][^16].

Tabel 17 menyajikan checklist deployment.

Tabel 17. Checklist deployment

| Item | Aksi | Catatan |
|---|---|---|
| Environment | Set MONGOSTRING, PUBLICKEY, PRIVATEKEY, WEBHOOKURL, WEBHOOKSECRET | Terpusat di config; gunakan panel atau secrets |
| SSH | Konfigurasi akses | Pastikan login berbasis password tersedia |
| APIKEY | Generate token | Tambahkan ke GitHub Actions secrets |
| Scheduled tasks | Tambah task akses URL | Set endpoint refresh per 3 minggu |
| Workflow CI/CD | Pilih provider (Alwaysdata/fly.io) | Hapus ekstensi .template |

Dari sisi observability, gunakan middleware Logger untuk log permintaan dan Monitor untuk metrik runtime. Jika diperlukan profiling, aktifkan Pprof di lingkungan staging/non-produksi. Meskipun praktik observability lanjutan belum terdokumentasi dalam sumber GoCroot, kombinasi middleware yang ada sudah memadai untuk monitoring dasar dan korelasi permintaan[^5].

## 13. Kesimpulan dan Rekomendasi Strategis

GoCroot menghadirkan struktur yang jelas, routing yang ekspresif, middleware lengkap, controller yang sempit, serta integrasi MongoDB dan WhatsAuth yang siap pakai. Di atas Fiber, GoCroot menyederhanakan pembangunan layanan produksi tanpa mengorbankan kinerja atau kebersihan arsitektur. Di sisi performa, pemahaman zero-allocation dan trade-off Immutable menjadi kunci; sementara di sisi keamanan, kombinasi CORS, Helmet, CSRF, dan Limiter menawarkan lapisan perlindungan yang solid untuk berbagai skenario[^1][^4][^6].

Rekomendasi implementasi:
- Mulai dari boilerplate GoCroot, kemudian perluas dengan service/repository pattern jika kebutuhan domain meningkat.
- Terapkan ErrorHandler terpusat dan middleware Recover secara global.
- Gunakan constraints di rute untuk mengurangi validasi boilerplate; validasi kompleks dilayani di lapisan service/domain.
- Standarisasi pengelolaan lingkungan melalui .env (dev) dan secrets (production), dan pastikan scheduled tasks dipadwalkan untuk rotasi token.
- Lengkapi praktik observability (structured logging, tracing) dan uji otomatis (unit/integration) untuk produksi, karena aspek-aspek tersebut belum terdokumentasi lengkap di sumber resmi.

Dengan mengikuti blueprint ini, tim akan memiliki landasan yang kuat untuk menilai kelayakan perubahan, menjaga konsistensi kontrak API, dan mencegah debt teknis yang tidak perlu.

## Information Gaps

- Contoh resmi controller non-WhatsAuth dalam GoCroot tidak tersedia dalam sumber; contoh yang ada berfokus pada WhatsAuth.
- Validasi input ketat secara menyeluruh (misalnya via middleware khusus atau schema validation) tidak terdokumentasi di boilerplate.
- Panduan uji otomatis (unit/integration) tidak disediakan di repositori resmi.
- Dokumentasi dependency injection (DI) resmi untuk GoCroot belum lengkap; integrasi pihak ketiga seperti Parsley bisa dipertimbangkan namun tidak menjadi bagian dari boilerplate.
- Praktik observability lanjutan (logging terstruktur, tracing, metrics) belum didokumentasikan beyond middleware Logger dan Monitor.
- Daftar lengkap variabel lingkungan “resmi” belum distandardisasi lintas repositori (gocroot/gocroot vs jasminemutiara03/gocroot).

---

## References

[^1]: gocroot/gocroot: boilerplate untuk GoCroot - GitHub. https://github.com/gocroot/gocroot
[^2]: Situs/Dokumentasi GoCroot. https://gocroot.github.io/gocroot/
[^3]: Paket GoCroot (jasminemutiara03) - Go Packages. https://pkg.go.dev/github.com/jasminemutiara03/gocroot
[^4]: Dokumentasi Fiber v2. https://docs.gofiber.io/
[^5]: Katalog Middleware Fiber. https://docs.gofiber.io/category/-middleware/
[^6]: Panduan Error Handling Fiber. https://docs.gofiber.io/guide/error-handling/
[^7]: Panduan Routing Fiber. https://docs.gofiber.io/guide/routing/
[^8]: Panduan Grouping Fiber. https://docs.gofiber.io/guide/grouping/
[^10]: Request and Response Handling in Fiber - WithCodeExample. https://withcodeexample.com/request-response-handling-fiber-powerful-web-apps/
[^11]: CRUD MongoDB - Manual. https://www.mongodb.com/docs/manual/crud/
[^12]: Tutorial: Go REST API dengan Fiber dan MongoDB. https://withmike.co.za/articles/build-a-go-rest-api-with-fiber-and-mongodb/
[^13]: Dokumentasi Resmi Go. https://go.dev/doc/
[^14]: Effective Go. https://go.dev/doc/effective_go
[^15]: fiber-mongo-example - GitHub. https://github.com/bmdavis419/fiber-mongo-example
[^16]: Alwaysdata (hosting/SSH/CI-CD). https://www.alwaysdata.com/en/
[^17]: WhatsAuth Docs (Integrasi layanan WhatsApp). https://whatsauth.my.id/docs/
[^18]: WhatsAuth Signup (API docs link). https://wa.my.id/
[^19]: WhatsAuth Webhook Reply JSON. https://whatsauth.my.id/webhook/iteung.reply.json