# Plan implementacji widoku Historii Zapytań

## 1. Przegląd
Widok Historii Zapytań umożliwia użytkownikowi przeglądanie, zarządzanie i usuwanie swoich poprzednich interakcji Q&A. Prezentuje on listę zapytań z możliwością rozwinięcia szczegółów (pełna odpowiedź i źródła) oraz opcjami usuwania pojedynczych wpisów lub całej historii.

## 2. Routing widoku
- **Ścieżka widoku**: `/history`
- **Dostępność**: Widok chroniony, dostępny tylko dla zalogowanych użytkowników. Wymaga aktywnej sesji.

## 3. Struktura komponentów
```
/src
└── views
    └── HistoryView.tsx
        ├── components
        │   ├── HistoryList.tsx
        │   ├── HistoryItem.tsx
        │   └── DeleteConfirmationDialog.tsx
        └── hooks
            └── useHistory.ts
```

- **`HistoryView`**: Główny kontener widoku, który zarządza pobieraniem danych historii i wyświetla listę.
- **`HistoryList`**: Komponent, który renderuje listę zapytań (`HistoryItem`) oraz przycisk "Usuń wszystko".
- **`HistoryItem`**: Komponent reprezentujący pojedynczy wpis w historii. Używa `Accordion` z Shadcn/ui do pokazywania i ukrywania szczegółów odpowiedzi. Zawiera przycisk do usuwania pojedynczego elementu.
- **`DeleteConfirmationDialog`**: Reużywalny modal (`AlertDialog`) do potwierdzania akcji usuwania (zarówno pojedynczego wpisu, jak i całej historii).
- **`useHistory`**: Niestandardowy hook, który enkapsuluje logikę `react-query` do pobierania, paginacji i usuwania danych historii.

## 4. Szczegóły komponentów

### `HistoryView`
- **Opis**: Pobiera i wyświetla paginowaną listę historii Q&A.
- **Główne elementy**: Nagłówek (`h1`), `HistoryList`.
- **Propsy**: Brak.

### `HistoryList`
- **Opis**: Renderuje listę `HistoryItem` oraz przycisk "Usuń wszystko".
- **Główne elementy**: `div` lub `ul` jako kontener, mapowanie po liście historii i renderowanie `HistoryItem`, `Button` "Usuń wszystko".
- **Propsy**: `items: QAListItemDTO[]`, `onDeleteAll: () => void`.

### `HistoryItem`
- **Opis**: Wyświetla pojedyncze zapytanie. Umożliwia rozwinięcie pełnej odpowiedzi i usunięcie wpisu.
- **Główne elementy**: `AccordionItem`, `AccordionTrigger` (z pytaniem i datą), `AccordionContent` (z pełną odpowiedzią i źródłami), `Button` (ikona kosza).
- **Obsługiwane interakcje**: Rozwijanie/zwijanie szczegółów, kliknięcie przycisku usuwania.
- **Propsy**: `item: QAListItemDTO`, `onDelete: (id: string) => void`.

## 5. Typy

### Typy DTO
```typescript
// GET /api/v1/qa - Response
interface QAListResponseDTO {
  items: QAListItemDTO[];
  next_cursor?: string;
  has_more: boolean;
}

interface QAListItemDTO {
  id: string;
  question: string;
  answer_preview: string;
  created_at: string;
  sources_count: number;
}

// GET /api/v1/qa/{id} - Response
interface QADetailDTO { /* ... */ }

// DELETE /api/v1/qa & /api/v1/qa/{id} - Response
interface MessageResponseDTO {
  message: string;
}
```

## 6. Zarządzanie stanem
Logika zarządzania stanem zostanie zebrana w hooku `useHistory`.

