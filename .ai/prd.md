# Dokument wymagań produktu (PRD) - Ask Your Feed (MVP)

## 1. Przegląd produktu

Ask Your Feed to webowa aplikacja Q&A nad prywatnym feedem użytkownika z portalu X. System cyklicznie agreguje wyłącznie oryginalne wpisy obserwowanych kont (bez reply/retweet/repost/quote), konwertuje treści multimedialne do tekstu, zapisuje je w relacyjnej bazie i udostępnia prosty interfejs do zadawania pytań. Odpowiedzi generowane są przez LLM w trybie feed-only (bez przeglądania WWW) w oparciu o materiały z określonego zakresu czasu. Każda odpowiedź zawiera listę punktów oraz sekcję Źródła z linkami do postów. Aplikacja zapewnia historię zapytań/odpowiedzi z możliwością usuwania.

## 2. Problem użytkownika

Użytkownicy konsumują feed X z różnymi intencjami: szybki przegląd dnia vs. głębokie wejście w wątki od konkretnych osób i tematów. Manualne filtrowanie i wyszukiwanie jest czasochłonne, media (obrazy/wideo) wymagają interpretacji, a wątki i szum (odpowiedzi, retweety) utrudniają skrótowe wnioskowanie. Ask Your Feed automatyzuje selekcję i syntezę najważniejszych informacji z własnego feedu użytkownika oraz pozwala szybko zadawać pytania i otrzymywać odpowiedzi z cytowanymi źródłami.

## 3. Wymagania funkcjonalne

3.1. Konta i uwierzytelnianie
* Logowanie wyłącznie przez OAuth X; mapowanie 1:1: użytkownik aplikacji ↔ konto X.
* Szyfrowanie i rotacja tokenów dostępowych.
* Obsługa wygaśnięcia/odwołania dostępu (wymuszenie ponownej autoryzacji).

3.2. Agregacja i ingest feedu
* Zakres: tylko oryginalne wpisy obserwowanych kont; wątki traktowane jako oddzielne wpisy; edycje ignorowane.
* Harmonogram: co 15 min z jitterem ±3 min.
* Delta-fetch po since_id, odporność na rate-limit (retry z eksponentialnym backoffem).
* SLO: średni czas od publikacji do widoczności w aplikacji ≤1 h.
* Backfill przy pierwszym uruchomieniu: domyślnie ostatnie 24 h (patrz 4. Granice – kwestia do potwierdzenia).

3.3. Przechowywanie danych
* Relacyjna baza danych; pełnotekstowe przechowywanie treści postów i opisów mediów.
* Media: konwersja obrazów i krótkich wideo do tekstu podczas ingestu; limity wideo ≤90 s lub ≤25 MB; maks. 4 obrazy/post.
* Treści multimedialne przechowywane w formie tekstowej; nie przechowuje URL/miniatur – patrz 4. Granice.

3.4. Wyszukiwanie i indeksowanie
* Wyszukiwanie wyłącznie pełnotekstowe (bez indeksów wektorowych).
* Dostępne filtrowanie po dacie publikacji.

3.5. Pytania i odpowiedzi (Q&A)
* LLM działa w trybie feed-only (bez web-browsingu).
* Domyślny horyzont czasowy: 24 h; użytkownik może zmienić zakres dat.
* Format odpowiedzi: lista punktów + sekcja Źródła (≥3 posty, jeśli dostępne; w przeciwnym razie wszystkie dostępne).
* Alternatywne formaty wymuszone komendą w promptcie (np. „timeline”).
* Przy braku materiału zwracany jest komunikat z sugestią rozszerzenia zakresu.

3.6. Historia
* Lista zapytań/odpowiedzi z paginacją i podglądem odpowiedzi.
* Usuwanie pojedynczego wpisu i „Usuń wszystko”.
* Brak automatycznej retencji i brak eksportu.

3.7. Telemetria i koszt
* Telemetria wyłącznie operacyjna, bez PII (latencja, error rate, rate-limit hits, wykorzystanie budżetu).
* Budżet globalny aplikacji ≤15 €/msc; po wyczerpaniu natychmiastowa pauza ingestu i blokada nowych odpowiedzi LLM; czytelny komunikat w UI.

