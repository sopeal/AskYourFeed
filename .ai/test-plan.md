**1. Kluczowe komponenty projektu:**

Na podstawie dostarczonego kodu i struktury plików, projekt **AskYourFeed** składa się z następujących kluczowych elementów:

*   **Backend (Go/Gin):**
    *   **Warstwa API (Handlers):** Obsługa żądań HTTP dla autentykacji (`auth_handler`), zarządzania sesją, ingestu danych (`ingest_handler`) oraz funkcjonalności Q&A (`qa_handler`).
    *   **Logika Biznesowa (Services):**
        *   `AuthService`: Rejestracja, logowanie, zarządzanie sesjami JWT/Cookie.
        *   `IngestService`: Kluczowy komponent pobierający dane z Twittera (`twitterapi.io`), obsługujący limity zapytań (rate limiting z exponential backoff) oraz przetwarzający media (OCR/opis obrazów przez OpenRouter).
        *   `QAService`: Logika RAG (Retrieval Augmented Generation) – pobieranie postów z bazy i wysyłanie ich do LLM w celu generowania odpowiedzi.
        *   `LLMService` / `OpenRouterClient`: Integracja z modelami AI.
    *   **Dostęp do Danych (Repositories):** Wykorzystanie `sqlx` do komunikacji z bazą. Implementacja repozytoriów dla Użytkowników, Postów, Autorów, Sesji i Historii Ingestu.
*   **Baza Danych (PostgreSQL):**
    *   Zaawansowane wykorzystanie **Row Level Security (RLS)**. Tabela `posts`, `ingest_runs`, `qa_messages` są izolowane na poziomie użytkownika SQL (`current_setting('app.user_id')`).
    *   Migracje SQL w katalogu `db/`.
*   **Frontend (React/TypeScript):**
    *   Struktura oparta na widokach: Dashboard, History, Login, Register.
    *   Wykorzystanie komponentów UI (Shadcn/Tailwind).
    *   Zarządzanie stanem poprzez hooki (`useAskQuestion`, `useSyncStatus`).
*   **Infrastruktura/DevOps:**
    *   `Makefile` do zarządzania kontenerami Dockerowymi bazy danych.
    *   Docker do konteneryzacji PostgreSQL w środowisku deweloperskim i testowym.

**2. Specyfika stosu technologicznego i wpływ na testowanie:**

*   **Go & Gin:** Statyczne typowanie ułatwia unikanie prostych błędów, ale wymaga solidnych testów jednostkowych dla logiki biznesowej. Wykorzystanie `httptest` w istniejących testach integracyjnych to dobra praktyka, którą należy rozszerzyć.
*   **PostgreSQL RLS:** To krytyczny punkt. Testy integracyjne **muszą** weryfikować, czy ustawienie kontekstu użytkownika (`SET LOCAL app.user_id`) działa poprawnie. Standardowe testy mogą przeoczyć wyciek danych między użytkownikami, jeśli używają jednego konta admina. Należy testować scenariusze "User A próbuje odczytać dane User B".
*   **Zależności Zewnętrzne (Twitter API & OpenRouter):** Aplikacja jest silnie uzależniona od zewnętrznych API. Strategia testowa musi zakładać **mockowanie** tych serwisów w testach CI/CD, aby uniknąć kosztów i fluktuacji (flakiness). Testy E2E powinny działać na środowisku "sandbox" lub nagranych odpowiedziach (VCR/replay).
*   **Asynchroniczność (Goroutines):** Ingest danych dzieje się w tle (`go func()`). Testy muszą uwzględniać mechanizmy synchronizacji (wait groups, kanały lub polling statusu w bazie), aby nie kończyły się przed wykonaniem zadania ("flaky tests").

**3. Priorytety testowe:**

1.  **Bezpieczeństwo i Izolacja Danych (RLS):** Weryfikacja, czy użytkownik widzi tylko swoje posty i historię Q&A. Błąd tutaj to krytyczna luka bezpieczeństwa.
2.  **Stabilność Ingestu (Twitter API):** Obsługa błędów 429 (Rate Limit), wznawianie przerwanego pobierania (`since_id`), obsługa duplikatów postów.
3.  **Logika Q&A i Kontekst LLM:** Sprawdzenie, czy do LLM trafiają odpowiednie posty (filtrowanie po dacie), czy prompt jest poprawnie budowany i czy system radzi sobie z pustym wynikiem wyszukiwania.
4.  **Autentykacja i Sesje:** Poprawność cyklu życia tokena sesji, bezpieczne wylogowanie, haszowanie haseł.

