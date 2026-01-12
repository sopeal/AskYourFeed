# Plan implementacji widoku Logowanie

## 1. Przegląd
Widok Logowania umożliwia powracającym użytkownikom uwierzytelnienie się w aplikacji "Ask Your Feed" za pomocą adresu e-mail i hasła. Po pomyślnym zalogowaniu, użytkownik jest przekierowywany do głównego panelu aplikacji, a jego sesja jest bezpiecznie zarządzana.

## 2. Routing widoku
- **Ścieżka widoku**: `/login`
- **Dostępność**: Widok publiczny, domyślny dla niezalogowanych użytkowników. Zalogowany użytkownik, który spróbuje uzyskać dostęp do tej ścieżki, powinien zostać automatycznie przekierowany do panelu głównego (`/`).

## 3. Struktura komponentów
Struktura komponentów będzie bardzo podobna do widoku Rejestracji, promując reużywalność.

```
/src
└── views
    └── LoginView.tsx
        └── components
            └── LoginForm.tsx
                ├── shared
                │   ├── InputField.tsx
                │   └── SubmitButton.tsx
                └── (komponenty z biblioteki Shadcn/ui)
```

- **`LoginView`**: Główny kontener strony, odpowiedzialny za wyświetlenie i centrowanie formularza logowania.
- **`LoginForm`**: Komponent formularza, który zarządza stanem, walidacją i komunikacją z API logowania.

## 4. Szczegóły komponentów

### `LoginView`
- **Opis komponentu**: Strona logowania, która prezentuje formularz logowania oraz link do strony rejestracji.
- **Główne elementy**: Główny kontener (`div`), nagłówek (`h1` z tekstem "Zaloguj się"), komponent `LoginForm` oraz link (`<Link to="/register">`) dla nowych użytkowników.
- **Obsługiwane interakcje**: Nawigacja do strony rejestracji.
- **Typy**: Brak.
- **Propsy**: Brak.

### `LoginForm`
- **Opis komponentu**: Formularz do zbierania danych logowania (e-mail i hasło) i wysyłania ich do API.
- **Główne elementy**:
  - `form`
  - `InputField` dla adresu e-mail.
  - `InputField` (typu `password`) dla hasła.
  - `SubmitButton` do wysłania formularza.
- **Obsługiwane interakcje**:
  - Wprowadzanie danych.
  - Wysłanie formularza.
- **Obsługiwana walidacja**: Walidacja po stronie klienta (poprawność formatu e-mail, niepuste pola) oraz obsługa błędów z API (np. "Nieprawidłowy e-mail lub hasło").
- **Typy**: `LoginFormViewModel`, `LoginFormValidation`.
- **Propsy**: Brak.

## 5. Typy

### Typy DTO
Zgodne z `dto.go`.

```typescript
// POST /api/v1/auth/login - Request Body
interface LoginCommand {
  email: string;
  password: string;
}

// POST /api/v1/auth/login - Success Response
interface LoginResponseDTO {
  user_id: string; // UUID
  email: string;
  x_username: string;
  x_display_name: string;
  session_token: string;
  session_expires_at: string; // ISO 8601 Date
}
```

### Typy ViewModel

```typescript
// Stan formularza
interface LoginFormViewModel {
  email: string;
  password: string;
}

// Stan błędów walidacji
type LoginFormValidation = {
  [key in keyof LoginFormViewModel]?: string;
};
```

## 6. Zarządzanie stanem
Podobnie jak w przypadku rejestracji, stan formularza będzie zarządzany przez `react-hook-form` i `zod`, a komunikacja z API przez `react-query`.

- **`useForm<LoginFormViewModel>`**: Hook do zarządzania stanem formularza logowania.
- **`useMutation`**: Hook `useLogin` do obsługi mutacji `POST /api/v1/auth/login`.
- **`AuthContext`**: Po pomyślnym logowaniu, dane sesji zostaną zaktualizowane w tym globalnym kontekście.

## 7. Integracja API
- **Endpoint**: `POST /api/v1/auth/login`
- **Akcja**: Wywołanie mutacji `useLogin` po walidacji i wysłaniu formularza.
- **Request**:
  - **Typ**: `LoginCommand`
- **Response (Success)**:
  - **Typ**: `LoginResponseDTO`
  - **Obsługa**: W `onSuccess` mutacji:
    1. Zapisz `session_token` i dane użytkownika w `AuthContext`.
    2. Przekieruj użytkownika na stronę główną (`/`).
- **Response (Error)**:
  - **Typ**: `ErrorResponseDTO`
  - **Obsługa**: W `onError` mutacji, na podstawie kodu błędu (np. 401), wyświetl ogólny komunikat "Nieprawidłowy e-mail lub hasło" za pomocą `Toast`.

## 8. Interakcje użytkownika
- **Wypełnianie formularza**: Użytkownik wpisuje e-mail i hasło.
- **Kliknięcie "Zaloguj się"**: Uruchamia walidację i wysłanie danych. Przycisk przechodzi w stan ładowania.
- **Pomyślne logowanie**: Użytkownik jest przekierowywany do panelu głównego.
- **Błąd logowania**: Użytkownik widzi komunikat o błędzie.

## 9. Warunki i walidacja
- **Email**: Musi być poprawnym adresem e-mail, pole wymagane.
- **Hasło**: Pole wymagane.

## 10. Obsługa błędów
- **Błędy walidacji klienta**: Wyświetlane pod polami formularza (np. "To pole jest wymagane").
- **Błędy API**:
  - **`401 Unauthorized`**: Komunikat `Toast`: "Nieprawidłowy adres e-mail lub hasło."
  - **`500 Internal Server Error`**: Komunikat `Toast`: "Wystąpił błąd serwera. Spróbuj ponownie później."

## 11. Kroki implementacji
1.  **Struktura plików**: Utwórz pliki `LoginView.tsx` i `LoginForm.tsx`.
2.  **Definicja typów**: Zdefiniuj typy `LoginCommand`, `LoginResponseDTO` i `LoginFormViewModel`.
3.  **Budowa formularza**: Zaimplementuj `LoginForm` przy użyciu `react-hook-form` i `zod`.
4.  **Integracja API**: Stwórz hook `useLogin` (`useMutation`) do komunikacji z endpointem logowania.
5.  **Obsługa stanu**: Zintegruj stany `isLoading`, `isSuccess`, `isError` z interfejsem użytkownika (przycisk, toasty).
6.  **Kontekst autoryzacji**: Po sukcesie, zaktualizuj `AuthContext` i przekieruj użytkownika.
7.  **Routing**: Dodaj ścieżkę `/login` w głównym routerze aplikacji.
8.  **Styling**: Dopracuj wygląd i responsywność widoku.