3.8. Platforma i UX
* Responsywna aplikacja web (PL), dark mode.
* Proste, spójne komunikaty błędów i stanów (brak materiału, rate-limit, budżet, brak autoryzacji).
* Wskazanie czasu ostatniej aktualizacji danych (np. „Ostatnia synchronizacja: 12:15”).

3.9. Bezpieczeństwo
* Szyfrowanie i rotacja tokenów OAuth X; bezpieczne składowanie sekretów.
* Minimalny zakres uprawnień OAuth niezbędny do odczytu feedu.

## 4. Granice produktu

4.1. Poza zakresem MVP
* Brak aktualizacji feedu w czasie rzeczywistym; wyłącznie interwałowo.
* Zaawansowane filtrowanie inne niż przez LLM i datę.
* Web-browsing i źródła spoza feedu użytkownika.
* Eksport historii i automatyczna retencja.
* Indeksy wektorowe, RAG, personalizowane rankingi.
* Analityka zachowań użytkowników (np. DAU/WAU) – poza telemetrią techniczną.
* Wielokrotne konta X na jednego użytkownika (mapowanie 1:1).

4.2. Ograniczenia i decyzje architektoniczne
* Twarde filtry ingestu: tylko oryginalne wpisy; wątki osobno; edycje ignorowane.
* Media zredukowane do tekstu, limity: wideo ≤90 s lub ≤25 MB; ≤4 obrazy/post.
* Źródła w odpowiedzi zawsze linkują do postów z feedu.
* Budżet kosztowy jest twardym limitem; brak degradacji – po wyczerpaniu następuje pauza i blokada.

## 5. Historyjki użytkowników

US-001<br />
Tytuł: Logowanie przez OAuth X<br />
Opis: Jako użytkownik chcę zalogować się przez X, aby aplikacja mogła czytać mój feed i odpowiadać na pytania.<br />
Kryteria akceptacji:
- Zakładając brak aktywnej sesji, kiedy kliknę „Zaloguj przez X”, wtedy zostanę przekierowany do autoryzacji OAuth i po sukcesie wrócę do aplikacji zalogowany.
- Zakładając udaną autoryzację, kiedy wrócę do aplikacji, wtedy zobaczę komunikat o rozpoczęciu ingestu.

US-002<br />
Tytuł: Onboarding i pierwszy ingest<br />
Opis: Jako nowy użytkownik chcę, by system zaczął pobierać mój feed i przygotował dane do Q&A.<br />
Kryteria akceptacji:<br />
- Po pierwszym zalogowaniu ingest uruchamia się automatycznie.
- Zakres backfillu domyślnie wynosi 24 h (do potwierdzenia).
- Widzę wskaźnik „Ostatnia synchronizacja” po zakończeniu pierwszego cyklu.

US-003<br />
Tytuł: Szybkie pytanie z domyślnym zakresem 24 h<br />
Opis: Jako użytkownik chcę zadać pytanie i otrzymać odpowiedź z ostatnich 24 h.<br />
Kryteria akceptacji:
- Gdy wpiszę pytanie bez parametrów, odpowiedź opiera się na danych z 24 h.
- Odpowiedź jest listą punktów i zawiera sekcję Źródła.

US-004<br />
Tytuł: Filtrowanie po dacie<br />
Opis: Jako użytkownik chcę ustawić zakres dat dla Q&A.<br />
Kryteria akceptacji:
- UI pozwala wskazać od-do lub predefiniowane zakresy (24 h, 7 dni).
- LLM używa tylko postów z wybranego zakresu.
- Źródła w odpowiedzi mieszczą się w tym zakresie.

US-005<br />
Tytuł: Alternatywny format odpowiedzi komendą w promptcie<br />
Opis: Jako użytkownik chcę wymusić format (np. „timeline”).<br />
Kryteria akceptacji:
- Wpisanie komendy w promptcie zmienia format odpowiedzi zgodnie z dokumentacją.
- Sekcja Źródła nadal jest dołączona, o ile nie określono inaczej.

