package schema

// Errors returned by handlers.
type Errors map[string]string

// SimpleError returns an Errors instance with a single error.
func SimpleError(err error) Errors {
	return Errors{"error": err.Error()}
}
