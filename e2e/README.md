# E2E Tests - Playwright

Ten folder zawiera testy end-to-end (E2E) dla aplikacji AskYourFeed, napisane przy użyciu Playwright.

## Struktura testów

- `register.spec.ts` - Testy rejestracji użytkownika (RegisterView)
- `example.spec.ts` - Przykładowe testy Playwright (do usunięcia w produkcji)

## Wymagania wstępne

Przed uruchomieniem testów upewnij się, że:

1. **Zainstalowane są wszystkie zależności:**
   ```bash
   npm install
   ```

2. **Zainstalowane są przeglądarki Playwright:**
   ```bash
   npx playwright install
   ```

3. **Backend jest uruchomiony:**
   ```bash
   # W katalogu backend/
   go run cmd/main.go
   ```
   Backend powinien działać na domyślnym porcie (zazwyczaj `http://localhost:8080`)

4. **Frontend jest uruchomiony:**
   ```bash
   # W katalogu frontend/
   npm run dev
   ```
   Frontend powinien działać na `http://localhost:5173` (domyślny port Vite)

5. **Baza danych jest uruchomiona i skonfigurowana:**
   - PostgreSQL powinien być uruchomiony
   - Migracje powinny być zastosowane
   - Baza danych powinna być w czystym stanie testowym

## Uruchamianie testów

### Uruchomienie wszystkich testów

```bash
# Z głównego katalogu projektu
npx playwright test
```

### Uruchomienie konkretnego pliku testowego

```bash
# Test rejestracji
npx playwright test e2e/register.spec.ts
```

### Uruchomienie testów w trybie UI (interaktywnym)

```bash
npx playwright test --ui
```

Ten tryb pozwala na:
- Wizualne śledzenie wykonywania testów
- Debugowanie testów krok po kroku
- Podgląd screenshotów i logów

### Uruchomienie testów w trybie headed (z widoczną przeglądarką)

```bash
npx playwright test --headed
```

### Uruchomienie testów w konkretnej przeglądarce

```bash
# Chromium
npx playwright test --project=chromium

# Firefox
npx playwright test --project=firefox

# WebKit (Safari)
npx playwright test --project=webkit
```

### Uruchomienie konkretnego testu

```bash
# Użyj flagi -g (grep) do filtrowania testów po nazwie
npx playwright test -g "should successfully register a new user"
```

### Tryb debug

```bash
npx playwright test --debug
```

Ten tryb:
- Otwiera Playwright Inspector
- Pozwala na wykonywanie testów krok po kroku
- Umożliwia inspekcję elementów strony

## Raporty testów

### Wyświetlenie raportu HTML

Po uruchomieniu testów, raport HTML jest automatycznie generowany:

```bash
npx playwright show-report
```

Raport zawiera:
- Wyniki wszystkich testów
- Screenshoty w przypadku błędów
- Trace'y dla nieudanych testów
- Szczegółowe logi

### Generowanie trace'ów

Trace'y są automatycznie generowane dla pierwszej próby nieudanego testu (konfiguracja: `trace: 'on-first-retry'`).

Aby zawsze generować trace'y:

```bash
npx playwright test --trace on
```

## Konfiguracja testów

Konfiguracja znajduje się w pliku `playwright.config.ts` w głównym katalogu projektu.

### Ważne ustawienia:

- **testDir**: `./e2e` - katalog z testami
- **fullyParallel**: `true` - testy uruchamiane równolegle
- **retries**: `0` lokalnie, `2` na CI
- **reporter**: `html` - raport w formacie HTML

### Dostosowanie BASE_URL

Jeśli frontend działa na innym porcie niż `5173`, zaktualizuj zmienną `BASE_URL` w plikach testowych:

```typescript
const BASE_URL = 'http://localhost:TWOJ_PORT';
```

## Struktura testu rejestracji (register.spec.ts)

Test zawiera następujące scenariusze:

