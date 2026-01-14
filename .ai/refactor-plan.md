## Analiza TOP 5 plików o największej liczbie linii kodu w folderze backend

### 1) TOP 5 plików według LOC (Lines of Code):

1. **backend/test/integration/following_integration_test.go** - 736 LOC
2. **backend/test/integration/qa_integration_test.go** - 715 LOC
3. **backend/test/integration/ingest_status_test.go** - 542 LOC
4. **backend/internal/repositories/qa_repository.go** - 356 LOC
5. **backend/internal/handlers/qa_handler.go** - 342 LOC

---

### 2) Rekomendacje refaktoryzacji dla każdego pliku:

---

#### **Plik #1: following_integration_test.go (736 LOC)**

**Zidentyfikowane problemy:**
- Bardzo długi plik testowy z wieloma powtarzającymi się wzorcami
- Duplikacja kodu w setupie testów (tworzenie użytkowników, autorów, relacji)
- Każdy test case zawiera podobną strukturę: setup → request → assertions
- Brak wydzielonych helper functions dla wspólnych operacji

**Rekomendacje refaktoryzacji:**

1. **Table-Driven Tests Pattern** - Zastosuj wzorzec table-driven tests dla podobnych scenariuszy:
```go
func TestFollowingPagination(t *testing.T) {
    testCases := []struct {
        name           string
        limit          int
        cursor         int64
        expectedItems  int
        expectedHasMore bool
        expectedCursor int64
    }{
        {"FirstPage", 2, 0, 2, true, 400},
        {"SecondPage", 2, 400, 2, true, 200},
        {"ThirdPage", 2, 200, 1, false, 0},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

2. **Test Fixtures & Builders** - Wprowadź pattern Builder dla tworzenia danych testowych:
```go
type FollowingTestFixture struct {
    db     *sqlx.DB
    userID uuid.UUID
}

func (f *FollowingTestFixture) WithAuthors(count int) *FollowingTestFixture {
    // Create authors
    return f
}

func (f *FollowingTestFixture) WithFollowing(authorIDs []int64) *FollowingTestFixture {
    // Create following relationships
    return f
}
```

3. **Assertion Helpers** - Wydziel funkcje pomocnicze dla powtarzających się asercji:
```go
func assertFollowingResponse(t *testing.T, response dto.FollowingListResponseDTO, expected ExpectedResponse) {
    t.Helper()
    assert.Equal(t, expected.itemCount, len(response.Items))
    assert.Equal(t, expected.totalCount, response.TotalCount)
    assert.Equal(t, expected.hasMore, response.HasMore)
}
```

4. **Subtests Organization** - Pogrupuj testy w logiczne sekcje używając subtestów z shared setup.

**Argumentacja:** Redukcja duplikacji kodu o ~40%, lepsza czytelność, łatwiejsze dodawanie nowych przypadków testowych, zgodność z Go testing best practices.

---

#### **Plik #2: qa_integration_test.go (715 LOC)**

**Zidentyfikowane problemy:**
- Podobne problemy jak w pliku #1
- Zakomentowany kod (testCreateQAHappyPath) - code smell
- Powtarzające się wzorce setupu i teardownu
- Brak centralizacji komunikatów błędów

**Rekomendacje refaktoryzacji:**

1. **Shared Test Context** - Wprowadź strukturę kontekstu testowego:
```go
type QATestContext struct {
    router  *gin.Engine
    db      *sqlx.DB
    helper  *DatabaseHelper
    userID  uuid.UUID
}

func setupQATest(t *testing.T) *QATestContext {
    // Common setup
}

func (ctx *QATestContext) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
    // Centralized request making
}
```

2. **Test Data Factories** - Zastosuj Factory Pattern dla danych testowych:
```go
type QAFactory struct {
    db *sqlx.DB
}

func (f *QAFactory) CreateQAWithSources(userID uuid.UUID, sourcesCount int) string {
    // Create QA with specified number of sources
}
```

3. **Custom Matchers/Assertions** - Wprowadź custom matchers dla złożonych asercji:
```go
func AssertQADetail(t *testing.T, actual dto.QADetailDTO) *QADetailAssertion {
    return &QADetailAssertion{t: t, actual: actual}
}

