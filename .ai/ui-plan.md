# Architektura UI dla Ask Your Feed

## 1. Przegląd struktury UI

Architektura interfejsu użytkownika (UI) dla aplikacji "Ask Your Feed" zostanie zrealizowana jako Single Page Application (SPA) w oparciu o bibliotekę **React** z użyciem **TypeScript**. Do zarządzania routingiem posłuży **React Router**, a za warstwę wizualną i komponenty odpowiadać będzie biblioteka **Shadcn/ui** w połączeniu z **Tailwind CSS**, co zapewni spójność, responsywność oraz wsparcie dla motywów (dark/light mode) i wysoką dostępność (a11y).

Zarządzanie stanem serwera (pobieranie danych, mutacje, buforowanie) zostanie powierzone bibliotece **React Query (TanStack Query)**. Globalny, rzadko zmieniający się stan aplikacji, taki jak informacje o sesji użytkownika i status synchronizacji danych, będzie zarządzany przez **React Context**.

Aplikacja będzie składać się z publicznych widoków (logowanie, rejestracja) oraz widoków chronionych, dostępnych po uwierzytelnieniu (główny panel Q&A, historia zapytań).

## 2. Lista widoków

### Widok Rejestracji
- **Nazwa widoku:** Rejestracja
- **Ścieżka widoku:** `/register`
- **Główny cel:** Umożliwienie nowemu użytkownikowi stworzenia konta poprzez podanie adresu e-mail, hasła oraz nazwy użytkownika na portalu X.
- **Kluczowe informacje do wyświetlenia:** Formularz z polami na e-mail, hasło, potwierdzenie hasła i nazwę użytkownika X.
- **Kluczowe komponenty widoku:** `Card`, `Form`, `Input`, `Button`, `Toast` (do informacji zwrotnych).
- **UX, dostępność i względy bezpieczeństwa:**
    - **UX:** Asynchroniczna walidacja nazwy użytkownika X (`onBlur`) z wyraźnym wskaźnikiem ładowania, aby dać natychmiastową informację zwrotną. Po pomyślnej rejestracji użytkownik jest automatycznie logowany i przekierowywany do panelu głównego.
    - **Dostępność:** Poprawne etykiety dla pól formularza, obsługa walidacji i błędów zgodna z ARIA.
    - **Bezpieczeństwo:** Wymuszenie polityki skomplikowania hasła po stronie klienta, komunikacja z API przez HTTPS.

### Widok Logowania
- **Nazwa widoku:** Logowanie
- **Ścieżka widoku:** `/login`
- **Główny cel:** Uwierzytelnienie istniejącego użytkownika.
- **Kluczowe informacje do wyświetlenia:** Formularz z polami na e-mail i hasło.
- **Kluczowe komponenty widoku:** `Card`, `Form`, `Input`, `Button`, `Toast`.
- **UX, dostępność i względy bezpieczeństwa:**
    - **UX:** Czytelne komunikaty o błędach (np. "Nieprawidłowy e-mail lub hasło"), link do strony rejestracji.
    - **Dostępność:** Zapewnienie, że komunikaty o błędach są powiązane z odpowiednimi polami formularza.
    - **Bezpieczeństwo:** Ochrona przed atakami typu brute-force (obsługiwana przez rate-limiting na API), bezpieczne przesyłanie danych.

### Widok Główny (Dashboard)
- **Nazwa widoku:** Panel Główny
- **Ścieżka widoku:** `/`
- **Główny cel:** Umożliwienie użytkownikowi zadawania pytań do swojego feedu i przeglądania odpowiedzi wygenerowanych przez LLM.
- **Kluczowe informacje do wyświetlenia:**
    - Pole do wprowadzania pytania.
    - Selektor zakresu dat.
    - Obszar wyświetlania odpowiedzi (lista punktów).
    - Sekcja "Źródła" z listą postów użytych do wygenerowania odpowiedzi.
    - Informacje o stanie synchronizacji (w nagłówku/stopce).
- **Kluczowe komponenty widoku:** `Textarea`, `Button`, `DateRangePicker` (desktop), `Select` (mobile), `Skeleton` (dla stanu ładowania), `Card` (dla odpowiedzi i źródeł).
- **UX, dostępność i względy bezpieczeństwa:**
    - **UX:** Domyślny zakres dat ustawiony na 24h. Interfejs wejściowy jest blokowany podczas generowania odpowiedzi, a stan ładowania jest komunikowany za pomocą animowanego szkieletu komponentu. Wyraźny komunikat w przypadku braku treści w danym zakresie dat.
    - **Dostępność:** Zapewnienie, że stan ładowania jest anonsowany przez czytniki ekranu. Odpowiedź i źródła mają logiczną strukturę nagłówków.
    - **Bezpieczeństwo:** Sanityzacja treści wprowadzanej przez użytkownika po stronie klienta (dodatkowa walidacja oprócz serwerowej).

### Widok Historii
- **Nazwa widoku:** Historia Zapytań
- **Ścieżka widoku:** `/history`
- **Główny cel:** Przeglądanie, zarządzanie i usuwanie poprzednich zapytań i odpowiedzi.
- **Kluczowe informacje do wyświetlenia:** Lista poprzednich zapytań z datą i skrótem odpowiedzi.
- **Kluczowe komponenty widoku:** `Accordion` (do rozwijania szczegółów), `Button` (do usuwania), `AlertDialog` (do potwierdzania usunięcia), `Pagination`.
- **UX, dostępność i względy bezpieczeństwa:**
    - **UX:** Szczegóły odpowiedzi (pełna treść i źródła) są ładowane dynamicznie po rozwinięciu elementu listy, co optymalizuje początkowe ładowanie strony. Akcje usuwania wymagają potwierdzenia, aby zapobiec przypadkowej utracie danych.
    - **Dostępność:** Elementy akordeonu są w pełni dostępne z klawiatury, a akcje usuwania mają odpowiednie etykiety ARIA.
    - **Bezpieczeństwo:** Wszystkie operacje usuwania są autoryzowane i weryfikowane po stronie serwera.

