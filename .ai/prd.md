# Dokument wymagań produktu (PRD) - Ask Your Feed (MVP)

## 1. Przegląd produktu

Ask Your Feed to webowa aplikacja Q&A nad prywatnym feedem użytkownika z portalu X. System cyklicznie agreguje wyłącznie oryginalne wpisy obserwowanych kont (bez reply/retweet/repost/quote, z wyjątkiem wątków self-reply), konwertuje treści multimedialne do tekstu, zapisuje je w relacyjnej bazie i udostępnia prosty interfejs do zadawania pytań. Odpowiedzi generowane są przez LLM w trybie feed-only (bez przeglądania WWW) w oparciu o materiały z określonego zakresu czasu. Każda odpowiedź zawiera listę punktów oraz sekcję Źródła z linkami do postów. Aplikacja zapewnia historię zapytań/odpowiedzi z możliwością usuwania.

## 2. Problem użytkownika

Użytkownicy konsumują feed X z różnymi intencjami: szybki przegląd dnia vs. głębokie wejście w wątki od konkretnych osób i tematów. Manualne filtrowanie i wyszukiwanie jest czasochłonne, media (obrazy/wideo) wymagają interpretacji, a wątki i szum (odpowiedzi, retweety) utrudniają skrótowe wnioskowanie. Ask Your Feed automatyzuje selekcję i syntezę najważniejszych informacji z własnego feedu użytkownika oraz pozwala szybko zadawać pytania i otrzymywać odpowiedzi z cytowanymi źródłami.

## 3. Wymagania funkcjonalne

### 3.1. Konta i uwierzytelnianie

**Model uwierzytelniania:**
* Rejestracja i logowanie przez email + hasło (standardowa autentykacja aplikacji).
* Podczas rejestracji użytkownik podaje swój username X (np. "elonmusk") jako parametr profilu.
* Mapowanie 1:1: użytkownik aplikacji ↔ username X.
* Brak OAuth X – system używa twitterapi.io API z kluczem API do pobierania publicznych danych.

**Bezpieczeństwo:**
* Szyfrowane przechowywanie haseł użytkowników (bcrypt/argon2).
* Bezpieczne składowanie klucza API twitterapi.io (zmienne środowiskowe, secrets manager).
* Sesje użytkowników z tokenami JWT.
* Walidacja username X podczas rejestracji (sprawdzenie czy konto istnieje przez `/twitter/user/info`).

### 3.2. Agregacja i ingest feedu

**Proces agregacji:**
1. Pobranie listy obserwowanych użytkowników przez endpoint `/twitter/user/followings?userName={username}`.
2. Dla każdego obserwowanego użytkownika pobranie ostatnich tweetów przez `/twitter/user/last_tweets?userName={followed_username}`.
3. Filtrowanie i zapisanie tylko oryginalnych postów zgodnie z regułami poniżej.

**Zakres i filtry:**
* **Tylko oryginalne wpisy:** Zapisywane są tweety spełniające warunki:
  - `isReply === false` ALBO (`isReply === true` AND `inReplyToUserId === author.id`) – wątki (self-replies) są dozwolone
  - Brak pola `retweeted_tweet` (wykluczenie retweetów)
  - Brak pola `quoted_tweet` (wykluczenie quote tweetów)
* **Limit obserwowanych:** Maksymalnie 150 obserwowanych użytkowników dla MVP. Jeśli użytkownik obserwuje więcej, system pobiera dane tylko dla pierwszych 150 (sortowanie według aktywności lub daty dodania do followings).
* **Edycje:** Ignorowane (twitterapi.io zwraca aktualną wersję tweeta).

**Harmonogram:**
* **Regularny ingest:** Co 4 godziny ±15 min (jitter dla rozłożenia obciążenia).
* **Paginacja:** 
  - Podczas regularnego ingestu: tylko pierwsza strona dla każdego followed user (~20 najnowszych tweetów).
  - Podczas backfillu: paginacja do momentu osiągnięcia 24h wstecz lub braku kolejnych stron (`has_next_page === false`).