US-006<br />
Tytuł: Sekcja Źródła z min. 3 linkami<br />
Opis: Jako użytkownik chcę widzieć referencje do postów, aby zweryfikować odpowiedź.<br />
Kryteria akceptacji:
- Jeśli dostępne ≥3 posty, Źródła zawierają co najmniej 3 linki.
- Jeśli dostępne <3 posty, Źródła zawierają wszystkie dostępne linki.
- Linki prowadzą do oryginalnych postów w X.

US-007<br />
Tytuł: Brak treści w danym zakresie<br />
Opis: Jako użytkownik chcę jasny komunikat, jeśli nie ma treści.<br />
Kryteria akceptacji:
- Gdy brak trafień, system nie podaje fałszywych treści i zwraca komunikat o jej braku.
- Komunikat sugeruje rozszerzenie zakresu dat.
- Nie pojawia się sekcja Źródła, jeśli brak jakichkolwiek postów.

US-008<br />
Tytuł: Uwzględnianie treści multimedialnych<br />
Opis: Jako użytkownik chcę, by obrazy i krótkie wideo z feedu były uwzględniane w odpowiedziach.<br />
Kryteria akceptacji:
- Media są konwertowane do tekstu podczas ingestu i widoczne dla LLM.
- Źródła mogą odwoływać się do postów zawierających media.
- Wideo >90 s lub >25 MB jest pomijane.

US-009<br />
Tytuł: Historia zapytań – lista z paginacją<br />
Opis: Jako użytkownik chcę przejrzeć wcześniejsze pytania i odpowiedzi.<br />
Kryteria akceptacji:
- Widok historii pokazuje listę pozycji z paginacją.
- Każda pozycja zawiera skrót pytania, datę i link do podglądu odpowiedzi.
- Parametry paginacji są stałe w MVP

US-010<br />
Tytuł: Podgląd odpowiedzi z historii<br />
Opis: Jako użytkownik chcę otworzyć pełną treść wcześniejszej odpowiedzi.<br />
Kryteria akceptacji:
- Kliknięcie pozycji historii otwiera szczegóły z pełnym tekstem i Źródłami.
- Widok szczegółów jest tylko do odczytu.

US-011<br />
Tytuł: Usuwanie pojedynczej pozycji z historii<br />
Opis: Jako użytkownik chcę usunąć wybraną odpowiedź z historii.<br />
Kryteria akceptacji:
- Kliknięcie „Usuń” przy pozycji nie wymaga potwierdzenia.
- Po potwierdzeniu pozycja znika i nie jest dostępna w UI.

US-012<br />
Tytuł: „Usuń wszystko” w historii<br />
Opis: Jako użytkownik chcę skasować całą historię.<br />
Kryteria akceptacji:
- Akcja nie wymaga potwierdzenia.
- Po potwierdzeniu historia jest pusta.
- Operacja nie wpływa na token OAuth ani dane ingestu.

US-013<br />
Tytuł: Wyczerpany budżet kosztowy<br />
Opis: Jako użytkownik chcę jednoznaczny komunikat i przewidywalne zachowanie, gdy budżet zostanie wyczerpany.<br />
Kryteria akceptacji:
- Po wyczerpaniu budżetu próba zadania pytania zwraca komunikat o blokadzie.
- Ingest zostaje wstrzymany do odnowienia budżetu.
- UI nie oferuje degradacji jakości – jedynie blokadę.

US-014<br />
Tytuł: Prywatność telemetrii<br />
Opis: Jako użytkownik nie chcę, aby aplikacja gromadziła moje PII w logach/metrykach.<br />
Kryteria akceptacji:
- Telemetria zawiera tylko metryki operacyjne (latencja, error rate, rate-limit hits, budżet).
- Brak zapisów PII w telemetrycznych eventach.

US-015<br />
Tytuł: Rotacja i bezpieczne składowanie tokenów<br />
Opis: Jako użytkownik chcę, by moje tokeny OAuth były zabezpieczone.<br />
Kryteria akceptacji:
- Tokeny są przechowywane w szyfrowanym magazynie.
- Rotacja odbywa się zgodnie z polityką X lub po wykryciu incydentu.

