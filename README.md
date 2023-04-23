# resperr

Resperr is a Go package to associate status codes and messages with errors.
This package is forked from https://godoc.org/github.com/carlmjohnson/resperr.
Compared to the original, this package allows to customize the default status codes and error messages.
It also provides a more flexible API.

## Example usage

Attaching information to an error
```go
if err != nil {
	// Attach message and status code.
	return resperr.WithCodeAndMessage(err, http.StatusBadRequest, "You did something wrong.")
}

if err != nil {
	// Attach status code.
	return resperr.WithStatusCode(err, http.StatusBadRequest)
}

if err != nil {
	// Attach message.
	return resperr.WithUserMessage(err, "Could not fetch data.")
}

if !ok {
	// Generate a new error with a message attached.
	return resperr.WithUserMessage(nil, "Something is not ok.")
}
```

Getting information from an error
```go
func writeError(w http.ResponseWriter, err error) {
	// Default status code and message can be specified.
	w.WriteHeader(resperr.StatusCode(err))
	w.Write([]byte(`{"error":"` + resperr.UserMessage(err) + `"}`))
}
```