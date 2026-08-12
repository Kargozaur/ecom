package token

type TokenType int

const (
	Access TokenType = iota
	Refresh
)
