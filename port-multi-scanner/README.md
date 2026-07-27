# Port Multi Scanner

`ip_feed.json` icindeki IP adresleri TCP portlariyla taranir. Is sirasi her zaman **port -> tum IP'ler** seklindedir.

## Calistirma

```powershell
cd frontend
npm install
npm run build
cd ..
go run ./backend
```

Panel: `http://localhost:8080`

## E-posta raporu

Panel, ekrandaki acik-port sonuclarini (tarama noktasi, musteri, bayi, cihaz, IP, port ve tespit zamani) e-posta olarak gonderebilir. SMTP bilgilerini kaynak koda koymadan ortam degiskenlerinde tanimlayin:

```powershell
$env:SMTP_HOST="smtp.example.com"
$env:SMTP_PORT="587"
$env:SMTP_USERNAME="scanner@example.com"
$env:SMTP_PASSWORD="uygulama-sifresi"
$env:SMTP_FROM="scanner@example.com"
go run ./backend
```

## Turkiye ve yurtdisi tarama noktalari

Her kurulumun bir `SCANNER_NAME` degeri vardir. Bu deger ekranda, acik-port sonucunda ve e-posta raporunda **Kaynak** olarak gorunur.

- Yerel kurulum: `SCANNER_NAME=Turkiye`
- DigitalOcean kurulumu: `SCANNER_NAME=Yurtdisi - Frankfurt` (veya secilen bolge)

DigitalOcean icin Docker imaji hazirdir:

```powershell
docker build -t port-multi-scanner .
docker run -d --name port-scanner -p 8080:8080 --env-file .env -e SCANNER_NAME="Yurtdisi - Frankfurt" port-multi-scanner
```

DigitalOcean sunucusu da ayni `ip_feed.json` dosyasina sahip olmali. Iki konum ayni araligi bagimsiz tarar; sonuc tablosundaki Kaynak sutunu, portun hangi konumdan goruldugunu ayirt eder.

Merkez panelde iki sonucu tek tabloda gostermek icin DigitalOcean sunucusunda `SCANNER_API_TOKEN` tanimlayin. Yerel/merkez kurulumunda ayni degeri `REMOTE_SCANNER_TOKEN` ve DigitalOcean adresini `REMOTE_SCANNER_URL` olarak verin. Merkezden baslatilan veya durdurulan tarama, uzak tarayiciya da iletilir; merkez paneli iki konumun sonucunu birlestirir.

> Yalnizca taramaya yetkili oldugunuz sistem ve aglarda kullanin.
