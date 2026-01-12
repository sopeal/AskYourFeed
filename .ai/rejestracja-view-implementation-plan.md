# Plan implementacji widoku Rejestracja

## 1. Przegląd
Widok Rejestracji umożliwia nowym użytkownikom utworzenie konta w aplikacji "Ask Your Feed". Użytkownik podaje swój adres e-mail, hasło oraz nazwę użytkownika z serwisu X (dawniej Twitter). System weryfikuje poprawność danych, w tym istnienie konta X, a po pomyślnej rejestracji automatycznie loguje użytkownika i przekierowuje go do głównego panelu aplikacji.

## 2. Routing widoku
- **Ścieżka widoku**: `/register`
- **Dostępność**: Widok publiczny, dostępny dla niezalogowanych użytkowników. Zalogowany użytkownik próbujący uzyskać dostęp do tej ścieżki powinien zostać przekierowany do panelu głównego (`/`).

## 3. Struktura komponentów
Hierarchia komponentów dla widoku Rejestracji będzie następująca:

```
/src
└── views
    └── RegisterView.tsx
        └── components
            └── RegisterForm.tsx
                ├── shared
                │   ├── InputField.tsx
                │   └── SubmitButton.tsx
                └── (komponenty z biblioteki Shadcn/ui)
```

- **`RegisterView`**: Główny kontener strony, odpowiedzialny za layout i otoczenie formularza.
- **`RegisterForm`**: Komponent zawierający logikę formularza, zarządzanie jego stanem, walidację i obsługę wysyłania danych do API.
- **`InputField`**: Reużywalny komponent pola tekstowego (bazujący na `Input` z Shadcn/ui) z obsługą etykiet, walidacji i wyświetlania błędów.
- **`SubmitButton`**: Reużywalny przycisk (bazujący na `Button` z Shadcn/ui) z obsługą stanu ładowania.

## 4. Szczegóły komponentów

### `RegisterView`
- **Opis komponentu**: Strona rejestracji, która centruje i wyświetla formularz rejestracyjny.
- **Główne elementy**: Główny kontener (`div`), nagłówek (`h1` z tekstem "Stwórz konto"), oraz komponent `RegisterForm`.
- **Obsługiwane interakcje**: Brak bezpośrednich interakcji, deleguje wszystko do `RegisterForm`.
- **Typy**: Brak.
- **Propsy**: Brak.

### `RegisterForm`
- **Opis komponentu**: Sercem widoku jest formularz, który zbiera dane od użytkownika, waliduje je i wysyła do API w celu utworzenia konta.
- **Główne elementy**:
  - `form`
  - `InputField` dla adresu e-mail.
  - `InputField` (typu `password`) dla hasła.
  - `InputField` (typu `password`) dla potwierdzenia hasła.
  - `InputField` dla nazwy użytkownika X.
  - `SubmitButton` do wysłania formularza.
- **Obsługiwane interakcje**:
  - Wprowadzanie danych w polach formularza.
  - Wysłanie formularza.
- **Obsługiwana walidacja**:
  - Walidacja po stronie klienta przed wysłaniem (szczegóły w sekcji 9).
  - Wyświetlanie błędów walidacji zwróconych przez API (np. zajęty e-mail, nieistniejący użytkownik X).
- **Typy**: `RegisterFormViewModel`, `RegisterFormValidation`.
- **Propsy**: Brak.

## 5. Typy

### Typy DTO (Data Transfer Objects)
Typy te odzwierciedlają strukturę danych wymienianą z API i bazują na pliku `dto.go`.

```typescript
// POST /api/v1/auth/register - Request Body
interface RegisterCommand {
  email: string;
  password: string;
  password_confirmation: string;
  x_username: string;
}

// POST /api/v1/auth/register - Success Response
interface RegisterResponseDTO {
  user_id: string; // UUID
  email: string;
  x_username: string;
  x_display_name: string;
  created_at: string; // ISO 8601 Date
  session_token: string;
}

// API Error Response
interface ErrorDetailDTO {
  code: string;
  message: string;
  details?: Record<string, any>;
}

interface ErrorResponseDTO {
  error: ErrorDetailDTO;
}
```

### Typy ViewModel
Typy używane wewnętrznie przez komponenty do zarządzania stanem formularza.

```typescript
// Stan formularza
interface RegisterFormViewModel {
  email: string;
  password: string;
  passwordConfirmation: string;
  xUsername: string;
}

// Stan błędów walidacji
type RegisterFormValidation = {
  [key in keyof RegisterFormViewModel]?: string;
};
```

## 6. Zarządzanie stanem
Stan formularza będzie zarządzany przy użyciu biblioteki `react-hook-form`, która upraszcza obsługę pól, walidację i proces wysyłania danych. Do walidacji schematu danych zostanie użyta biblioteka `zod`.

- **`useForm<RegisterFormViewModel>`**: Hook z `react-hook-form` do zarządzania stanem formularza, błędami i statusem wysyłania.
- **`useMutation`**: Hook z `react-query` (`@tanstack/react-query`) będzie użyty do obsługi mutacji (wysyłania danych) do endpointu `POST /api/v1/auth/register`. Zapewni on obsługę stanów `isLoading`, `isError`, `isSuccess` oraz callbacki `onSuccess` i `onError`.
- **`AuthContext`**: Globalny kontekst Reacta do zarządzania stanem uwierzytelnienia w całej aplikacji. Po pomyślnej rejestracji, `session_token` i dane użytkownika zostaną zapisane w tym kontekście.