**4. Obszary ryzyka:**

*   **Koszty API:** Błędy w pętli retry w `IngestService` mogą generować ogromne koszty (OpenRouter) lub bany (Twitter API).
*   **Spójność Danych:** Problem z "Backfill" vs "Incremental Update". Czy nie gubimy postów pomiędzy uruchomieniami ingestu?
*   **Obsługa Mediów:** Przetwarzanie wideo i obrazów przez OpenRouter może być wolne i podatne na timeouty. Należy przetestować zachowanie systemu, gdy te usługi odpowiadają z opóźnieniem.
*   **PostgreSQL Connection Pooling:** Przy dużej liczbie zapytań równoległych i użyciu `SET LOCAL` dla RLS, zarządzanie pulą połączeń jest kluczowe, aby sesja jednego użytkownika nie "przeciekła" do zapytania drugiego.

</analiza_projektu>

<plan_testów>

# Plan Testów dla Projektu AskYourFeed

**Wersja dokumentu:** 1.0  
**Data:** 23.05.2025  
**Projekt:** AskYourFeed (MVP)

---

## 1. Wprowadzenie i cele

Celem niniejszego planu jest zdefiniowanie strategii zapewnienia jakości (QA) dla aplikacji **AskYourFeed**. System ten agreguje posty z serwisu X (Twitter), przetwarza je (w tym analizę mediów przy użyciu AI) i umożliwia użytkownikom zadawanie pytań do własnego feedu przy użyciu technologii LLM (RAG).

**Główne cele testowania:**
*   Zapewnienie absolutnej izolacji danych użytkowników (weryfikacja Row Level Security).
*   Gwarancja stabilności procesu pobierania danych (Ingest) w obliczu limitów API zewnętrznych.
*   Weryfikacja poprawności generowania odpowiedzi przez moduł Q&A.
*   Zapewnienie bezbłędnego procesu rejestracji i autentykacji.

## 2. Zakres testów

### W zakresie (In-Scope):
*   **Backend API (Go):** Wszystkie endpointy REST API zdefiniowane w `AskYourFeed.postman_collection.json`.
*   **Baza Danych (PostgreSQL):** Schemat, migracje, polityki bezpieczeństwa (RLS), procedury składowane.
*   **Logika Biznesowa:** Serwisy: Auth, Ingest (w tym retry mechanism), QA, Following.
*   **Integracje:** Obsługa błędów i formatów danych z Twitter API oraz OpenRouter (LLM/Vision).
*   **Frontend (React):** Kluczowe ścieżki użytkownika (rejestracja, dashboard, zadawanie pytań).

### Poza zakresem (Out-of-Scope):
*   Testy wydajnościowe serwerów Twittera/X (nie testujemy zewnętrznego API obciążeniowo).
*   Weryfikacja jakości merytorycznej odpowiedzi modelu LLM (poza podstawową spójnością).
*   Testy kompatybilności na starszych przeglądarkach (wspieramy tylko nowoczesne: Chrome, Firefox, Safari, Edge).

## 3. Typy testów

### 3.1. Testy Jednostkowe (Unit Tests)
*   **Backend:** Testowanie serwisów i helperów w izolacji (mockowanie repozytoriów i klientów HTTP).
    *   Fokus: `IngestService` (logika retry, filtrowanie postów), `AuthService` (walidacja haseł), konwersja DTO.
*   **Frontend:** Testowanie komponentów UI oraz hooków (`useSyncStatus`, `useAskQuestion`) przy użyciu Jest/Vitest.

### 3.2. Testy Integracyjne (Integration Tests)
*   **API + DB:** Testy uruchamiane na rzeczywistej bazie danych (Docker), weryfikujące działanie handlerów, middleware autoryzacji oraz zapytań SQL.
*   **Kluczowy aspekt:** Weryfikacja, czy middleware poprawnie ustawia `app.user_id` i czy RLS w bazie blokuje dostęp do danych innych użytkowników.
*   Wykorzystanie istniejącej struktury w `backend/test/integration`.

