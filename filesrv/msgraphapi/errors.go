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
	Err     string
	Code    ApiCode
	Message string
}

const (
	UnknownError               ApiCode = -1
	InvalidAuthenticationToken ApiCode = 1000
)

func (p ApiError) Error() string {
	return p.Err + ": " + p.Message
}

func NewApiError(code ApiCode, errCode string, message string) *ApiError {
	return &ApiError{errCode, code, message}
}

func (p *MSGraphClient) ParseError(e *ErrorWrapper) *ApiError {
	if e.Error.Code == "InvalidAuthenticationToken" {
		return NewApiError(InvalidAuthenticationToken, e.Error.Code, e.Error.Message)
	}
	return NewApiError(UnknownError, e.Error.Code, e.Error.Message)
}
