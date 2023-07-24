# Capturer package

Captures error and returns a wrapped error with a flag
that signals that original error has been already captured/logged.

```go
var err Error = capturer.CaptureError(ctx, err)
```
