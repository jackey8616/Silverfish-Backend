package interf

// IRouter export
type IRouter interface {
	VerifyRecaptcha(token *string) (bool, error)
}
