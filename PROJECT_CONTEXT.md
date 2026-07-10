# Legender Proje Bağlamı

> Amaç: Bu dosya, projeyi sonraki çalışmalarda baştan sona tekrar okumadan hızlıca hatırlamak için oluşturulmuştur. Kod değiştikçe ilgili bölümler güncellenmelidir.

## 1. Projenin amacı

Legender, Windows üzerinde çalışan bir League of Legends masaüstü yardımcısıdır. Riot League Client'ın yerel LCU (League Client Update) API'sine bağlanır ve şimdilik şu işleri hedefler:

- League Client'ın açık olup olmadığını tespit etmek.
- Oturum açmış oyuncunun adını, seviyesini ve profil ikonunu göstermek.
- Maç bulunduğunda hazır kontrolünü otomatik kabul etmek.
- Şampiyon seçiminde birincil, ikincil veya kullanıcı havuzundan rastgele bir şampiyonu otomatik kilitlemek.
- Kullanıcı tercihlerini `settings.json` içinde kalıcı tutmak.

Uygulama erken aşamadadır; kök `README.md` hâlâ varsayılan Wails şablon metnidir ve otomatik test bulunmamaktadır.

## 2. Teknoloji yığını

- Masaüstü kabuğu ve JS/Go köprüsü: Wails v2.11
- Backend: Go 1.23
- LCU WebSocket: `github.com/gorilla/websocket`
- Frontend: React 18 + TypeScript
- Geliştirme/bundle: Vite
- Stil: Tailwind CSS 4 ve az miktarda global CSS
- Pencere/UI: Frameless Wails penceresi, sürüklenebilir özel title bar ve flat koyu tasarım sistemi
- Hedef işletim sistemi: Windows; LCU keşfi PowerShell ve WMI/CIM kullandığı için mevcut hali platform bağımlıdır.

## 3. Üst düzey mimari

```text
React UI
  Home.tsx / ChampionModal.tsx
          |
          v
frontend/src/services/lcu.ts
          |
          v  Wails tarafından üretilen binding
app.go (main.App)
          |
          +--> internal/config.Store  <--> %AppData%/Legender/settings.json
          |
          +--> internal/lcu.Client
                    |
                    +--> HTTPS LCU endpointleri
                    +--> WSS LCU eventleri

Harici görsel/veri kaynakları:
  Riot Data Dragon (şampiyon listesi ve görselleri)
  CommunityDragon (profil ikonları)
```

Go tarafında dışarı açılan `App` metotları Wails tarafından `frontend/wailsjs/go/main/App.*` dosyalarına üretilir. Üretilen `frontend/wailsjs` dosyaları elle düzenlenmemelidir.

## 4. Çalışma akışları

### Uygulama açılışı

1. `main.go`, `frontend/dist` içeriğini binary içine gömer ve 1024x768 Wails penceresini açar.
2. `app.startup`, Wails context'ini saklar ve uygulamaya ait `config.Store` örneğini yükler.
3. `Home.tsx` mount olduğunda ayarları Go tarafından alır.
4. Ortak `services/champions.ts`, LCU'dan gelen oyun sürümü ve locale ile uyumlu Data Dragon kataloğunu seçer; istek ve son başarılı katalog WebView içinde cache'lenir.
5. `useLCU`, bağlantıyı açılışta ve ardından 5 saniyede bir kontrol eder; **Bağlantıyı Yenile** düğmesi ayrıca manuel kontrol sağlar.

### LCU bağlantısı