## 7. Integracja API
Integracja z API będzie realizowana za pomocą `react-query` i `axios` (lub `fetch`).

- **Endpoint**: `POST /api/v1/auth/register`
- **Akcja**: Wywołanie mutacji `useRegister` po kliknięciu przycisku "Zarejestruj się".
- **Request**:
  - **Typ**: `RegisterCommand`
  - **Payload**: Dane z formularza `RegisterFormViewModel` zmapowane na `RegisterCommand`.
- **Response (Success)**:
  - **Typ**: `RegisterResponseDTO`
  - **Obsługa**: W callbacku `onSuccess` mutacji:
    1. Zapisz `session_token` i dane użytkownika w `AuthContext`.
    2. Przekieruj użytkownika na stronę główną (`/`).
    3. Wyświetl powiadomienie (`Toast`) o pomyślnej rejestracji.
- **Response (Error)**:
  - **Typ**: `ErrorResponseDTO`
  - **Obsługa**: W callbacku `onError` mutacji:
    1. Przeanalizuj kod błędu (`error.response.data.error.code`).
    2. Wyświetl odpowiedni komunikat błędu dla użytkownika (np. w `Toast` lub przy konkretnym polu formularza).

## 8. Interakcje użytkownika
- **Wprowadzanie danych**: Użytkownik wypełnia pola formularza. Stan jest zarządzany przez `react-hook-form`.
- **Kliknięcie "Zarejestruj się"**:
  1. `react-hook-form` uruchamia walidację po stronie klienta.
  2. Jeśli walidacja przejdzie, wywoływana jest mutacja `useRegister`.
  3. Przycisk `SubmitButton` przechodzi w stan ładowania (`isLoading`).
- **Pomyślna rejestracja**: Przycisk wraca do stanu aktywnego, a użytkownik jest przekierowywany.
- **Błąd rejestracji**: Przycisk wraca do stanu aktywnego, a użytkownik widzi komunikat błędu.

## 9. Warunki i walidacja
Walidacja będzie realizowana dwuetapowo: po stronie klienta (natychmiastowy feedback) i serwera (ostateczna weryfikacja). Do walidacji po stronie klienta zostanie użyty `zod`.

- **Email**:
  - Musi być poprawnym adresem e-mail.
  - Pole wymagane.
- **Hasło**:
  - Minimum 8 znaków.
  - Musi zawierać co najmniej jedną wielką literę, jedną małą literę, jedną cyfrę i jeden znak specjalny.
  - Pole wymagane.
- **Potwierdzenie hasła**:
  - Musi być identyczne z hasłem.
  - Pole wymagane.
- **Nazwa użytkownika X**:
  - Pole wymagane.
  - Walidacja istnienia konta jest przeprowadzana po stronie serwera po wysłaniu formularza.

## 10. Obsługa błędów
Komunikaty o błędach będą wyświetlane przy użyciu komponentu `Toast` z Shadcn/ui oraz jako tekst pomocniczy pod polami formularza.

- **Błędy walidacji klienta**: Wyświetlane natychmiast pod odpowiednimi polami formularza (np. "Hasła nie są zgodne").
- **Błędy API**:
  - **`400 Bad Request`**: Ogólny błąd "Nieprawidłowe dane."
  - **`409 Conflict` (Email already registered)**: Komunikat "Podany adres e-mail jest już zajęty."
  - **`422 Unprocessable Entity` (X username does not exist)**: Komunikat "Konto X o nazwie '{username}' nie istnieje. Sprawdź poprawność nazwy użytkownika." wyświetlony pod polem `xUsername`.
  - **`500 Internal Server Error` / `503 Service Unavailable`**: Ogólny komunikat "Wystąpił błąd serwera. Spróbuj ponownie później."

## 11. Kroki implementacji
1.  **Struktura plików**: Utwórz pliki `RegisterView.tsx` i `RegisterForm.tsx` w odpowiednich katalogach.
2.  **Definicja typów**: Zdefiniuj wszystkie potrzebne typy (DTO i ViewModel) w dedykowanym pliku `types.ts`.
3.  **Layout widoku**: Zaimplementuj komponent `RegisterView` z podstawowym layoutem i nagłówkiem.
4.  **Budowa formularza**: W komponencie `RegisterForm` stwórz formularz przy użyciu `react-hook-form` i komponentów z `Shadcn/ui` (`Input`, `Button`, `Form`).
5.  **Walidacja klienta**: Zdefiniuj schemat walidacji `zod` i zintegruj go z `react-hook-form` do walidacji pól po stronie klienta.
6.  **Integracja API**: Stwórz hook `useRegister` wykorzystujący `useMutation` z `react-query` do obsługi żądania `POST /api/v1/auth/register`.
7.  **Obsługa stanu wysyłania**: Połącz stan `isLoading` z mutacji z komponentem `SubmitButton`, aby wyświetlać wskaźnik ładowania.
8.  **Obsługa sukcesu**: W `onSuccess` mutacji zaimplementuj logikę zapisu tokena sesji (do `AuthContext`) i przekierowania użytkownika.
9.  **Obsługa błędów**: W `onError` mutacji zaimplementuj logikę wyświetlania komunikatów błędów na podstawie kodu statusu i treści odpowiedzi API.
10. **Routing**: Dodaj ścieżkę `/register` w głównym routerze aplikacji (`App.tsx` lub dedykowany plik routingu) i powiąż ją z `RegisterView`.
11. **Styling i responsywność**: Dopracuj wygląd widoku przy użyciu Tailwind CSS, upewniając się, że jest responsywny i zgodny z motywem (dark/light mode).
