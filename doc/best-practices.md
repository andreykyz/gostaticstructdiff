# gostaticstructdiff Best Practices

This document outlines recommended practices for using `gostaticstructdiff` effectively in production environments.

## Struct Design

### 1. Use Descriptive Field Names

Choose meaningful field names that clearly indicate purpose:

```go
// Good
type User struct {
    Username     string `structtomap:"username"`
    EmailAddress string `structtomap:"email_address"`
    IsActive     bool   `structtomap:"is_active"`
}

// Avoid
type User struct {
    U string `structtomap:"u"`
    E string `structtomap:"e"`
    A bool   `structtomap:"a"`
}
```

### 2. Consider Field Order

Place frequently changed fields first in the struct definition for better cache locality:

```go
type Document struct {
    Title     string    `structtomap:"title"`     // Frequently updated
    Content   string    `structtomap:"content"`   // Frequently updated
    AuthorID  int       `structtomap:"author_id"` // Rarely changed
    CreatedAt time.Time `structtomap:"created_at"` // Never changed
}
```

### 3. Use Pointers for Optional Fields

Use pointers to distinguish between "not set" and "set to zero value":

```go
type Config struct {
    // Optional field - nil means not specified
    Timeout *time.Duration `structtomap:"timeout"`
    
    // Required field - zero value has meaning
    Retries int `structtomap:"retries"`
}
```

### 4. Limit Struct Size

Large structs with many fields can lead to performance issues:

- **Keep structs focused**: Split large structs into logical groupings
- **Consider composition**: Use embedded structs for related fields
- **Benchmark**: Profile diff generation for structs with 50+ fields

```go
// Instead of one large struct
type UserProfile struct {
    // 50+ fields...
}

// Use composition
type UserBasic struct {
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
}

type UserPreferences struct {
    Theme    string `structtomap:"theme"`
    Language string `structtomap:"language"`
}

type UserProfile struct {
    Basic       UserBasic       `structtomap:"basic"`
    Preferences UserPreferences `structtomap:"preferences"`
}
```

## Tag Usage

### 1. Consistent Tag Naming

Use consistent naming conventions for `structtomap` tags:

```go
// Snake case (recommended)
type Product struct {
    ProductName string `structtomap:"product_name"`
    UnitPrice   float64 `structtomap:"unit_price"`
}

// Or camel case
type Product struct {
    ProductName string `structtomap:"productName"`
    UnitPrice   float64 `structtomap:"unitPrice"`
}
```

### 2. Tag All Relevant Fields

Ensure all fields that might change are tagged:

```go
type Order struct {
    ID         int       `structtomap:"id"`          // ✓
    Status     string    `structtomap:"status"`      // ✓
    Total      float64   `structtomap:"total"`       // ✓
    CreatedAt  time.Time `structtomap:"created_at"`  // ✓
    internalID int       // ✗ Not tagged - won't be diffed
}
```

### 3. Avoid Tagging Immutable Fields

Fields that never change don't need diffing:

```go
type Account struct {
    ID        int       `structtomap:"id"`         // Immutable after creation
    CreatedAt time.Time `structtomap:"created_at"` // Immutable
    Balance   float64   `structtomap:"balance"`    // Changes frequently
    // Consider omitting immutable fields from tags
}
```

## Performance Optimization

### 1. Batch Updates

When processing multiple updates, batch them to reduce overhead:

```go
// Instead of:
for _, update := range updates {
    diff := UserPatch(current, update)
    processDiff(diff)
}

// Consider:
var batchDiff UserDiff
for _, update := range updates {
    // Accumulate changes
    diff := UserPatch(current, update)
    batchDiff = mergeDiffs(batchDiff, diff)
}
processDiff(batchDiff)
```

### 2. Pool Diff Objects

Reuse diff structs to reduce allocations:

```go
var diffPool = sync.Pool{
    New: func() interface{} {
        return &UserDiff{}
    },
}

func getDiff() *UserDiff {
    return diffPool.Get().(*UserDiff)
}

func putDiff(d *UserDiff) {
    // Reset fields
    *d = UserDiff{}
    diffPool.Put(d)
}
```