## 3. Mapa podróży użytkownika

**Główny przepływ: Od rejestracji do zadania pierwszego pytania**

1.  **Lądowanie i Rejestracja:** Nowy użytkownik trafia na stronę logowania (`/login`) i klika link do rejestracji, przechodząc do widoku `/register`.
2.  **Wypełnienie formularza:** Użytkownik podaje e-mail, hasło i swoją nazwę użytkownika X. Nazwa użytkownika jest walidowana asynchronicznie po opuszczeniu pola (`onBlur`).
3.  **Automatyczne logowanie i Onboarding:** Po pomyślnej rejestracji, system automatycznie loguje użytkownika i przekierowuje go na panel główny (`/`). Wyświetlony zostaje komunikat `Toast` informujący o rozpoczęciu pierwszej synchronizacji danych (backfill 24h).
4.  **Oczekiwanie na dane:** Wskaźnik w nagłówku informuje o trwającym procesie synchronizacji.
5.  **Zadanie pytania:** Po zakończeniu synchronizacji, użytkownik wpisuje pytanie w polu tekstowym. Domyślny zakres dat to ostatnie 24 godziny.
6.  **Generowanie odpowiedzi:** Po kliknięciu "Zapytaj", pole do wprowadzania pytania i przycisk zostają zablokowane, a w miejscu odpowiedzi pojawia się animowany placeholder (`Skeleton`).
7.  **Otrzymanie odpowiedzi:** Po otrzymaniu danych z API, wyświetlana jest odpowiedź w formie listy punktów oraz sekcja "Źródła" zawierająca klikalne karty z linkami do oryginalnych postów.
8.  **Weryfikacja źródeł:** Użytkownik może kliknąć na dowolne źródło, co otworzy oryginalny post na portalu X w nowej karcie przeglądarki.
9.  **Przeglądanie historii:** Użytkownik przechodzi do widoku `/history`, gdzie jego ostatnie zapytanie jest widoczne na górze listy.
10. **Wylogowanie:** Użytkownik klika na swoje awatar/menu i wybiera opcję "Wyloguj", co kończy sesję i przekierowuje go z powrotem na stronę logowania.

## 4. Układ i struktura nawigacji

Nawigacja w aplikacji jest prosta i scentralizowana w stałym nagłówku, zapewniając spójność i łatwy dostęp do kluczowych sekcji.

- **Nawigacja główna:**
    - Umieszczona w nagłówku aplikacji.
    - Zawiera linki do:
        - **Panelu Głównego (`/`)**
        - **Historii (`/history`)**
- **Menu użytkownika:**
    - Zintegrowane z awatarem użytkownika w nagłówku.
    - Zawiera opcję **"Wyloguj"**.
- **Wskaźnik statusu synchronizacji:**
    - Również znajduje się w nagłówku, zawsze widoczny.
    - Wyświetla czas ostatniej synchronizacji oraz liczbę obserwowanych kont (np. "150/150").
    - Używa ikon do sygnalizowania stanu: "w toku" (animowana ikona) lub "błąd" (ikona ostrzeżenia z `Tooltip`em wyjaśniającym problem).
- **Routing:**
    - Realizowany przez `React Router`.
    - Widoki `/` i `/history` są chronione i wymagają uwierzytelnienia. Niezalogowany użytkownik próbujący uzyskać do nich dostęp jest przekierowywany na `/login`.
    - Zalogowany użytkownik próbujący wejść na `/login` lub `/register` jest automatycznie przekierowywany na `/`.

## 5. Kluczowe komponenty

Poniżej znajduje się lista kluczowych, reużywalnych komponentów, które będą stanowić fundament interfejsu użytkownika.

- **`PageLayout`**: Komponent-wrapper dla każdej strony, zawierający stały nagłówek (`Header`) i stopkę, zapewniający spójną strukturę i layout na wszystkich widokach.
- **`Header`**: Komponent nagłówka zawierający nawigację główną, menu użytkownika oraz komponent `SyncStatusIndicator`.
- **`SyncStatusIndicator`**: Mały, dedykowany komponent w nagłówku, który pobiera dane o stanie synchronizacji z globalnego kontekstu i wyświetla je w czasie rzeczywistym.
- **`SourceCard`**: Karta używana w sekcji "Źródła" do wyświetlania pojedynczego postu. Zawiera nazwę autora, podgląd tekstu, datę publikacji i jest w całości klikalnym linkiem do oryginalnego posta.
- **`ConfirmationDialog`**: Modal (`AlertDialog` z Shadcn/ui) używany do uzyskania od użytkownika potwierdzenia przed wykonaniem destrukcyjnej akcji, takiej jak usunięcie pojedynczego wpisu z historii lub całej historii.
- **`ErrorToast`**: Globalny komponent `Toast`, który subskrybuje stan błędów (np. z React Query) i wyświetla komunikaty o błędach API w sposób nieinwazyjny.
- **`QAResponse`**: Komponent do renderowania odpowiedzi z LLM, który wewnętrznie obsługuje wyświetlanie listy punktów oraz renderuje listę komponentów `SourceCard`.