1. `internal/lcu/connector.go`, ilk keşifte `LeagueClientUx.exe` yolunu PowerShell/CIM ile bulur ve lockfile yolunu `%AppData%/Legender/lcu-lockfile-path` içinde saklar.
2. Sonraki kontroller ve uygulama açılışları doğrudan League `lockfile` dosyasını okur ve PID'yi Windows API ile doğrular; PowerShell yalnızca kayıtlı yol çalışmıyorsa kontrollü aralıkla kullanılır.
3. `app.GetLCUData()`, port veya token değiştiyse eski context'i iptal eder, yeni `lcu.Client` oluşturur ve WebSocket dinleyicisini goroutine olarak başlatır.
4. `GET /lol-summoner/v1/current-summoner` ile oyuncu bilgisi alınır.
5. LCU'nun self-signed sertifikası nedeniyle hem HTTPS hem WSS istemcisinde TLS doğrulaması kapalıdır. Trafik yalnızca `127.0.0.1` adresine gider.
6. WebSocket geçici olarak koparsa 1-30 saniye aralığında artan gecikmeyle yeniden bağlanır; uygulama kapanırken context üzerinden durdurulur.
7. `/riotclient/region-locale` ve `/lol-patch/v1/game-version` bir kez okunup client üzerinde cache'lenir; frontend bu metadata ile CDN sürüm/locale seçer.

### Otomatik maç kabulü

1. WebSocket, `OnJsonApiEvent_lol-matchmaking_v1_ready-check` olayına abone olur.
2. Event verisinde `state == "InProgress"`, ayarda `autoAccept == true` ve `playerResponse == "None"` ise kabul isteği gönderilir.
3. Kullanılan endpoint: `POST /lol-matchmaking/v1/ready-check/accept`.

### Otomatik şampiyon seçimi

Amaçlanan sıra:

1. WebSocket `OnJsonApiEvent_lol-champ-select_v1_session` olayına abone olur ve `BAN_PICK` fazında yerel oyuncunun aktif `pick` action'ını bulur.
2. `GET /lol-champ-select/v1/pickable-champion-ids` çağrılır.
3. Seçilebilir durumdaysa `primaryChamp`, değilse `secondaryChamp` kullanılır.
4. İkisi de uygun değilse `champPool` içindeki seçilebilir adaylardan rastgele biri alınır.
5. `PATCH /lol-champ-select/v1/session/actions/{actionId}` ve `{ "championId": ..., "completed": true }` ile doğrudan kilitlenir.
6. Aynı `gameId:actionId` eventinin tekrarı bellekte izlenerek ikinci kez kilitleme isteği gönderilmez; session silinince kayıt temizlenir.

### Ayarların saklanması

Şema:

```json
{
  "schemaVersion": 1,
  "autoAccept": false,
  "autoPick": {
    "enabled": false,
    "primaryChamp": 0,
    "secondaryChamp": 0,
    "champPool": []
  }
}
```

- `GetSettings`: tüm ayarları frontend'e döndürür.
- `ToggleAutoAccept`: otomatik kabul durumunu değiştirip dosyaya yazar.
- `SaveAutoPickSettings`: hızlı seçim ayarlarını topluca değiştirip dosyaya yazar.
- Ayarlar instance tabanlı, mutex korumalı `config.Store` tarafından yönetilir; frontend ve WebSocket bağımsız snapshot'lar okur.
- Ana dosya `%AppData%/Legender/settings.json` konumundadır. Eski çalışma dizini `settings.json` dosyası ilk açılışta kopyalanır ve güvenlik için silinmez.
- `schemaVersion` gelecekteki migration'ları ayırır; v0 dosyaları v1'e normalize edilir, uygulamadan daha yeni şemalar reddedilir.

## 5. Önemli dosyalar

