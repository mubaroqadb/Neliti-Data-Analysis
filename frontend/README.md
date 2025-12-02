# Research Data Analysis - Frontend

Frontend aplikasi Research Data Analysis, dibangun dengan JSCroot framework dan di-deploy ke GitHub Pages.

## Teknologi yang Digunakan

- **Framework**: JSCroot (Vanilla JS dengan ES6 Modules)
- **Styling**: Custom CSS dengan CSS Variables
- **Charts**: Chart.js
- **Deployment**: GitHub Pages

## Struktur Direktori

```
frontend/
├── .github/
│   └── workflows/
│       └── deploy-pages.yml   # GitHub Actions untuk deployment
├── css/
│   └── style.css              # Stylesheet utama
├── js/
│   └── main.js                # JavaScript utama (JSCroot)
├── index.html                 # Halaman utama (SPA)
└── README.md
```

## Fitur

### Halaman Publik
- **Beranda**: Landing page dengan informasi aplikasi
- **Login**: Form login pengguna
- **Registrasi**: Form pendaftaran pengguna baru

### Halaman Terproteksi (Memerlukan Login)
- **Dashboard**: Ringkasan proyek dan statistik
- **Proyek Baru**: Form pembuatan proyek penelitian
- **Detail Proyek**: 
  - Tab Ringkasan: Informasi proyek
  - Tab Upload: Upload file data (CSV, Excel, JSON)
  - Tab Analisis: Dapatkan rekomendasi dan proses analisis
  - Tab Hasil: Lihat hasil analisis dan ekspor

## JSCroot Modules yang Digunakan

```javascript
// Element manipulation
import { setInner, addInner, getValue, setValue, onClick, hide, show } 
    from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/element.js';

// API calls
import { postJSON, getJSON, postFileWithHeader } 
    from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/api.js';

// Cookie management
import { setCookieWithExpireHour, getCookie, deleteCookie } 
    from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/cookie.js';

// URL routing
import { redirect, getHash, setHash, onHashChange } 
    from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/url.js';
```

## Konfigurasi

Edit `js/main.js` untuk mengubah URL API backend:

```javascript
const API_BASE_URL = 'https://asia-southeast1-YOUR_PROJECT_ID.cloudfunctions.net/ResearchDataAnalysis';
```

## Pengembangan Lokal

Untuk development lokal, gunakan HTTP server sederhana:

```bash
# Menggunakan Python
python -m http.server 8080

# Atau menggunakan Node.js
npx serve .

# Akses di browser
open http://localhost:8080
```

## Deployment

Deployment otomatis dilakukan melalui GitHub Actions saat push ke branch `main`.

### Langkah Manual

1. Fork atau clone repository
2. Edit `js/main.js` untuk mengatur `API_BASE_URL`
3. Push ke branch `main`
4. Enable GitHub Pages di Settings > Pages
5. Pilih source: "GitHub Actions"

## Responsif

Aplikasi mendukung tampilan responsif untuk:
- Desktop (> 768px)
- Tablet dan Mobile (< 768px)

## Browser Support

- Chrome (terbaru)
- Firefox (terbaru)
- Safari (terbaru)
- Edge (terbaru)

Browser harus mendukung ES6 Modules.