func (a *QADetailAssertion) HasQuestion(expected string) *QADetailAssertion {
    assert.Equal(a.t, expected, a.actual.Question)
    return a
}
```

4. **Remove Dead Code** - Usuń zakomentowany kod lub przenieś do osobnego pliku z adnotacją TODO.

**Argumentacja:** Eliminacja zakomentowanego kodu, redukcja boilerplate o ~35%, lepsza maintainability, zgodność z Clean Code principles.

---

#### **Plik #3: ingest_status_test.go (542 LOC)**

**Zidentyfikowane problemy:**
- TestMain z globalnym setupem - może powodować problemy z izolacją testów
- Duplikacja logiki walidacji parametrów (limit validation powtarza się)
- Brak parametryzacji dla edge cases

**Rekomendacje refaktoryzacji:**

1. **Refactor TestMain** - Przenieś setup do per-test fixtures zamiast globalnego TestMain:
```go
// Zamiast globalnego TestMain, użyj:
func setupTestDB(t *testing.T) *DatabaseHelper {
    t.Helper()
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    return NewDatabaseHelper(t)
}
```

2. **Parametrized Edge Cases** - Użyj table-driven tests dla walidacji:
```go
func TestIngestStatusValidation(t *testing.T) {
    validationTests := []struct {
        name          string
        queryParam    string
        expectedCode  int
        expectedError string
    }{
        {"LimitExceedsMax", "limit=51", 400, "INVALID_LIMIT"},
        {"LimitZero", "limit=0", 400, "INVALID_LIMIT"},
        {"LimitNegative", "limit=-5", 400, "INVALID_LIMIT"},
        {"LimitNotNumber", "limit=abc", 400, "INVALID_LIMIT"},
    }
    
    for _, tt := range validationTests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

3. **Test Isolation** - Zapewnij pełną izolację testów używając t.Parallel() gdzie możliwe:
```go
func TestIngestStatusIntegration(t *testing.T) {
    t.Run("HappyPath", func(t *testing.T) {
        t.Parallel()
        dbHelper := setupTestDB(t)
        defer dbHelper.Close()
        // Test implementation
    })
}
```

**Argumentacja:** Lepsza izolacja testów, możliwość równoległego uruchamiania, redukcja kodu o ~30%, zgodność z Go testing best practices.

---

#### **Plik #4: qa_repository.go (356 LOC)**

**Zidentyfikowane problemy:**
- Funkcja `mustParseTime` używa panic - niebezpieczne w production code
- Brak wykorzystania prepared statements dla często wykonywanych zapytań
- Długa funkcja `ListQA` z dynamicznym budowaniem query (SQL injection risk)
- Brak connection pooling configuration hints

**Rekomendacje refaktoryzacji:**

1. **Query Builder Pattern** - Wprowadź bezpieczny query builder:
```go
type QAQueryBuilder struct {
    baseQuery string
    conditions []string
    args      []interface{}
    argIndex  int
}

func (qb *QAQueryBuilder) WithUserID(userID uuid.UUID) *QAQueryBuilder {
    qb.conditions = append(qb.conditions, fmt.Sprintf("qa.user_id = $%d", qb.argIndex))
    qb.args = append(qb.args, userID)
    qb.argIndex++
    return qb
}

func (qb *QAQueryBuilder) WithCursor(cursor string) *QAQueryBuilder {
    if cursor != "" {
        qb.conditions = append(qb.conditions, fmt.Sprintf("qa.id < $%d", qb.argIndex))
        qb.args = append(qb.args, cursor)
        qb.argIndex++
    }
    return qb
}

func (qb *QAQueryBuilder) Build() (string, []interface{}) {
    query := qb.baseQuery + " WHERE " + strings.Join(qb.conditions, " AND ")
    return query, qb.args
}
```

2. **Safe Time Parsing** - Zastąp panic graceful error handling:
```go
func parseTime(s string) (time.Time, error) {
    formats := []string{
        time.RFC3339,
        "2006-01-02 15:04:05.999999-07",
        "2006-01-02 15:04:05",
    }
    
    for _, format := range formats {
        if t, err := time.Parse(format, s); err == nil {
            return t, nil
        }
    }
    
    return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
```

3. **Repository Method Decomposition** - Rozdziel złożone metody:
```go
func (r *QARepository) ListQA(ctx context.Context, userID uuid.UUID, limit int, cursor string) (*dto.QAListResponseDTO, error) {
    rows, err := r.fetchQARows(ctx, userID, limit, cursor)
    if err != nil {
        return nil, err
    }
    
    return r.buildQAListResponse(rows, limit), nil
}

func (r *QARepository) fetchQARows(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]QARow, error) {
    // Query execution
}

func (r *QARepository) buildQAListResponse(rows []QARow, limit int) *dto.QAListResponseDTO {
    // Response building
}
```

4. **Prepared Statements Cache** - Dla często wykonywanych zapytań:
```go
type QARepository struct {
    db              *sqlx.DB
    stmtGetByID     *sqlx.Stmt
    stmtDeleteByID  *sqlx.Stmt
}

func NewQARepository(database *sqlx.DB) (*QARepository, error) {
    repo := &QARepository{db: database}
    
    var err error
    repo.stmtGetByID, err = database.Preparex("SELECT ... WHERE id = $1 AND user_id = $2")
    if err != nil {
        return nil, err
    }
    
    return repo, nil
}
```

**Argumentacja:** Eliminacja panic w production code, zwiększenie bezpieczeństwa (SQL injection prevention), lepsza performance przez prepared statements, zgodność z Repository Pattern i SOLID principles.

---

#### **Plik #5: qa_handler.go (342 LOC)**

**Zidentyfikowane problemy:**
- Duplikacja kodu w każdej metodzie (user_id extraction, error handling)
- Brak middleware dla wspólnej logiki (auth, validation)
- Hardcoded error messages - brak i18n support
- Walidacja rozproszona między handler a validator

**Rekomendacje refaktoryzacji:**

1. **Middleware Pattern** - Wydziel wspólną logikę do middleware:
```go
// middleware/auth.go
func ExtractUserID() gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDValue, exists := c.Get("user_id")
        if !exists {
            respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Nieprawidłowy lub wygasły token sesji", nil)
            c.Abort()
            return
        }
        
        userID, ok := userIDValue.(uuid.UUID)
        if !ok {
            respondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Invalid user ID format", nil)
            c.Abort()
            return
        }
        
        c.Set("validated_user_id", userID)
        c.Next()
    }
}
```

2. **Handler Base Class/Composition** - Wprowadź bazową strukturę dla wspólnej funkcjonalności:
```go
type BaseHandler struct {
    validator *validator.Validate
}

func (h *BaseHandler) GetUserID(c *gin.Context) (uuid.UUID, error) {
    userID, exists := c.Get("validated_user_id")
    if !exists {
        return uuid.Nil, errors.New("user_id not found in context")
    }
    return userID.(uuid.UUID), nil
}

type QAHandler struct {
    BaseHandler
    qaService *services.QAService
}
```

3. **Error Response Builder** - Centralizuj error handling:
```go
type ErrorBuilder struct {
    statusCode int
    code       string
    message    string
    details    map[string]interface{}
}

func NewErrorBuilder(code string) *ErrorBuilder {
    return &ErrorBuilder{
        code:    code,
        details: make(map[string]interface{}),
    }
}

func (eb *ErrorBuilder) WithStatus(status int) *ErrorBuilder {
    eb.statusCode = status
    return eb
}

func (eb *ErrorBuilder) WithMessage(msg string) *ErrorBuilder {
    eb.message = msg
    return eb
}

func (eb *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
    eb.details[key] = value
    return eb
}

func (eb *ErrorBuilder) Send(c *gin.Context) {
    c.JSON(eb.statusCode, dto.ErrorResponseDTO{
        Error: dto.ErrorDetailDTO{
            Code:    eb.code,
            Message: eb.message,
            Details: eb.details,
        },
    })
}
```

4. **Validation Chain Pattern** - Uporządkuj walidację:
```go
type ValidationChain struct {
    errors []error
}

func (vc *ValidationChain) Validate(fn func() error) *ValidationChain {
    if err := fn(); err != nil {
        vc.errors = append(vc.errors, err)
    }
    return vc
}

func (vc *ValidationChain) HasErrors() bool {
    return len(vc.errors) > 0
}

// Usage:
chain := &ValidationChain{}
chain.
    Validate(func() error { return validateQuestion(cmd.Question) }).
    Validate(func() error { return validateDateRange(dateFrom, dateTo) })

if chain.HasErrors() {
    // Handle errors
}
```

5. **Constants for Error Messages** - Wprowadź stałe dla komunikatów:
```go
const (
    ErrMsgUnauthorized     = "Nieprawidłowy lub wygasły token sesji"
    ErrMsgQuestionRequired = "Pytanie jest wymagane i nie może być puste"
    ErrMsgQuestionTooLong  = "Pytanie przekracza maksymalną długość 2000 znaków"
    // ...
)
```

**Argumentacja:** Redukcja duplikacji kodu o ~50%, lepsza separacja concerns, łatwiejsza internationalization, zgodność z DRY principle i Gin best practices, łatwiejsze testowanie przez wydzielenie logiki do middleware.

---

### Podsumowanie:

Wszystkie zidentyfikowane pliki wykazują wysoką złożoność wynikającą głównie z:
- **Duplikacji kodu** (szczególnie w testach)
- **Braku abstrakcji** dla wspólnych operacji
- **Monolitycznych funkcji** wymagających dekompozycji

Zastosowanie zaproponowanych refaktoryzacji pozwoli na:
- Redukcję LOC o 30-50% przy zachowaniu funkcjonalności
- Zwiększenie testowalności i maintainability
- Lepszą zgodność z Go idioms i Clean Architecture
- Ułatwienie rozbudowy systemu w przyszłości