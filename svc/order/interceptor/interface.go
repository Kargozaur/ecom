package interceptor

type TokenCarrier interface {
	GetToken() string
}