- **`useInfiniteQuery<QAListResponseDTO>`**: Główny hook `react-query` do pobierania paginowanej listy historii. Będzie zarządzał stronami, kursorem (`next_cursor`) i stanem `has_more`.
- **`useQuery<QADetailDTO>`**: Używany wewnątrz `HistoryItem` do dynamicznego pobierania pełnych szczegółów odpowiedzi (`GET /api/v1/qa/{id}`) tylko wtedy, gdy użytkownik rozwinie dany element akordeonu. Opcja `enabled: false` będzie domyślna i aktywowana po kliknięciu.
- **`useMutation` (dla usuwania)**: Dwie mutacje:
  - `useDeleteOne`: Do usuwania pojedynczego wpisu (`DELETE /api/v1/qa/{id}`).
  - `useDeleteAll`: Do usuwania całej historii (`DELETE /api/v1/qa`).
- W `onSuccess` obu mutacji, `queryClient.invalidateQueries(['history'])` zostanie wywołane, aby odświeżyć listę.

## 7. Integracja API
- **Listowanie historii**: `useInfiniteQuery` będzie wywoływać `GET /api/v1/qa` z parametrami `limit` i `cursor`.
- **Pobieranie szczegółów**: `useQuery` w `HistoryItem` będzie wywoływać `GET /api/v1/qa/{id}`.
- **Usuwanie**: Mutacje będą wywoływać odpowiednie endpointy `DELETE`.

## 8. Interakcje użytkownika
- **Przewijanie listy**: Po dojechaniu do końca listy, `useInfiniteQuery` automatycznie pobierze kolejną stronę danych (`fetchNextPage`), jeśli `has_more` jest `true`.
- **Rozwijanie elementu**: Kliknięcie na `AccordionTrigger` aktywuje `useQuery` do pobrania szczegółów i wyświetla je w `AccordionContent`.
- **Usuwanie pojedynczego wpisu**: Kliknięcie ikony kosza otwiera `DeleteConfirmationDialog`. Po potwierdzeniu, wywoływana jest mutacja `useDeleteOne`, a lista jest odświeżana.
- **Usuwanie wszystkiego**: Kliknięcie przycisku "Usuń wszystko" otwiera `DeleteConfirmationDialog`. Po potwierdzeniu, wywoływana jest mutacja `useDeleteAll`, a lista jest czyszczona.

## 9. Warunki i walidacja
- Brak złożonej walidacji po stronie klienta. Główna logika polega na potwierdzaniu akcji destrukcyjnych.

## 10. Obsługa błędów
- **Błędy pobierania listy/szczegółów**: Wyświetlanie ogólnego komunikatu o błędzie w miejscu listy lub w `Toast`.
- **Błędy usuwania**: Wyświetlanie komunikatu o błędzie w `Toast`, np. "Nie udało się usunąć wpisu. Spróbuj ponownie."

## 11. Kroki implementacji
1.  **Struktura plików**: Utwórz pliki `HistoryView.tsx`, `HistoryList.tsx`, `HistoryItem.tsx`, `DeleteConfirmationDialog.tsx` i `useHistory.ts`.
2.  **Hook `useHistory`**: Zaimplementuj `useInfiniteQuery` do paginacji oraz mutacje `useDeleteOne` i `useDeleteAll`.
3.  **Widok `HistoryView`**: Użyj hooka `useHistory` do pobrania danych i przekaż je do `HistoryList`.
4.  **Komponent `HistoryList`**: Zaimplementuj renderowanie listy oraz przycisk "Usuń wszystko" z podpiętą akcją otwierającą dialog potwierdzający.
5.  **Komponent `HistoryItem`**: Zbuduj element listy z `Accordion`. W `AccordionContent` zaimplementuj dynamiczne pobieranie szczegółów (`useQuery`). Podłącz akcję usuwania do przycisku.
6.  **Dialog potwierdzający**: Stwórz generyczny komponent `DeleteConfirmationDialog`, który przyjmuje `onConfirm` jako prop.
7.  **Paginacja "Infinite Scroll"**: Zintegruj `react-intersection-observer` lub podobne narzędzie z `useInfiniteQuery`, aby wywoływać `fetchNextPage` podczas przewijania.
8.  **Routing**: Dodaj chronioną ścieżkę `/history` w routerze aplikacji.
9.  **Styling**: Dopracuj wygląd listy, akordeonu i modali, zapewniając spójność z resztą aplikacji.
