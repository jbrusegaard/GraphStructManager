package types

type Comparitor string

const (
	EQ       Comparitor = "="
	NEQ      Comparitor = "!="
	GT       Comparitor = ">"
	GTE      Comparitor = ">="
	LT       Comparitor = "<"
	LTE      Comparitor = "<="
	IN       Comparitor = "in"
	CONTAINS Comparitor = "contains"
)