### 3.3. Testy Systemowe / E2E (End-to-End)
*   Testowanie pełnych scenariuszy użytkownika od Frontendu do Bazy Danych (z zamockowanymi API zewnętrznymi w celu stabilności).
*   Scenariusz: Rejestracja -> Trigger Ingest -> Zadanie Pytania -> Otrzymanie odpowiedzi -> Weryfikacja Historii.

### 3.4. Testy Bezpieczeństwa (Security)
*   Próby dostępu do zasobów innego użytkownika (IDOR - Insecure Direct Object References) przy włączonym RLS.
*   Weryfikacja wygasania sesji i tokenów.

## 4. Scenariusze testowe

### 4.1. Autentykacja i Autoryzacja
| ID | Scenariusz | Oczekiwany rezultat | Priorytet |
|----|------------|---------------------|-----------|
| AUTH-01 | Rejestracja nowego użytkownika z poprawnymi danymi | Konto utworzone, sesja aktywna, token w cookie/header | Wysoki |
| AUTH-02 | Próba rejestracji na istniejący email | Błąd 409 Conflict | Średni |
| AUTH-03 | Rejestracja z nieistniejącym username X (walidacja API) | Błąd 422 lub odpowiedni komunikat walidacji | Średni |
| AUTH-04 | Dostęp do chronionego endpointu bez tokena | Błąd 401 Unauthorized | Wysoki |
| AUTH-05 | Wylogowanie użytkownika | Token unieważniony, brak dostępu do API | Średni |

### 4.2. Ingest Danych (Pobieranie Postów)
| ID | Scenariusz | Oczekiwany rezultat | Priorytet |
|----|------------|---------------------|-----------|
| ING-01 | Ręczne wyzwolenie ingestu (Trigger) | Status 202, proces startuje w tle (goroutine) | Wysoki |
| ING-02 | Pobieranie statusu ingestu | Poprawny status (running/ok/error), statystyki pobranych postów | Średni |
| ING-03 | Obsługa Rate Limit (429) od Twittera | System czeka (backoff) i ponawia próbę (max 3 razy), status "rate_limited" w przypadku niepowodzenia | Wysoki |
| ING-04 | Próba uruchomienia ingestu gdy jeden już trwa | Błąd 409 Conflict | Średni |
| ING-05 | Przetwarzanie mediów (obrazy/wideo) | Linki do mediów są zastępowane/uzupełniane opisem z OpenRouter w treści posta | Średni |

### 4.3. Q&A (RAG)
| ID | Scenariusz | Oczekiwany rezultat | Priorytet |
|----|------------|---------------------|-----------|
| QA-01 | Zadanie pytania w poprawnym zakresie dat | Otrzymanie odpowiedzi i listy źródeł (postów) | Wysoki |
| QA-02 | Zadanie pytania dla zakresu dat bez postów | Odpowiedź informująca o braku danych, brak błędu 500 | Średni |
| QA-03 | Pobranie historii pytań | Lista pytań posortowana malejąco po dacie, paginacja działa | Niski |
| QA-04 | Usunięcie wpisu z historii Q&A | Wpis znika z listy, dane w bazie oznaczone jako usunięte lub usunięte fizycznie | Niski |

### 4.4. Bezpieczeństwo i RLS
| ID | Scenariusz | Oczekiwany rezultat | Priorytet |
|----|------------|---------------------|-----------|
| SEC-01 | User A próbuje pobrać Q&A Usera B znając UUID | Błąd 404 Not Found (dzięki RLS) | Krytyczny |
| SEC-02 | User A próbuje wyzwolić ingest dla Usera B | Błąd 401/403 lub operacja wykonana dla Usera A (zależnie od implementacji) | Krytyczny |
| SEC-03 | Weryfikacja izolacji w tabeli `posts` | Zapytanie `SELECT * FROM posts` dla zalogowanego Usera A zwraca tylko jego rekordy | Krytyczny |

## 5. Środowisko testowe

