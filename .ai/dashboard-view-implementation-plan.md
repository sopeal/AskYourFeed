# Plan implementacji widoku Panelu Głównego (Dashboard)

## 1. Przegląd
Panel Główny jest centralnym punktem aplikacji, gdzie użytkownik może zadawać pytania dotyczące swojego feedu z serwisu X. Widok ten umożliwia wprowadzanie zapytań, filtrowanie ich według zakresu dat, a następnie prezentuje odpowiedzi wygenerowane przez LLM wraz z listą źródeł. Dodatkowo, informuje użytkownika o stanie synchronizacji danych.

## 2. Routing widoku
- **Ścieżka widoku**: `/`
- **Dostępność**: Widok chroniony, dostępny tylko dla zalogowanych użytkowników. Próba dostępu bez aktywnej sesji powinna skutkować przekierowaniem na stronę logowania (`/login`).

## 3. Struktura komponentów
```
/src
└── views
    └── DashboardView.tsx
        ├── components
        │   ├── QAForm.tsx
        │   ├── QAResponse.tsx
        │   └── SourceCard.tsx
        └── hooks
            └── useSyncStatus.ts
```

- **`DashboardView`**: Główny kontener widoku, który zarządza layoutem i integruje pozostałe komponenty.
- **`QAForm`**: Formularz do zadawania pytań, zawierający pole tekstowe (`Textarea`) i selektor zakresu dat (`DateRangePicker`).
- **`QAResponse`**: Komponent do wyświetlania odpowiedzi z LLM. Renderuje listę punktów oraz listę źródeł (`SourceCard`). Wyświetla również stan ładowania (`Skeleton`) lub komunikat o braku danych.
- **`SourceCard`**: Karta prezentująca pojedyncze źródło (post z X), zawierająca autora, treść i link do oryginału.
- **`useSyncStatus`**: Niestandardowy hook, który w regularnych odstępach czasu odpytuje API o status synchronizacji (`/api/v1/ingest/status`) i dostarcza te dane do wyświetlenia w UI (np. w nagłówku aplikacji).

## 4. Szczegóły komponentów

### `DashboardView`
- **Opis**: Integruje formularz Q&A i obszar odpowiedzi. Zarządza stanem zapytania i odpowiedzi.
- **Główne elementy**: `QAForm`, `QAResponse`.
- **Propsy**: Brak.

### `QAForm`
- **Opis**: Umożliwia użytkownikowi wprowadzenie pytania i wybranie zakresu dat.
- **Główne elementy**: `form`, `Textarea`, `DateRangePicker` (lub `Select` z predefiniowanymi zakresami na mobile), `Button`.
- **Obsługiwane interakcje**: Wprowadzanie tekstu, zmiana zakresu dat, wysłanie formularza.
- **Propsy**: `onSubmit(question: string, dateFrom?: Date, dateTo?: Date)`, `isLoading: boolean`.

### `QAResponse`
- **Opis**: Wyświetla wynik zapytania Q&A.
- **Główne elementy**: `Card` do opakowania odpowiedzi, `ul`/`li` dla punktów odpowiedzi, lista komponentów `SourceCard`.
- **Warunki**: Renderuje `Skeleton` gdy `isLoading` jest `true`. Renderuje komunikat "Brak treści w wybranym zakresie" gdy odpowiedź jest pusta. Renderuje odpowiedź i źródła w przypadku sukcesu.
- **Propsy**: `data: QADetailDTO | null`, `isLoading: boolean`.

## 5. Typy

### Typy DTO
```typescript
// POST /api/v1/qa - Request
interface CreateQACommand {
  question: string;
  date_from?: string; // ISO 8601
  date_to?: string;   // ISO 8601
}

// POST /api/v1/qa - Response
interface QADetailDTO {
  id: string;
  question: string;
  answer: string;
  date_from: string;
  date_to: string;
  created_at: string;
  sources: QASourceDTO[];
}

interface QASourceDTO {
  x_post_id: number;
  author_handle: string;
  author_display_name: string;
  published_at: string;
  url: string;
  text_preview: string;
}

// GET /api/v1/ingest/status - Response
interface IngestStatusDTO { /* ...zgodnie z dto.go... */ }
```

### Typy ViewModel
```typescript
interface QAFormViewModel {
  question: string;
  dateRange: { from?: Date; to?: Date };
}
```