* **Filtrowanie czasowe:** Tweety filtrowane po polu `createdAt` (brak parametru `since_id` w twitterapi.io).
* **Odporność na błędy:** Retry z eksponentialnym backoffem przy błędach API (429, 5xx).

**SLO:**
* Średni czas od publikacji do widoczności w aplikacji ≤4.5 h (uwzględniając interwał 4h + czas przetwarzania).

**Backfill przy pierwszym uruchomieniu:**
* Domyślnie ostatnie 24h.
* Paginacja przez wszystkie strony dla każdego followed user do osiągnięcia 24h wstecz.

### 3.3. Przechowywanie danych

**Baza danych:**
* Relacyjna baza danych (PostgreSQL).
* Pełnotekstowe przechowywanie treści postów (`text` field z Tweet object).
* Przechowywanie metadanych: `id`, `url`, `createdAt`, `author` (userName, name, id), `likeCount`, `retweetCount`, `replyCount`, `viewCount`.

**Media:**
* **Konwersja do tekstu:** Obrazy i krótkie wideo konwertowane do tekstu za pomocą zewnętrznego API Vision (np. OpenAI OpenRouter, Google Cloud Vision).
* **Limity:** 
  - Wideo: ≤90 s lub ≤25 MB
  - Obrazy: maksymalnie 4 na post
* **Przechowywanie:** Tylko tekstowe opisy mediów; brak przechowywania URL/miniatur.
* **Obsługa błędów:** Media przekraczające limity są pomijane; system kontynuuje przetwarzanie pozostałych danych.

### 3.4. Wyszukiwanie i indeksowanie

* Wyszukiwanie wyłącznie pełnotekstowe (PostgreSQL full-text search lub podobne).
* Brak indeksów wektorowych w MVP.
* Filtrowanie po dacie publikacji (`createdAt`).

### 3.5. Pytania i odpowiedzi (Q&A)

