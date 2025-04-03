package validators

// URLValidatorInterface defines the contract for URL validation
type URLValidatorInterface interface {
	Validate(urlStr string) error
}
