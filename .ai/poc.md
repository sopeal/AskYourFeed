# Aplikacja - "Ask Your Feed" (MVP)

## Główny problem
Kierują nami różne motywacje, gdy przeglądamy portale takie jak X. Czasem zależy nam jedynie na szybkim przeglądzie najważniejszych wydarzeń dnia, innym razem chcemy zagłębić się w opinie osób, które śledzimy w ramach konkretnego tematu. Pozwól sztucznej inteligencji przeszukiwać wiadomości, selekcjonować najciekawsze treści i błyskawicznie odpowiadać na Twoje pytania.

## Najmniejszy zestaw funkcjonalności
- Prosty sytem kont użytkowników do powiązania użytkownika z jego feedem.
- Agregowanie danych z portalu X i zapisywanie do relacyjnej bazy danych.
- Filtrowanie danych za pomocą daty.
- LLM na podstawie zapytania od użytkownika analizuje informacje i udziela odpowiedzi.
- Dostęp do historii zapytań i odpowiedzi LLMu. Możliwość usuwania.
- Obrazki i wideo z feedu są również analizowane.

## Co NIE wchodzi w zakres MVP
- Feed nie jest uaktualniany w czasie rzeczywistym a interwałowo.
- Zaawansowane filtrowanie w sposób inny niż za pomocą LLMa.

## Kryteria sukcesu
- 50% użytkowników tworzy zapytania co najmniej raz w tygodniu