*   **Lokalne (Docker):**
    *   Baza danych: PostgreSQL 16 (obraz `postgres:16-alpine`).
    *   Uruchamianie bazy testowej poprzez `make db-start` lub automatycznie w testach Go (`docker_helper.go`).
    *   Zmienne środowiskowe zdefiniowane w pliku `.env.test`.
*   **Mock Server:**
    *   Dla testów integracyjnych i E2E zaleca się użycie mock servera (np. WireMock lub wbudowany w Go `httptest`) symulującego odpowiedzi Twitter API i OpenRouter, aby unikać kosztów i limitów.

## 6. Narzędzia do testowania

*   **Backend:**
    *   `go test` - standardowy runner testów.
    *   `testify` - asercje i mocki (już używane w projekcie).
    *   `sqlx` + `docker` - do testów integracyjnych z bazą danych (setup/teardown).
*   **API Manualne:**
    *   Postman (kolekcja `backend/AskYourFeed.postman_collection.json` jest aktualna i powinna być używana).
*   **Frontend (sugerowane):**
    *   Vitest / Jest - testy jednostkowe.
    *   Playwright - testy E2E.
*   **Statyczna Analiza:**
    *   `golangci-lint` - linter dla Go.
    *   `ESLint` - linter dla TS/React.

## 7. Harmonogram testów

1.  **Testy Ciągłe (CI):** Uruchamiane przy każdym Push/PR.
    *   Linting.
    *   Build aplikacji.
    *   Unit Testy Backend i Frontend.
    *   Testy Integracyjne Backend (z użyciem efemerycznej bazy Docker).
2.  **Testy Regresji:** Przed każdym wydaniem (Release).
    *   Pełen zestaw testów manualnych z kolekcji Postman.
    *   Weryfikacja migracji bazy danych (`make db-reset` -> `make db-init`).
3.  **Testy Bezpieczeństwa:** Raz na sprint.
    *   Audyt reguł RLS w bazie danych.

## 8. Kryteria akceptacji

Produkt uznaje się za gotowy do wydania, gdy:
1.  Wszystkie testy automatyczne (Unit + Integration) przechodzą (Pass rate 100%).
2.  Pokrycie kodu (Code Coverage) dla kluczowych serwisów (`IngestService`, `AuthService`) wynosi min. 80%.
3.  Scenariusze krytyczne bezpieczeństwa (SEC-01, SEC-02, SEC-03) zostały zweryfikowane pozytywnie.
4.  Brak otwartych błędów o priorytecie Krytycznym lub Wysokim.
5.  Migracje bazy danych aplikują się bezbłędnie na czystej instancji oraz na instancji z danymi (upgrade).

## 9. Role i odpowiedzialności

*   **Developer:**
    *   Pisanie testów jednostkowych dla nowego kodu.
    *   Utrzymanie testów integracyjnych (backend).
    *   Uruchamianie testów lokalnie przed commit'em.
*   **QA Engineer:**
    *   Tworzenie i utrzymanie planu testów.
    *   Implementacja testów E2E / Scenariuszy Postman.
    *   Testy manualne i eksploracyjne.
    *   Weryfikacja bezpieczeństwa (RLS).
    *   Zgłaszanie i weryfikacja błędów.

## 10. Procedury raportowania błędów

Błędy należy zgłaszać w systemie śledzenia zadań (np. Jira/GitHub Issues) według szablonu:

*   **Tytuł:** Krótki opis problemu.
*   **Priorytet:** Krytyczny / Wysoki / Średni / Niski.
*   **Środowisko:** (np. Lokalny Docker, Staging).
*   **Kroki do reprodukcji:** Dokładna lista kroków.
*   **Oczekiwany rezultat:** Co powinno się stać.
*   **Rzeczywisty rezultat:** Co się stało (w tym kody błędów HTTP, logi z kontenera).
*   **Załączniki:** Screenshoty, logi z `make db-logs`, payload zapytania.

---
**Zalecenie specjalne dla AskYourFeed:**
Ze względu na użycie `SET LOCAL app.user_id` w `AuthMiddleware`, zaleca się stworzenie dedykowanego zestawu testów integracyjnych w Go, który otwiera **dwa oddzielne połączenia** do bazy danych (symulując dwóch użytkowników) i próbuje wykonywać zapytania krzyżowe (cross-access), aby matematycznie dowieść szczelności RLS.