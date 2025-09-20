package comparator

type Comparator string

const (
	EQ       Comparator = "="
	NEQ      Comparator = "!="
	GT       Comparator = ">"
	GTE      Comparator = ">="
	LT       Comparator = "<"
	LTE      Comparator = "<="
	IN       Comparator = "in"
	CONTAINS Comparator = "contains"
	WITHOUT  Comparator = "without"
)
