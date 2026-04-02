# gostaticstructdiff Tutorial

This tutorial walks through practical examples of using `gostaticstructdiff` for various use cases, from basic to advanced.

## Prerequisites

- Go 1.26 or later installed
- `gostaticstructdiff` installed (see [Installation](./user-guide.md#installation))
- Basic understanding of Go structs and tags

## Tutorial 1: Basic User Management

### Step 1: Create the User Struct

Create a file `user.go`:

```go
package models

type User struct {
    ID       int    `structtomap:"id"`
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
    Active   bool   `structtomap:"active"`
}
```

### Step 2: Generate Diff Code

Run the generator:

```bash
gostaticstructdiff -input user.go -output user_diff.go
```

This creates `user_diff.go`:

```go
package models

type UserDiff struct {
    ID struct {
        Value int
        Set   bool
    }
    Username struct {
        Value string
        Set   bool
    }
    Email struct {
        Value string
        Set   bool
    }
    Active struct {
        Value bool
        Set   bool
    }
}

func UserPatch(original, new User) UserDiff {
    // Implementation computes diff between original and new
}

func UserPatch(original User, diff UserDiff) User {
    // Implementation applies diff to original
}
```

### Step 3: Use the Generated Code

Create a test file `user_test.go`:

```go
package models

import (
    "testing"
)

func TestUserPatch(t *testing.T) {
    // Original user
    original := User{
        ID:       1,
        Username: "alice",
        Email:    "alice@example.com",
        Active:   true,
    }
    
    // Updated user (only email changed)
    updated := User{
        ID:       1,
        Username: "alice",
        Email:    "alice@work.com",
        Active:   true,
    }
    
    // Compute diff
    diff := UserPatch(original, updated)
    
    // Verify only email is marked as changed
    if !diff.Email.Set {
        t.Error("Email should be marked as changed")
    }
    if diff.Email.Value != "alice@work.com" {
        t.Errorf("Expected email 'alice@work.com', got %s", diff.Email.Value)
    }
    
    // Verify other fields are not changed
    if diff.Username.Set || diff.ID.Set || diff.Active.Set {
        t.Error("Only email should be changed")
    }
    
    // Apply diff back to original
    patched := UserPatch(original, diff)
    
    // Verify patched equals updated
    if patched != updated {
        t.Errorf("Patched user doesn't match updated: %+v vs %+v", patched, updated)
    }
}
```

### Step 4: Run the Test

```bash
go test ./...
```

## Tutorial 2: Nested Structures

### Step 1: Create Nested Structs

Create `profile.go`:

```go
package models

type Address struct {
    Street string `structtomap:"street"`
    City   string `structtomap:"city"`
    Zip    string `structtomap:"zip"`
}

type Profile struct {
    Name    string  `structtomap:"name"`
    Age     int     `structtomap:"age"`
    Address Address `structtomap:"address"`
}
```

### Step 2: Generate Diff Code

```bash
gostaticstructdiff -input profile.go -output profile_diff.go
```

### Step 3: Examine Generated Code

The generated `ProfileDiff` will have nested `AddressDiff`:

```go
type ProfileDiff struct {
    Name struct {
        Value string
        Set   bool
    }
    Age struct {
        Value int
        Set   bool
    }
    Address struct {
        Value struct {
            Street struct {
                Value string
                Set   bool
            }
            City struct {
                Value string
                Set   bool
            }
            Zip struct {
                Value string
                Set   bool
            }
        }
        Set bool
    }
}
```

### Step 4: Test Nested Updates

```go
func TestNestedPatch(t *testing.T) {
    original := Profile{
        Name: "Bob",
        Age:  30,
        Address: Address{
            Street: "123 Main St",
            City:   "Anytown",
            Zip:    "12345",
        },
    }
    
    // Only change the city
    updated := Profile{
        Name: "Bob",
        Age:  30,
        Address: Address{
            Street: "123 Main St",
            City:   "Newtown",  // Changed
            Zip:    "12345",
        },
    }
    
    diff := ProfilePatch(original, updated)
    
    // Verify nested change detection
    if !diff.Address.Set {
        t.Error("Address should be marked as changed")
    }
    if !diff.Address.Value.City.Set {
        t.Error("City should be marked as changed")
    }
    if diff.Address.Value.City.Value != "Newtown" {
        t.Errorf("Expected city 'Newtown', got %s", diff.Address.Value.City.Value)
    }
}
```

## Tutorial 3: Maps and Slices

### Step 1: Create Struct with Collections

Create `config.go`:

```go
package models

type Config struct {
    Name     string            `structtomap:"name"`
    Settings map[string]string `structtomap:"settings"`
    Tags     []string          `structtomap:"tags"`
}
```

### Step 2: Generate and Examine

The generated `ConfigDiff` will have special handling for maps and slices:

```go
type ConfigDiff struct {
    Name struct {
        Value string
        Set   bool
    }
    Settings struct {
        Add map[string]string
        Del map[string]struct{}
        Set bool
    }
    Tags struct {
        Value []string
        Set   bool
    }
}
```

### Step 3: Test Map Operations

```go
func TestMapDiff(t *testing.T) {
    original := Config{
        Name: "server",
        Settings: map[string]string{
            "timeout": "30s",
            "retries": "3",
        },
        Tags: []string{"prod", "web"},
    }
    
    updated := Config{
        Name: "server",
        Settings: map[string]string{
            "timeout": "60s",  // Changed value
            "retries": "3",
            "cache":   "true", // Added key
            // "timeout" key stays, "cache" added, nothing deleted
        },
        Tags: []string{"prod", "api"}, // Changed slice
    }
    
    diff := ConfigPatch(original, updated)
    
    // Verify map changes
    if !diff.Settings.Set {
        t.Error("Settings should be marked as changed")
    }
    
    // Check added key
    if diff.Settings.Add["cache"] != "true" {
        t.Error("Added key 'cache' not found in Add map")
    }
    
    // Check modified value
    if diff.Settings.Add["timeout"] != "60s" {
        t.Error("Modified value for 'timeout' not correct")
    }
    
    // Verify slice change
    if !diff.Tags.Set {
        t.Error("Tags should be marked as changed")
    }
    if len(diff.Tags.Value) != 2 || diff.Tags.Value[1] != "api" {
        t.Error("Tags slice not updated correctly")
    }
}
```

## Tutorial 4: Real-World API Versioning

### Scenario: REST API with Partial Updates

Create `api/models.go`:

```go
package api

type UserUpdateRequest struct {
    Username *string           `structtomap:"username"`  // Pointer for optional
    Email    *string           `structtomap:"email"`     // Optional field
    Settings map[string]string `structtomap:"settings"`  // Partial map update
    Metadata struct {
        Version string `structtomap:"version"`
        Source  string `structtomap:"source"`
    } `structtomap:"metadata"`
}
```

### Generate Diff Code

```bash
gostaticstructdiff -input api/models.go -output api/models_diff.go
```

### API Handler Implementation

```go
package api

import (
    "encoding/json"
    "net/http"
)

type UserService interface {
    UpdateUser(userID string, diff UserUpdateRequestDiff) error
}

func UpdateUserHandler(service UserService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var update UserUpdateRequest
        if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        // Get current user from database
        currentUser, err := getUserFromDB(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Convert current user to request format
        currentRequest := convertToRequestFormat(currentUser)
        
        // Compute diff between current and update
        diff := UserUpdateRequestPatch(currentRequest, update)
        
        // Apply update using diff
        if err := service.UpdateUser(currentUser.ID, diff); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
    }
}
```

### Benefits

1. **Partial updates**: Only changed fields are processed
2. **Audit logging**: Diff can be logged for compliance
3. **Conflict detection**: Can detect concurrent modifications
4. **Efficient storage**: Store only diffs in history

## Tutorial 5: Integration with go generate

### Step 1: Add go:generate Directive

Add to your struct file:

```go
//go:generate gostaticstructdiff -input $GOFILE -output ${GOFILE%.go}_diff.go

package models

type Product struct {
    ID    int    `structtomap:"id"`
    Name  string `structtomap:"name"`
    Price float64 `structtomap:"price"`
}
```

### Step 2: Run go generate

```bash
# Generate for this file only
go generate ./models/product.go

# Or generate for entire package
go generate ./models/...
```

### Step 3: Automate in CI/CD

Add to your build script:

```bash
#!/bin/bash
# pre-build.sh

# Ensure generated code is up to date
go generate ./...

# Check if any files changed
if git diff --name-only | grep -q '_diff\.go$'; then
    echo "Generated files are out of date. Please run 'go generate ./...' and commit changes."
    exit 1
fi
```

## Tutorial 6: Custom Diff Strategies

### Scenario: Special Handling for Time Fields

Create a custom template for `time.Time` fields:

1. **Create custom template** `internal/templates/time_diff.tmpl`:

```go
{{define "time_field"}}
{{.FieldName}} struct {
    Value time.Time
    Set   bool
    Valid bool  // Additional validation flag
}
{{end}}
```

2. **Register custom type handler**:

```go
func init() {
    RegisterTypeHandler("time.Time", &TimeDiffStrategy{})
}

type TimeDiffStrategy struct{}

func (s *TimeDiffStrategy) Generate(field Field) (string, error) {
    return executeTemplate("time_field", field)
}
```

3. **Rebuild and test**:

```go
type Event struct {
    Timestamp time.Time `structtomap:"timestamp"`
    // ...
}
```

## Tutorial 7: Performance Optimization

### Benchmarking Diff Operations

Create benchmark tests:

```go
func BenchmarkUserPatch(b *testing.B) {
    original := User{/* ... */}
    updated := User{/* ... */}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = UserPatch(original, updated)
    }
}

func BenchmarkLargeStructPatch(b *testing.B) {
    // Test with struct having 50+ fields
}
```

### Optimization Techniques

1. **Pool diff objects**: Reuse diff structs to reduce allocations
2. **Lazy computation**: Compute diffs only when needed
3. **Batch operations**: Process multiple structs together
4. **Selective generation**: Only generate diffs for frequently updated fields

## Tutorial 8: Error Handling and Validation

### Validating Diff Application

```go
func ApplyUserUpdateSafe(original User, diff UserDiff) (User, error) {
    // Validate diff before application
    if diff.ID.Set && diff.ID.Value <= 0 {
        return User{}, errors.New("invalid ID in diff")
    }
    
    if diff.Email.Set && !isValidEmail(diff.Email.Value) {
        return User{}, errors.New("invalid email in diff")
    }
    
    // Apply diff
    return UserPatch(original, diff), nil
}
```

### Handling Partial Failures

```go
type UpdateResult struct {
    Success bool
    Applied UserDiff  // Fields that were successfully applied
    Failed  UserDiff  // Fields that failed
    Errors  []error
}

func ApplyUpdateWithRollback(original User, diff UserDiff) UpdateResult {
    result := UpdateResult{}
    
    // Try to apply each field individually
    // Track successes and failures
    // Provide rollback capability
    
    return result
}
```

## Common Patterns and Recipes

### Pattern 1: Change Notification

```go
func notifyOnChange(original, updated User, diff UserDiff) {
    if diff.Email.Set {
        sendEmailChangeNotification(original.Email, updated.Email)
    }
    
    if diff.Active.Set && !updated.Active {
        sendDeactivationNotification(original.Username)
    }
}
```

### Pattern 2: Audit Logging

```go
type AuditEntry struct {
    Timestamp time.Time
    UserID    string
    Diff      UserDiff
    Metadata  map[string]interface{}
}

func logUserChange(userID string, diff UserDiff) {
    entry := AuditEntry{
        Timestamp: time.Now(),
        UserID:    userID,
        Diff:      diff,
        Metadata: map[string]interface{}{
            "ip":        getClientIP(),
            "userAgent": getUserAgent(),
        },
    }
    
    saveAuditEntry(entry)
}
```

### Pattern 3: Conflict Resolution

```go
func resolveConflict(base, server, client User) (User, error) {
    serverDiff := UserPatch(base, server)
    clientDiff := UserPatch(base, client)
    
    // Merge diffs, preferring server for conflicts
    merged := mergeDiffs(serverDiff, clientDiff)
    
    return UserPatch(base, merged), nil
}
```

## Next Steps

1. **Explore the examples directory** for more complex scenarios
2. **Read the API reference** for detailed function documentation
3. **Check best practices** for production usage
4. **Contribute to the project** by adding new features or fixing bugs

## Troubleshooting Common Issues

### Issue: Generated code doesn't compile

**Solution**: Check for:
- Missing imports in source file
- Circular dependencies
- Unsupported field types
- Syntax errors in source structs

### Issue: Diff doesn't detect changes

**Solution**: Verify:
- All fields have `structtomap` tags
- Field types are supported
- You're comparing the right structs
- Zero values are handled correctly

### Issue: Performance problems

**Solution**: Consider:
- Reducing struct size
- Using pointers for large nested structs
- Batching updates
- Profiling to identify bottlenecks

## Additional Resources

- [Complete Example Code](../examples/)
- [API Reference](./api-reference.md)
- [Best Practices](./best-practices.md)
- [Contributing Guide](./contributing.md)