## 6. Zarządzanie stanem
- **`useForm<QAFormViewModel>`**: Do zarządzania stanem formularza `QAForm`.
- **`useMutation<QADetailDTO, Error, CreateQACommand>`**: Hook `useAskQuestion` do wysyłania zapytania do `POST /api/v1/qa`. Jego stan `isLoading` i `data` będą przekazywane do komponentów `QAForm` i `QAResponse`.
- **`useQuery<IngestStatusDTO>`**: Hook `useSyncStatus` będzie używał `useQuery` do okresowego pobierania danych z `GET /api/v1/ingest/status` z opcją `refetchInterval`.
- **`useState<QADetailDTO | null>`**: Lokalny stan w `DashboardView` do przechowywania ostatniej pomyślnej odpowiedzi, aby nie znikała podczas kolejnego zapytania.

## 7. Integracja API
- **Pytanie Q&A**:
  - **Endpoint**: `POST /api/v1/qa`
  - **Akcja**: Wywołanie mutacji `useAskQuestion` z danymi z `QAForm`.
  - **Obsługa**: Wynik mutacji (dane, stan ładowania, błąd) jest przekazywany do `QAResponse`.
- **Status synchronizacji**:
  - **Endpoint**: `GET /api/v1/ingest/status`
  - **Akcja**: Hook `useSyncStatus` będzie cyklicznie (`refetchInterval: 30000`) odpytywał ten endpoint.
  - **Obsługa**: Dane będą dostępne globalnie (np. przez kontekst) i wyświetlane w komponencie `SyncStatusIndicator` w nagłówku aplikacji.

## 8. Interakcje użytkownika
- **Zadawanie pytania**: Użytkownik wpisuje pytanie, opcjonalnie zmienia datę i klika "Zapytaj". Formularz jest blokowany, a `QAResponse` pokazuje `Skeleton`.
- **Otrzymanie odpowiedzi**: `QAResponse` renderuje odpowiedź i źródła.
- **Kliknięcie w źródło**: Otwiera oryginalny post w nowej karcie przeglądarki.

## 9. Warunki i walidacja
- **Pytanie**: Pole wymagane, niepuste. Walidacja po stronie klienta przed wysłaniem.
- **Zakres dat**: `date_from` musi być wcześniejsza lub równa `date_to`. Walidacja w komponencie `DateRangePicker`.

## 10. Obsługa błędów
- **Brak treści**: Jeśli API zwróci pustą odpowiedź (`sources: []`), `QAResponse` wyświetli komunikat "Brak treści w wybranym zakresie. Spróbuj rozszerzyć zakres dat."
- **Błędy API (`POST /api/v1/qa`)**:
  - **`429 Too Many Requests`**: `Toast` z komunikatem "Przekroczono limit zapytań. Spróbuj ponownie później."
  - **`5xx`**: `Toast` z komunikatem "Wystąpił błąd serwera podczas generowania odpowiedzi."
- **Błędy API (`GET /api/v1/ingest/status`)**: Błędy będą dyskretnie obsługiwane w `SyncStatusIndicator`, np. przez zmianę ikony na ikonę błędu z tooltipem.

## 11. Kroki implementacji
1.  **Struktura plików**: Utwórz pliki `DashboardView.tsx`, `QAForm.tsx`, `QAResponse.tsx`, `SourceCard.tsx` oraz hook `useSyncStatus.ts`.
2.  **Komponent `QAForm`**: Zbuduj formularz z `Textarea` i `DateRangePicker`, używając `react-hook-form`.
3.  **Komponent `QAResponse`**: Zaimplementuj logikę warunkowego renderowania dla stanów ładowania, błędu, braku danych i sukcesu.
4.  **Komponent `SourceCard`**: Stwórz reużywalną kartę dla pojedynczego źródła.
5.  **Hook `useAskQuestion`**: Zaimplementuj mutację `react-query` do wysyłania pytań.
6.  **Widok `DashboardView`**: Zintegruj `QAForm` i `QAResponse`, przekazując między nimi stan (dane, `isLoading`).
7.  **Hook `useSyncStatus`**: Zaimplementuj `useQuery` z `refetchInterval` do pobierania statusu synchronizacji.
8.  **Integracja `SyncStatusIndicator`**: Stwórz komponent w nagłówku, który będzie konsumował dane z `useSyncStatus` (prawdopodobnie przez kontekst).
9.  **Routing**: Upewnij się, że ścieżka `/` jest chroniona i renderuje `DashboardView`.
10. **Styling**: Dopracuj wygląd, responsywność i dostępność wszystkich komponentów.