### 3. Selective Generation

Only generate diffs for structs that actually need them:

```bash
# Generate only for frequently updated structs
gostaticstructdiff -input models.go -struct User,Order,Product
```

### 4. Profile and Benchmark

Regularly profile your application:

```go
func BenchmarkUserDiff(b *testing.B) {
    u1 := User{/* ... */}
    u2 := User{/* ... */}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = UserPatch(u1, u2)
    }
}
```

## Memory Management

### 1. Be Mindful of Map Copies

Map diffs create copies of map contents:

```go
// Original map with 1000 entries
original := make(map[string]string, 1000)

// Diff will copy all added/modified entries
// Consider using pointers for large map values
type LargeConfig struct {
    Data map[string]*LargeValue `structtomap:"data"`
}
```

### 2. Slice Considerations

Slice diffs replace entire slices:

```go
// Large slices are copied entirely
// Consider incremental updates for very large slices
type Document struct {
    // For large text, consider diff algorithms
    Content string `structtomap:"content"`
}
```

### 3. Pointer vs Value Semantics

Choose appropriate semantics for your use case:

```go
// Value semantics (default)
type User struct {
    Profile Profile `structtomap:"profile"` // Copied on diff
}

// Pointer semantics (less copying)
type User struct {
    Profile *Profile `structtomap:"profile"` // Pointer copied, struct not
}
```

## Testing Strategies

### 1. Test Generated Code

Always write tests for code that uses generated diffs:

```go
func TestUserPatchCompleteness(t *testing.T) {
    // Test that all fields are handled
    original := User{/* all fields set */}
    updated := User{/* all fields different */}
    
    diff := UserPatch(original, updated)
    
    // Verify all fields are marked as changed
    if !diff.Username.Set || !diff.Email.Set || !diff.Active.Set {
        t.Error("Not all fields marked as changed")
    }
}
```

### 2. Golden File Tests

Use golden files to ensure generated code doesn't regress:

```go
func TestGenerationGolden(t *testing.T) {
    generated := generateCode("testdata/user.go")
    golden := readFile("testdata/user_diff.golden.go")
    
    if generated != golden {
        t.Errorf("Generated code doesn't match golden file")
        // Optionally update golden file
        if *updateFlag {
            writeFile("testdata/user_diff.golden.go", generated)
        }
    }
}
```

### 3. Property-Based Testing

Use property-based tests to verify diff properties:

```go
func TestPatchInverse(t *testing.T) {
    // For all users u1, u2:
    // Patch(u1, u2) then Patch(u1, diff) should equal u2
    f := func(u1, u2 User) bool {
        diff := UserPatch(u1, u2)
        result := UserPatch(u1, diff)
        return result == u2
    }
    
    if err := quick.Check(f, nil); err != nil {
        t.Error(err)
    }
}
```

## Error Handling

### 1. Validate Diffs Before Application

```go
func ApplyUserUpdateSafe(original User, diff UserDiff) (User, error) {
    // Validate business rules
    if diff.Email.Set && !isValidEmail(diff.Email.Value) {
        return User{}, ErrInvalidEmail
    }
    
    if diff.Age.Set && (diff.Age.Value < 0 || diff.Age.Value > 150) {
        return User{}, ErrInvalidAge
    }
    
    return UserPatch(original, diff), nil
}
```

### 2. Handle Partial Application

```go
type ApplyResult struct {
    Success    bool
    Applied    User
    FailedFields []string
    Errors     []error
}

func ApplyUpdateWithValidation(original User, diff UserDiff) ApplyResult {
    result := ApplyResult{}
    
    // Try to apply field by field
    // Collect successes and failures
    
    return result
}
```

## Integration Patterns

### 1. API Design

Design APIs to accept diffs for partial updates:

```go
type UpdateUserRequest struct {
    Diff UserDiff `json:"diff"`
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    current := getCurrentUser()
    updated := UserPatch(current, req.Diff)
    
    if err := saveUser(updated); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}
```

