# gw-suarahati
Bot telegram penghantar suara hati ke mastodon

## Installasi
- [Go](https://go.dev) telah diinstal di sistem
- Bot mastodon telah disiapkan
- Bot telegram telah disiapkan
- **Privacy Mode** di bot telegram sudah dimatikan
- Satu chat grup telah disiapkan

Salin `config.yaml.example` menjadi `config.yaml`, Lalu edit file `config.yaml`, Isi kolom yang diperlukan.

Mulai mengcompile:
```
go build -v
```

Lalu jalankan:
```
./gw-suarahati
```
