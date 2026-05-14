# Error Handling

> How errors are handled in this project.

---

## Overview

<!--
Document your project's error handling conventions here.

Questions to answer:
- What error types do you define?
- How are errors propagated?
- How are errors logged?
- How are errors returned to clients?
-->

(To be filled by the team)

---

## Error Types

<!-- Custom error classes/types -->

(To be filled by the team)

---

## Error Handling Patterns

<!-- Try-catch patterns, error propagation -->

(To be filled by the team)

---

## API Error Responses

<!-- Standard error response format -->

(To be filled by the team)

---

## Common Mistakes

### Ignoring error return values from Close()

**Wrong**:
```go
if a.store != nil {
    if closer, ok := a.store.(interface{ Close() }); ok {
        closer.Close()  // error ignored!
    }
}
```

**Correct**:
```go
var errs []error
if a.store != nil {
    if closer, ok := a.store.(interface{ Close() error }); ok {
        if err := closer.Close(); err != nil {
            errs = append(errs, err)
        }
    }
}
// ... accumulate more errors ...
if len(errs) > 0 {
    return errors.Join(errs...)
}
return nil
```

**Why**: `Close()` errors indicate failure to flush/sync data. Ignoring them risks data loss.

(To be filled by the team)
