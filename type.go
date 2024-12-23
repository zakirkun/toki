package toki

// SQLExpression represents a raw SQL expression
type SQLExpression interface {
	SQL() string
}

// Raw creates a raw SQL expression
type Raw string

func (r Raw) SQL() string { return string(r) }