| Dosya | Sorumluluk |
|---|---|
| `main.go` | Frameless Wails penceresi, minimum boyutlar, gömülü frontend assetleri ve uygulama başlangıcı |
| `app.go` | Frontend'e açılan Go API'si, LCU client yaşam döngüsü |
| `internal/config/settings.go` | Eşzamanlı güvenli JSON ayar store'u ve snapshot yönetimi |
| `internal/diagnostics/log.go` | Dönen AppData logu ve logger kurulumu |
| `internal/lcu/connector.go` | Windows lockfile/PID tabanlı League Client keşfi |
| `internal/lcu/connector_unsupported.go` | Windows dışı platformlar için açık desteklenmiyor sınırı |
| `internal/lcu/client.go` | LCU HTTPS çağrıları ve veri modelleri |
| `internal/lcu/ws.go` | WAMP/WebSocket event döngüsü, auto-accept ve auto-pick otomasyonları |
| `frontend/src/services/lcu.ts` | Wails bindingleri üzerinde ince servis katmanı |
| `frontend/src/services/champions.ts` | Data Dragon şampiyon verisi, dönüşümü ve istek cache'i |
| `frontend/src/services/champions.test.ts` | CDN locale, patch eşleştirme ve fallback sırası testleri |
| `frontend/src/components/FallbackImage.tsx` | CommunityDragon locale/version görsel fallback zinciri |
| `frontend/src/components/TitleBar.tsx` | Sürüklenebilir özel başlık alanı ve pencere kontrolleri |
| `frontend/src/hooks/useLCU.ts` | Açılış/polling/manüel LCU bağlantı kontrolü ve state'i |
| `frontend/src/pages/Home.tsx` | Ana ekran, ayar state'i, profil ve otomasyon kartları |
| `frontend/src/components/ChampionModal.tsx` | Data Dragon şampiyon araması ve tekli/çoklu seçim modalı |
| `frontend/wailsjs/` | Wails tarafından üretilen JS/TS bindingleri ve modeller |
| `%AppData%/Legender/settings.json` | Çalışma zamanı kullanıcı tercihleri |
| `wails.json` | Wails build/dev komutları ve uygulama metadata'sı |

## 6. Frontend state özeti

`Home.tsx` şu state'leri yönetir:

- `lcuData`, `isLoading`: `useLCU` hook'undan gelir.
- `isAutoAcceptOn`: UI toggle durumu.
- `autoPick`: `enabled`, `primaryChamp`, `secondaryChamp`, `champPool`.
- `modalConfig`: modal açık/kapalı ve seçim türü (`primary`, `secondary`, `pool`).
- `championsData`: seçilmiş ID'leri isim ve görsele çevirmek için kullanılan Data Dragon listesi.

`ChampionModal`, `Home` tarafından verilen ortak şampiyon listesini kullanır. Modal her açıldığında arama ve çoklu seçim state'i güncel prop değerleriyle sıfırlanır.

## 7. Geliştirme ve doğrulama komutları

Proje kökü: `legender/`

```powershell
# Wails canlı geliştirme
wails dev

# Üretim uygulaması
wails build

# Yalnızca frontend
Set-Location frontend
npm run dev
npm run build
npm test

# Go test ve race kontrolü
Set-Location ..
go test ./...
go test -race ./...
go vet ./...

# İsteğe bağlı: açık gerçek istemcide yalnızca discovery/metadata/summoner GET testi
$env:LEGENDER_LCU_INTEGRATION = '1'
go test ./internal/lcu -run '^TestRealLCUReadOnly$' -v
```

Frontend build sırası `tsc && vite build` şeklindedir. Wails production build, oluşan `frontend/dist` klasörünü Go binary'sine gömer.

## 8. Bilinen riskler ve teknik borç

Koddan doğrudan görülebilen kalan başlıca noktalar:

1. **Platform kapsamı:** Platform kodu build-tag ile ayrıldı ve Windows dışı derleme açık hata verir; gerçek League Client discovery desteği yalnızca Windows'tadır.
2. **Gerçek istemci write-flow testi:** Discovery, metadata ve summoner salt-okunur akışı gerçek League Client ile doğrulanmıştır. Ready-check kabulü ve champ-select kilitleme POST/PATCH akışları gerçek maç gerektirdiği için hâlâ manuel doğrulama gerektirir.
3. **Harici CDN sözleşmesi:** Locale normalizasyonu, patch eşleştirmesi ve CommunityDragon fallback sırası Vitest ile korunur; Riot/CommunityDragon dizin yapısı değişirse gerçek ağ smoke testi gerekir.
4. **Metin encoding işaretleri:** Bazı terminal kurulumlarında Türkçe kaynak metinleri mojibake görünebilir; editör ve terminal UTF-8 ayarı korunmalıdır.