### 2. Event Sourcing

Use diffs for event sourcing:

```go
type UserEvent struct {
    Timestamp time.Time
    UserID    string
    Diff      UserDiff
    Metadata  map[string]interface{}
}

func recordUserChange(userID string, diff UserDiff) {
    event := UserEvent{
        Timestamp: time.Now(),
        UserID:    userID,
        Diff:      diff,
        Metadata: map[string]interface{}{
            "source": "api",
            "ip":     getClientIP(),
        },
    }
    
    saveEvent(event)
}
```

### 3. Conflict Resolution

Implement conflict resolution using diffs:

```go
func resolveConflict(base, server, client User) (User, error) {
    serverDiff := UserPatch(base, server)
    clientDiff := UserPatch(base, client)
    
    // Merge strategy
    merged := mergeDiffs(serverDiff, clientDiff)
    
    return UserPatch(base, merged), nil
}
```

## Maintenance

### 1. Version Control Generated Files

Commit generated `*_diff.go` files to version control:

```bash
# .gitignore exception
!*_diff.go
```

### 2. Automate Regeneration

Use `go generate` in your build process:

```bash
# In CI/CD pipeline
go generate ./...
git diff --exit-code || echo "Generated files out of date"
```

### 3. Document Breaking Changes

When changing structs that affect generated code:

```markdown
## Breaking Changes in v2.0

- Removed `User.PhoneNumber` field
- Changed `User.Email` from `string` to `*string`
- Added `User.Metadata` map field

Migration: Regenerate all diff files and update client code.
```

## Security Considerations

### 1. Input Validation

Validate diffs from untrusted sources:

```go
func validateUserDiff(diff UserDiff) error {
    if diff.Admin.Set && diff.Admin.Value {
        // Only allow admin flag changes from authorized sources
        if !isAuthorizedAdminChange() {
            return ErrUnauthorized
        }
    }
    
    if diff.Balance.Set {
        // Validate balance changes
        if !isValidBalanceChange(diff.Balance.Value) {
            return ErrInvalidBalance
        }
    }
    
    return nil
}
```

### 2. Size Limits

Limit diff size to prevent DoS attacks:

```go
const maxDiffSize = 10 * 1024 * 1024 // 10MB

func processDiff(diff UserDiff) error {
    size := estimateDiffSize(diff)
    if size > maxDiffSize {
        return ErrDiffTooLarge
    }
    
    // Process diff
    return nil
}
```

## Monitoring and Observability

### 1. Metrics

Track diff usage metrics:

```go
var (
    diffComputeCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "diff_compute_total",
            Help: "Total number of diff computations",
        },
        []string{"struct_type"},
    )
    
    diffSizeBytes = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "diff_size_bytes",
            Help:    "Size of computed diffs in bytes",
            Buckets: prometheus.ExponentialBuckets(100, 10, 6),
        },
        []string{"struct_type"},
    )
)

func computeDiffWithMetrics(original, new User) UserDiff {
    start := time.Now()
    diff := UserPatch(original, new)
    
    diffComputeCount.WithLabelValues("User").Inc()
    diffSizeBytes.WithLabelValues("User").Observe(float64(estimateSize(diff)))
    
    return diff
}
```

### 2. Logging

Log significant diff operations:

```go
func logDiffOperation(operation string, diff UserDiff) {
    changedFields := []string{}
    if diff.Username.Set {
        changedFields = append(changedFields, "username")
    }
    if diff.Email.Set {
        changedFields = append(changedFields, "email")
    }
    
    log.WithFields(log.Fields{
        "operation":     operation,
        "changed_count": len(changedFields),
        "changed_fields": changedFields,
    }).Info("Diff operation completed")
}
```

## Conclusion

Following these best practices will help you:

1. **Improve performance** through proper struct design and optimization
2. **Increase reliability** with comprehensive testing
3. **Enhance security** through validation and monitoring
4. **Simplify maintenance** with automation and documentation

Remember to adapt these practices to your specific use case and regularly review your implementation as requirements evolve.