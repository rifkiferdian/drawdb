# DrawDB

DrawDB adalah aplikasi web monolith untuk merancang Entity Relationship Diagram (ERD) secara visual. Aplikasi dibangun menggunakan Go, Gin, Go HTML Template, Vanilla JavaScript, Tailwind CSS v4, dan MySQL.

Tidak ada login atau role. Project dan seluruh data diagram disimpan langsung ke MySQL.

## Fitur

### Project

- Membuat, mengganti nama, menduplikasi, dan menghapus project.
- Membuka diagram langsung dari daftar project.
- Penyimpanan otomatis ke MySQL.

### Editor ERD

- Membuat, mengedit, menduplikasi, dan menghapus tabel.
- Memindahkan tabel menggunakan drag-and-drop.
- Menambah, mengedit, menghapus, dan mengurutkan kolom.
- Dukungan tipe data MySQL dan PostgreSQL.
- Length, precision, scale, default value, dan komentar.
- Primary key, termasuk composite primary key.
- Nullable, unique, auto increment, dan identity.
- Pembuatan dan penghapusan index.
- Relasi one-to-one dan one-to-many.
- Junction table otomatis untuk many-to-many.
- Membuat relasi dengan menarik port antarkolom atau melalui panel detail.
- Pencarian tabel dan kolom ketika membuat relasi.
- Pengaturan `ON DELETE` dan `ON UPDATE`.
- Validasi kompatibilitas tipe data foreign key.
- Pan, zoom, fit-to-screen, dan auto layout.
- Undo dan redo.

### Import dan export

- Import dan export diagram JSON.
- Import SQL MySQL dan PostgreSQL.
- Preview dan export SQL MySQL atau PostgreSQL.
- Export diagram sebagai PNG.

## Teknologi

- Go 1.25+
- Gin
- MySQL 8+
- Go HTML Template
- Vanilla JavaScript
- Tailwind CSS v4 CLI
- Air untuk live reload (opsional)

## Prasyarat

- [Go](https://go.dev/dl/) 1.25 atau lebih baru.
- [Node.js](https://nodejs.org/) 20 atau lebih baru.
- MySQL 8 atau lebih baru.

## Instalasi

### 1. Buat database

Dengan password:

```powershell
mysql -u root -p -e "CREATE DATABASE drawdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
```

Tanpa password:

```powershell
mysql -u root -e "CREATE DATABASE drawdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
```

Migration tabel `projects` dijalankan otomatis saat aplikasi dimulai. File migration juga tersedia di `migrations/001_create_projects.sql`.

### 2. Siapkan environment

```powershell
Copy-Item .env.example .env
```

Contoh konfigurasi untuk user `root` tanpa password:

```env
APP_ADDR=:8080
DB_DSN=root:@tcp(127.0.0.1:3306)/drawdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

Format umum DSN:

```text
username:password@tcp(host:port)/database?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

File `.env` berisi credential lokal dan sudah diabaikan oleh Git.

### 3. Instal dependency dan build CSS

```powershell
go mod download
npm install
npm run css:build
```

### 4. Jalankan aplikasi

```powershell
go run .
```

Buka [http://localhost:8080](http://localhost:8080).

## Development dengan Air

Air akan memuat `.env`, membangun ulang Tailwind CSS, mengompilasi aplikasi Go, dan me-restart server ketika source berubah.

Instal Air satu kali:

```powershell
go install github.com/air-verse/air@latest
```

Jalankan dari root project:

```powershell
air
```

Jika perintah `air` belum dikenali, tambahkan `<hasil go env GOPATH>\bin` ke `PATH`, atau jalankan:

```powershell
go run github.com/air-verse/air@latest
```

Konfigurasi live reload berada di `.air.toml`.

## Development tanpa Air

Jalankan watcher Tailwind pada terminal pertama:

```powershell
npm run css:watch
```

Jalankan server pada terminal kedua:

```powershell
go run .
```

Perubahan file Go, template, atau JavaScript memerlukan restart server karena file web di-embed ke binary Go.

## Build production

```powershell
npm ci
npm run css:build
go build -o drawdb.exe .
./drawdb.exe
```

Pastikan `.env` tersedia di direktori aplikasi, atau sediakan `APP_ADDR` dan `DB_DSN` melalui environment sistem.

## Struktur project

```text
drawdb/
|-- main.go                    # Server, routing, API, dan repository MySQL
|-- migrations/               # Migration database
|-- web/
|   |-- assets/css/            # Source Tailwind CSS
|   |-- static/css/            # CSS hasil build
|   |-- static/js/             # Editor ERD dan interaksi halaman
|   `-- templates/             # Go HTML Template
|-- .air.toml                  # Konfigurasi live reload
|-- .env.example               # Contoh environment
|-- go.mod                     # Dependency Go
`-- package.json               # Dependency dan script Tailwind
```

## Routes

| Method | Route | Keterangan |
| --- | --- | --- |
| `GET` | `/` | Daftar project |
| `GET` | `/projects/:id` | Editor diagram |
| `POST` | `/api/projects` | Membuat project |
| `PUT` | `/api/projects/:id` | Mengganti nama project |
| `DELETE` | `/api/projects/:id` | Menghapus project |
| `POST` | `/api/projects/:id/duplicate` | Menduplikasi project |
| `PUT` | `/api/projects/:id/diagram` | Menyimpan diagram |

## Pemeriksaan source

```powershell
node --check web/static/js/editor.js
node --check web/static/js/projects.js
go test ./...
go vet ./...
```

## Troubleshooting

### `DB_DSN is required`

Pastikan file `.env` tersedia dan memiliki nilai `DB_DSN`.

### Database connection failed

Pastikan MySQL aktif, database `drawdb` sudah dibuat, serta username dan password pada `DB_DSN` benar.

### Perubahan tampilan belum terlihat

Build ulang CSS:

```powershell
npm run css:build
```

Lakukan hard refresh dengan `Ctrl + F5`. Jika tidak menggunakan Air, restart aplikasi karena asset web di-embed ke binary.

### Perintah `air` tidak ditemukan

Gunakan `go run github.com/air-verse/air@latest` atau tambahkan direktori `GOPATH\bin` ke `PATH`.