## 9. Değişiklik yaparken dikkat edilecekler

- Wails bindinglerini elle değiştirme; Go public metot/model değişikliklerinden sonra Wails ile yeniden üret.
- LCU token'ını loglama, UI'ya döndürme veya repoya yazma. Şu anda `LCUResponse` yalnızca port ve kullanıcı bilgisini döndürüyor.
- LCU otomasyonunu değiştirirken event aboneliği, action phase ve tekrar tetiklenme/idempotency davranışını birlikte kontrol et.
- `settings.json` kişisel çalışma zamanı verisi içerir. Yeni alan eklerken eski dosyaların eksik alanlarla yüklenmesini ve varsayılan/migration davranışını düşün.
- Şampiyon kimlikleri Data Dragon'daki sayısal `key` alanından gelir.
- Frontend'de gösterilen port hassas token değildir ama gerekmiyorsa kullanıcıya göstermemek düşünülebilir.
- Değişiklik sonrası en az `go test ./...` ve `frontend` altında `npm run build` çalıştır.
- Gerçek auto-accept/auto-pick davranışı ancak çalışan League Client ve uygun oyun akışıyla uçtan uca doğrulanabilir.

## 10. Son analiz durumu

- Bu bağlam dosyası 10 Temmuz 2026 tarihinde mevcut kaynak kod okunarak oluşturuldu.
- Öncelikli mimari düzeltmeler uygulandı: WebSocket abonelik/phase hatası, reconnect ve iptal yaşam döngüsü, thread-safe config store, HTTP timeout/status yönetimi, otomatik bağlantı takibi ve ortak Data Dragon cache'i.
- LCU sürüm/locale metadata'sı CDN seçimine bağlandı; CommunityDragon görsellerine fallback zinciri, Data Dragon'a en yeni beş kataloğu tutan kalıcı offline cache eklendi.
- Ayarlar AppData dizinine taşındı; legacy migration, schema version ve temel şampiyon ID normalizasyonu eklendi. Frontend hataları kullanıcıya banner ile gösterilir.
- LCU polling ilk keşiften sonra PowerShell yerine lockfile + Windows PID kontrolü kullanır; lockfile yolu AppData'da uygulama açılışları arasında saklanır. Windows'a özgü discovery build-tag sınırına alındı.
- Loglar `%AppData%/Legender/logs/legender.log` konumunda 2 MB'da döndürülür; hata banner'ından kullanıcı seçtiği konuma dışa aktarılabilir.
- Açık League Client ile salt-okunur entegrasyon testi discovery, `/riotclient/region-locale`, `/lol-patch/v1/game-version` ve current-summoner akışında başarılıdır. Eski `/riotclient/get_region_locale` URI'sinin 404 döndüğü gerçek istemcide görülerek güncel URI'ye geçilmiştir.
- Go testleri store concurrency/migration, metadata cache, discovery path, log rotation, seçim önceliği, gerçek HTTP test sunucusu üzerinden champ-select kilitleme ve tekrar-event korumasını kapsar. Vitest CDN locale/version/fallback kurallarını kapsar.
- Aynı tarihte `go test -race ./...`, `go vet ./...`, `npm test`, salt-okunur gerçek LCU entegrasyon testi, Linux hedefi için `internal/lcu` test derlemesi ve tam `wails build` başarıyla tamamlandı.
- Sonraki önemli değişikliklerde bu dosyanın özellikle mimari, akışlar, bilinen riskler ve komutlar bölümleri güncellenmelidir.
