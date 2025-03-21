package toki

// SQLExpression represents a raw SQL expression
type SQLExpression interface {
	SQL() string
}

// Raw creates a raw SQL expression
type Raw string

func (r Raw) SQL() string { return string(r) }

type Operator string

const (
	// Between operator
	AND Operator = "AND"
	// OR
	OR Operator = "OR"
)

// JoinType represents different types of JOINs
type JoinType string

const (
	INNER_JOIN JoinType = "INNER JOIN"
	LEFT_JOIN  JoinType = "LEFT JOIN"
	RIGHT_JOIN JoinType = "RIGHT JOIN"
	FULL_JOIN  JoinType = "FULL JOIN"
)