US-016<br />
Tytuł: Wylogowanie<br />
Opis: Jako użytkownik chcę zakończyć sesję w aplikacji.<br />
Kryteria akceptacji:
- Kliknięcie „Wyloguj” unieważnia sesję aplikacji.
- Po wylogowaniu nie mam dostępu do historii bez ponownego logowania.

US-017<br />
Tytuł: Wskaźnik świeżości danych<br />
Opis: Jako użytkownik chcę wiedzieć, kiedy dane były ostatnio aktualizowane.<br />
Kryteria akceptacji:
- UI wyświetla znacznik czasu „Ostatnia synchronizacja”.
- Wartość aktualizuje się po każdym udanym cyklu ingestu.

US-018<br />
Tytuł: Tylko oryginalne wpisy jako źródła<br />
Opis: Jako użytkownik nie chcę widzieć odpowiedzi/retweetów w źródłach.<br />
Kryteria akceptacji:
- Źródła w odpowiedzi pochodzą wyłącznie z oryginalnych wpisów obserwowanych kont.
- Wątki traktowane są jako oddzielne wpisy i mogą pojawiać się niezależnie.

US-019<br />
Tytuł: Obsługa limitów mediów<br />
Opis: Jako użytkownik chcę stabilności, gdy media przekraczają limity.<br />
Kryteria akceptacji:
- Wideo przekraczające limity nie jest przetwarzane; system działa dalej.
- Odpowiedzi nie odwołują się do pominiętych mediów.

US-020<br />
Tytuł: Błędy i stany systemowe w UI<br />
Opis: Jako użytkownik chcę proste i spójne komunikaty błędów.<br />
Kryteria akceptacji:
- Stany: brak treści, rate-limit, brak autoryzacji, wyczerpany budżet – każdy posiada czytelny komunikat.
- Komunikaty są w języku polskim, zgodne z dark mode.

US-021<br />
Tytuł: Zapytanie wykraczające poza feed<br />
Opis: Jako użytkownik chcę, aby system nie zmyślał odpowiedzi, jeśli temat nie występuje w moim feedzie.<br />
Kryteria akceptacji:
- Brak dowodów skutkuje komunikatem o braku treści.
- Źródła nigdy nie zawierają linków spoza feedu.

## 6. Metryki sukcesu

6.1. SLO i niezawodność
* Średni czas od publikacji do widoczności ≤1 h (metryka: różnica czasu publikacji postu vs. dostępność w wyszukiwaniu/LLM – definicja w 4.3.g).
* Stabilność ingestu: niski odsetek błędów i kontrola rate-limit hits (metryki: error rate, liczba retry, 429/min).
* Dostępność Q&A: odsetek udanych odpowiedzi vs. błędów systemowych.

6.2. Koszt i kontrola budżetu
* Wykorzystanie budżetu ≤15 €/msc; zero przekroczeń w normalnym użyciu.
* W przypadku wyczerpania – poprawne zastosowanie polityki pauzy i blokady (metryka: czas w stanie pauzy, liczba zablokowanych żądań Q&A).

6.3. Doświadczenie użytkownika
* Czas odpowiedzi Q&A w typowych warunkach MVP w akceptowalnym przedziale (target do ustalenia w testach wydajności).
* Jakość odpowiedzi: każda odpowiedź zawiera sekcję Źródła (≥3 linki, jeśli dostępne).
* Kompletność historii: brak błędów przy paginacji, działające usuwanie per wpis i „Usuń wszystko”.

6.4. Metryka produktowa celu MVP
* 50% użytkowników tworzy zapytania co najmniej raz w tygodniu – metryka docelowa produktu; w MVP formalne śledzenie zachowań nie jest wdrożone (pomiar planowany po MVP).

---

Lista kontrolna (wewnętrzna):
- Każda historia ma testowalne kryteria akceptacji.
- Kryteria są jasne i konkretne; obejmują scenariusze podstawowe, alternatywne i skrajne (brak materiału, rate-limit, budżet).
- Zestaw historyjek pokrywa pełne MVP: auth, ingest, Q&A, filtry dat, media→tekst, Źródła, historia, błędy, koszt.
- Wymagania dot. uwierzytelniania i bezpieczeństwa uwzględnione (US-001, US-015, US-017, US-018).