* LLM działa w trybie feed-only (bez web-browsingu).
* Domyślny horyzont czasowy: 24h; użytkownik może zmienić zakres dat.
* Format odpowiedzi: lista punktów + sekcja Źródła (≥3 posty, jeśli dostępne; w przeciwnym razie wszystkie dostępne).
* Alternatywne formaty wymuszone komendą w promptcie (np. „timeline").
* Przy braku materiału zwracany jest komunikat z sugestią rozszerzenia zakresu.
* Źródła zawierają linki do oryginalnych postów w formacie `https://twitter.com/{userName}/status/{tweetId}`.

### 3.6. Historia

* Lista zapytań/odpowiedzi z paginacją i podglądem odpowiedzi.
* Usuwanie pojedynczego wpisu i „Usuń wszystko".
* Brak automatycznej retencji i brak eksportu.

### 3.7. Telemetria

* Telemetria wyłącznie operacyjna, bez PII.
* Metryki: latencja, error rate, rate-limit hits (429), koszty API (liczba requestów, zwróconych tweetów/profili).

### 3.8. Platforma i UX

* Responsywna aplikacja web (PL), dark mode.
* Proste, spójne komunikaty błędów i stanów:
  - Brak materiału w wybranym zakresie
  - Rate-limit API (429)
  - Wyczerpany budżet kosztowy
  - Błąd walidacji username X
  - Błędy API twitterapi.io
* Wskazanie czasu ostatniej aktualizacji danych (np. „Ostatnia synchronizacja: 12:15").
* Informacja o liczbie obserwowanych użytkowników i limicie (np. "Synchronizujesz 150/150 obserwowanych").

### 3.9. Bezpieczeństwo

* Szyfrowane przechowywanie haseł użytkowników.
* Bezpieczne składowanie klucza API twitterapi.io (secrets manager, zmienne środowiskowe).
* Walidacja i sanityzacja danych wejściowych (username X, zapytania użytkownika).
* Rate limiting na poziomie aplikacji (ochrona przed nadużyciami).

## 4. Granice produktu

### 4.1. Poza zakresem MVP

* Brak aktualizacji feedu w czasie rzeczywistym; wyłącznie interwałowo (co 4h).
* Zaawansowane filtrowanie inne niż przez LLM i datę.
* Web-browsing i źródła spoza feedu użytkownika.
* Eksport historii i automatyczna retencja.
* Indeksy wektorowe, RAG, personalizowane rankingi.
* Analityka zachowań użytkowników (np. DAU/WAU) – poza telemetrią techniczną.
* Wielokrotne konta X na jednego użytkownika (mapowanie 1:1).
* OAuth X (zastąpione przez email/hasło + username X).
* Obsługa więcej niż 150 obserwowanych użytkowników.

### 4.2. Ograniczenia i decyzje architektoniczne

**Filtry ingestu:**
* Tylko oryginalne wpisy (bez retweetów i quote tweetów).
* Wątki (self-replies) są dozwolone jako wyjątek.
* Edycje ignorowane (API zwraca aktualną wersję).

**Media:**
* Konwersja do tekstu przez zewnętrzne API Vision.
* Limity: wideo ≤90 s lub ≤25 MB; ≤4 obrazy/post.
* Media przekraczające limity są pomijane.

**Koszty i limity:**
* Maksymalnie 150 obserwowanych użytkowników.
* Ingest co 4h (6 razy dziennie).
* Budżet kosztowy jest twardym limitem; po wyczerpaniu następuje pauza i blokada.
* Szacunkowy koszt przy 150 followings i 6 ingestów/dzień:
  - 150 requestów × 6 = 900 requestów/dzień
  - ~$0.135/dzień/użytkownik (~$4/miesiąc/użytkownik przy samym ingeście)
  - Dodatkowe koszty: OpenRouter, LLM

**Źródła:**
* Zawsze linkują do postów z feedu (format URL z twitterapi.io: `https://twitter.com/{userName}/status/{tweetId}`).

**API twitterapi.io:**
* Brak parametru `since_id` – filtrowanie po `createdAt`.
* Paginacja przez `cursor` i `has_next_page`.
* Pricing: $0.15/1k tweets, $0.18/1k user profiles, minimum $0.00015/request.

## 5. Historyjki użytkowników

### US-001
**Tytuł:** Rejestracja z email, hasłem i username X  
**Opis:** Jako nowy użytkownik chcę zarejestrować się w aplikacji, podając email, hasło i mój username X, aby system mógł agregować mój feed.  
**Kryteria akceptacji:**
- Formularz rejestracji zawiera pola: email, hasło, potwierdzenie hasła, username X.
- System waliduje username X przez endpoint `/twitter/user/info` (sprawdzenie czy konto istnieje).
- Jeśli username X nie istnieje, wyświetlany jest komunikat błędu.
- Po udanej rejestracji użytkownik jest zalogowany i przekierowany do dashboardu.
- System automatycznie rozpoczyna pierwszy backfill (24h).

### US-002
**Tytuł:** Logowanie email + hasło  
**Opis:** Jako zarejestrowany użytkownik chcę zalogować się do aplikacji używając email i hasła.  
**Kryteria akceptacji:**
- Formularz logowania zawiera pola: email, hasło.
- Po udanym logowaniu użytkownik jest przekierowany do dashboardu.
- Nieprawidłowe dane skutkują komunikatem błędu.
- Sesja użytkownika jest bezpiecznie zarządzana (JWT lub podobne).

### US-003
**Tytuł:** Onboarding i pierwszy backfill  
**Opis:** Jako nowy użytkownik chcę, by system zaczął pobierać mój feed i przygotował dane do Q&A.  
**Kryteria akceptacji:**
- Po pierwszej rejestracji backfill uruchamia się automatycznie.
- Zakres backfillu: ostatnie 24h.
- System pobiera listę followings (max 150) i dla każdego pobiera tweety z paginacją do 24h wstecz.
- Widzę wskaźnik postępu backfillu (np. "Synchronizacja: 45/150 użytkowników").
- Po zakończeniu widzę "Ostatnia synchronizacja: [timestamp]".

### US-004
**Tytuł:** Limit 150 obserwowanych użytkowników  
**Opis:** Jako użytkownik obserwujący więcej niż 150 kont chcę wiedzieć, że system synchronizuje tylko 150 z nich.  
**Kryteria akceptacji:**
- Jeśli użytkownik obserwuje >150 kont, system pobiera tylko pierwsze 150.
- UI wyświetla informację: "Synchronizujesz 150/150 obserwowanych (limit MVP)".
- Komunikat wyjaśnia, że pełna synchronizacja będzie dostępna w przyszłych wersjach.

### US-005
**Tytuł:** Regularny ingest co 4h  
**Opis:** Jako użytkownik chcę, by mój feed był regularnie aktualizowany.  
**Kryteria akceptacji:**
- System wykonuje ingest co 4h ±15 min.
- Podczas regularnego ingestu pobierana jest tylko pierwsza strona tweetów dla każdego followed user.
- Wskaźnik "Ostatnia synchronizacja" aktualizuje się po każdym cyklu.
- Użytkownik może zobaczyć czas następnej synchronizacji (opcjonalnie).

### US-006
**Tytuł:** Szybkie pytanie z domyślnym zakresem 24h  
**Opis:** Jako użytkownik chcę zadać pytanie i otrzymać odpowiedź z ostatnich 24h.  
**Kryteria akceptacji:**
- Gdy wpiszę pytanie bez parametrów, odpowiedź opiera się na danych z 24h.
- Odpowiedź jest listą punktów i zawiera sekcję Źródła.
- Źródła zawierają linki w formacie `https://twitter.com/{userName}/status/{tweetId}`.

### US-007
**Tytuł:** Filtrowanie po dacie  
**Opis:** Jako użytkownik chcę ustawić zakres dat dla Q&A.  
**Kryteria akceptacji:**
- UI pozwala wskazać od-do lub predefiniowane zakresy (24h, 7 dni, 30 dni).
- LLM używa tylko postów z wybranego zakresu (filtrowanie po `createdAt`).
- Źródła w odpowiedzi mieszczą się w tym zakresie.

### US-008
**Tytuł:** Alternatywny format odpowiedzi komendą w promptcie  
**Opis:** Jako użytkownik chcę wymusić format (np. „timeline").  
**Kryteria akceptacji:**
- Wpisanie komendy w promptcie zmienia format odpowiedzi zgodnie z dokumentacją.
- Sekcja Źródła nadal jest dołączona, o ile nie określono inaczej.

### US-009
**Tytuł:** Sekcja Źródła z min. 3 linkami  
**Opis:** Jako użytkownik chcę widzieć referencje do postów, aby zweryfikować odpowiedź.  
**Kryteria akceptacji:**
- Jeśli dostępne ≥3 posty, Źródła zawierają co najmniej 3 linki.
- Jeśli dostępne <3 posty, Źródła zawierają wszystkie dostępne linki.
- Linki prowadzą do oryginalnych postów w X (format: `https://twitter.com/{userName}/status/{tweetId}`).

### US-010
**Tytuł:** Brak treści w danym zakresie  
**Opis:** Jako użytkownik chcę jasny komunikat, jeśli nie ma treści.  
**Kryteria akceptacji:**
- Gdy brak trafień, system nie podaje fałszywych treści i zwraca komunikat o jej braku.
- Komunikat sugeruje rozszerzenie zakresu dat.
- Nie pojawia się sekcja Źródła, jeśli brak jakichkolwiek postów.

### US-011
**Tytuł:** Uwzględnianie treści multimedialnych przez OpenRouter  
**Opis:** Jako użytkownik chcę, by obrazy i krótkie wideo z feedu były uwzględniane w odpowiedziach.  
**Kryteria akceptacji:**
- Media są konwertowane do tekstu podczas ingestu przez zewnętrzne OpenRouter (np. OpenAI Vision).
- Tekstowe opisy mediów są przechowywane w bazie i widoczne dla LLM.
- Źródła mogą odwoływać się do postów zawierających media.
- Wideo >90 s lub >25 MB jest pomijane z logowaniem ostrzeżenia.
- Posty z >4 obrazami mają przetwarzane tylko pierwsze 4.

### US-012
**Tytuł:** Historia zapytań – lista z paginacją  
**Opis:** Jako użytkownik chcę przejrzeć wcześniejsze pytania i odpowiedzi.  
**Kryteria akceptacji:**
- Widok historii pokazuje listę pozycji z paginacją.
- Każda pozycja zawiera skrót pytania, datę i link do podglądu odpowiedzi.
- Parametry paginacji są stałe w MVP.

### US-013
**Tytuł:** Podgląd odpowiedzi z historii  
**Opis:** Jako użytkownik chcę otworzyć pełną treść wcześniejszej odpowiedzi.  
**Kryteria akceptacji:**
- Kliknięcie pozycji historii otwiera szczegóły z pełnym tekstem i Źródłami.
- Widok szczegółów jest tylko do odczytu.

### US-014
**Tytuł:** Usuwanie pojedynczej pozycji z historii  
**Opis:** Jako użytkownik chcę usunąć wybraną odpowiedź z historii.  
**Kryteria akceptacji:**
- Kliknięcie „Usuń" przy pozycji wymaga potwierdzenia.
- Po potwierdzeniu pozycja znika i nie jest dostępna w UI.

### US-015
**Tytuł:** „Usuń wszystko" w historii  
**Opis:** Jako użytkownik chcę skasować całą historię.  
**Kryteria akceptacji:**
- Akcja wymaga potwierdzenia.
- Po potwierdzeniu historia jest pusta.
- Operacja nie wpływa na dane ingestu.

### US-017
**Tytuł:** Prywatność telemetrii  
**Opis:** Jako użytkownik nie chcę, aby aplikacja gromadziła moje PII w logach/metrykach.  
**Kryteria akceptacji:**
- Telemetria zawiera tylko metryki operacyjne (latencja, error rate, rate-limit hits, koszty API).
- Brak zapisów PII (email, username X) w telemetrycznych eventach.
- Logi zawierają tylko anonimowe identyfikatory użytkowników.

### US-018
**Tytuł:** Bezpieczne składowanie credentials  
**Opis:** Jako użytkownik chcę, by moje hasło i dane były zabezpieczone.  
**Kryteria akceptacji:**
- Hasła są hashowane (bcrypt/argon2) przed zapisem w bazie.
- Klucz API twitterapi.io przechowywany w secrets manager lub zmiennych środowiskowych.
- Brak plain-text credentials w kodzie lub logach.

### US-019
**Tytuł:** Wylogowanie  
**Opis:** Jako użytkownik chcę zakończyć sesję w aplikacji.  
**Kryteria akceptacji:**
- Kliknięcie „Wyloguj" unieważnia sesję aplikacji.
- Po wylogowaniu nie mam dostępu do historii bez ponownego logowania.

### US-020
**Tytuł:** Wskaźnik świeżości danych  
**Opis:** Jako użytkownik chcę wiedzieć, kiedy dane były ostatnio aktualizowane.  
**Kryteria akceptacji:**
- UI wyświetla znacznik czasu „Ostatnia synchronizacja: [timestamp]".
- Wartość aktualizuje się po każdym udanym cyklu ingestu.
- Opcjonalnie: czas następnej synchronizacji (np. "Następna za ~3h 45min").

### US-021
**Tytuł:** Tylko oryginalne wpisy jako źródła  
**Opis:** Jako użytkownik nie chcę widzieć retweetów/quote tweetów w źródłach.  
**Kryteria akceptacji:**
- Źródła w odpowiedzi pochodzą wyłącznie z oryginalnych wpisów obserwowanych kont.
- Wątki (self-replies) są dozwolone i mogą pojawiać się w źródłach.
- Retweets (`retweeted_tweet` present) są wykluczane.
- Quote tweets (`quoted_tweet` present) są wykluczane.

### US-022
**Tytuł:** Obsługa limitów mediów  
**Opis:** Jako użytkownik chcę stabilności, gdy media przekraczają limity.  
**Kryteria akceptacji:**
- Wideo przekraczające limity (>90s lub >25MB) nie jest przetwarzane; system działa dalej.
- Odpowiedzi nie odwołują się do pominiętych mediów.

### US-023
**Tytuł:** Błędy i stany systemowe w UI  
**Opis:** Jako użytkownik chcę proste i spójne komunikaty błędów.  
**Kryteria akceptacji:**
- Stany obsługiwane z czytelnymi komunikatami:
  - Brak treści w wybranym zakresie
  - Rate-limit API twitterapi.io (429)
  - Wyczerpany budżet kosztowy
  - Błąd walidacji username X (konto nie istnieje)
  - Błędy API twitterapi.io (5xx, timeout)
  - Błędy OpenRouter
- Komunikaty są w języku polskim, zgodne z dark mode.

### US-024
**Tytuł:** Zapytanie wykraczające poza feed  
**Opis:** Jako użytkownik chcę, aby system nie zmyślał odpowiedzi, jeśli temat nie występuje w moim feedzie.  
**Kryteria akceptacji:**
- Brak dowodów skutkuje komunikatem o braku treści.
- Źródła nigdy nie zawierają linków spoza feedu.
- LLM działa w trybie feed-only (bez web-browsingu).

### US-025
**Tytuł:** Walidacja username X podczas rejestracji  
**Opis:** Jako nowy użytkownik chcę otrzymać natychmiastowy feedback, jeśli podany username X nie istnieje.  
**Kryteria akceptacji:**
- System wywołuje `/twitter/user/info?userName={username}` podczas rejestracji.
- Jeśli endpoint zwraca błąd 400 lub "User not found", wyświetlany jest komunikat: "Konto X o nazwie '{username}' nie istnieje. Sprawdź poprawność nazwy użytkownika."
- Rejestracja nie może być ukończona bez poprawnego username X.

## 6. Metryki sukcesu

### 6.1. SLO i niezawodność

* **Czas od publikacji do widoczności:** ≤4.5h (uwzględniając interwał 4h + czas przetwarzania).
* **Stabilność ingestu:** 
  - Error rate <5% dla requestów do twitterapi.io
  - Liczba retry <10% wszystkich requestów
  - Rate-limit hits (429) <1% requestów
* **Dostępność Q&A:** >99% udanych odpowiedzi (bez błędów systemowych).

### 6.2. Doświadczenie użytkownika

* **Czas odpowiedzi Q&A:** <10s dla typowych zapytań (target do ustalenia w testach wydajności).
* **Jakość odpowiedzi:** 
  - Każda odpowiedź zawiera sekcję Źródła (≥3 linki, jeśli dostępne).
  - Źródła są relevantne do pytania (weryfikacja przez testy użytkowników).
* **Kompletność historii:** 
  - Brak błędów przy paginacji.
  - Działające usuwanie per wpis i „Usuń wszystko".
* **Backfill:** Ukończenie pierwszego backfillu (24h, max 150 followings) w <30 min.

### 6.3. Metryka produktowa celu MVP

* **Engagement:** 50% użytkowników tworzy zapytania co najmniej raz w tygodniu.
* **Retencja:** 70% użytkowników wraca w ciągu 7 dni od rejestracji.
* **Uwaga:** W MVP formalne śledzenie zachowań nie jest wdrożone (pomiar planowany po MVP).

## 7. Integracja z twitterapi.io

### 7.1. Wykorzystywane endpointy

| Endpoint | Cel | Częstotliwość |
|----------|-----|---------------|
| `/twitter/user/info` | Walidacja username X podczas rejestracji | Raz na rejestrację |
| `/twitter/user/followings` | Pobranie listy obserwowanych użytkowników | Co 4h + backfill |
| `/twitter/user/last_tweets` | Pobranie tweetów dla każdego followed user | Co 4h + backfill (150× per ingest) |

### 7.2. Konfiguracja API

* **Klucz API:** Przechowywany w secrets manager lub zmiennych środowiskowych.
* **Header:** `x-api-key: {API_KEY}` dla wszystkich requestów.
* **Base URL:** `https://api.twitterapi.io`
* **Timeout:** 30s dla pojedynczego requestu.
* **Retry policy:** Eksponentialny backoff dla 429 i 5xx (max 3 próby).

### 7.3. Obsługa paginacji

```
cursor = ""
while True:
    response = GET /twitter/user/last_tweets?userName={user}&cursor={cursor}
    process(response.tweets)
    
    if not response.has_next_page:
        break
    
    cursor = response.next_cursor
    
    # Dla regularnego ingestu: break po pierwszej stronie
    # Dla backfillu: kontynuuj do 24h wstecz
```