1. ✅ **Wyświetlanie formularza** - sprawdza czy wszystkie pola są widoczne
2. ✅ **Walidacja pustego formularza** - sprawdza błędy walidacji
3. ✅ **Niezgodne hasła** - sprawdza walidację zgodności haseł
4. ✅ **Pomyślna rejestracja** - testuje pełny flow rejestracji
5. ✅ **Duplikat email** - sprawdza obsługę już zarejestrowanego emaila
6. ✅ **Nieistniejący użytkownik X** - sprawdza walidację nazwy użytkownika X
7. ✅ **Nawigacja do logowania** - sprawdza link do strony logowania
8. ✅ **Obsługa inputów** - sprawdza poprawność wprowadzania danych
9. ✅ **Dostępność (a11y)** - sprawdza atrybuty dostępności

## Przygotowanie środowiska testowego

### Opcja 1: Ręczne przygotowanie

1. Uruchom backend:
   ```bash
   cd backend
   go run cmd/main.go
   ```

2. W nowym terminalu uruchom frontend:
   ```bash
   cd frontend
   npm run dev
   ```

3. W trzecim terminalu uruchom testy:
   ```bash
   npx playwright test
   ```

### Opcja 2: Użycie Makefile (jeśli dostępne)

```bash
# Sprawdź dostępne komendy
make help

# Uruchom wszystko jednocześnie (jeśli skonfigurowane)
make test-e2e
```

## Debugowanie testów

### 1. Użyj Playwright Inspector

```bash
npx playwright test --debug
```

### 2. Dodaj breakpointy w kodzie

```typescript
await page.pause(); // Zatrzyma wykonanie testu
```

### 3. Włącz verbose logging

```bash
DEBUG=pw:api npx playwright test
```

### 4. Zapisz screenshoty

```typescript
await page.screenshot({ path: 'screenshot.png' });
```

### 5. Zapisz trace

```bash
npx playwright test --trace on
```

## Czyszczenie danych testowych

Po testach rejestracji mogą pozostać dane testowe w bazie. Aby je wyczyścić:

```sql
-- Usuń użytkowników testowych
DELETE FROM users WHERE email LIKE 'test%@example.com';
```

Lub użyj skryptu czyszczącego (jeśli dostępny):

```bash
make clean-test-data
```

## Najlepsze praktyki

1. **Izolacja testów** - każdy test powinien być niezależny
2. **Unikalne dane** - używaj timestamp'ów do generowania unikalnych danych
3. **Czekanie na elementy** - używaj `await expect().toBeVisible()` zamiast `waitFor()`
4. **Selektory semantyczne** - preferuj `getByRole`, `getByLabel` nad `locator()`
5. **Cleanup** - upewnij się, że testy nie zostawiają śmieci w bazie danych

## Rozwiązywanie problemów

### Problem: Testy nie mogą połączyć się z frontendem

**Rozwiązanie:**
- Sprawdź czy frontend działa: `curl http://localhost:5173`
- Sprawdź czy port jest poprawny w `BASE_URL`
- Upewnij się, że nie ma konfliktów portów

### Problem: Testy timeout'ują podczas rejestracji

**Rozwiązanie:**
- Sprawdź czy backend działa i odpowiada
- Sprawdź logi backendu pod kątem błędów
- Zwiększ timeout w teście: `{ timeout: 30000 }`

### Problem: Błąd "X username not found"

**Rozwiązanie:**
- Test używa nieistniejącej nazwy użytkownika X - to oczekiwane zachowanie
- Jeśli chcesz test z prawdziwym użytkownikiem, użyj istniejącej nazwy

### Problem: Przeglądarki nie są zainstalowane

**Rozwiązanie:**
```bash
npx playwright install
```

## CI/CD

Testy są skonfigurowane do uruchamiania na CI z następującymi ustawieniami:
- Retry: 2 próby dla nieudanych testów
- Workers: 1 (sekwencyjne wykonanie)
- Headless mode: zawsze

## Dodatkowe zasoby

- [Dokumentacja Playwright](https://playwright.dev/)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Guide](https://playwright.dev/docs/debug)
- [Selectors Guide](https://playwright.dev/docs/selectors)

## Kontakt

W przypadku problemów z testami, sprawdź:
1. Logi testów: `npx playwright show-report`
2. Logi backendu
3. Logi frontendu (konsola przeglądarki)
