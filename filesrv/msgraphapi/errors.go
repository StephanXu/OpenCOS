package msgraphapi

type (
	ErrorMessage struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	ErrorWrapper struct {
		Error ErrorMessage `json:"error"`
	}
)

type ApiCode int

type ApiError struct {
	Err  string
	Code ApiCode
}

const (
	InvalidAuthenticationToken ApiCode = 1000
)

func (p ApiError) Error() string {
	return p.Err
}

func NewApiError(code ApiCode, message string) *ApiError {
	return &ApiError{message, code}
}

func (p *MSGraphClient) ParseError(e *ErrorWrapper) *ApiError {
	if e.Error.Code == "InvalidAuthenticationToken" {
		return NewApiError(InvalidAuthenticationToken, e.Error.Code)
	}
	return nil
}
