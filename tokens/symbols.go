package tokens

type Comma struct{}

func (s Comma) Token() {}

func (s Comma) String() string { return "," }

type Equals struct{}

func (s Equals) Token() {}

func (s Equals) String() string { return "=" }

func NewNotEqualsToken(value string) Token {
	if value != "!=" && value != "<>" {
		panic(value + " is not a valid not equals token")
	}

	return NotEquals{Value: value}
}

type NotEquals struct {
	Value string
}

func (s NotEquals) Token() {}

func (s NotEquals) String() string { return s.Value }

type LessThan struct{}

func (s LessThan) Token() {}

func (s LessThan) String() string { return "<" }

type GreaterThan struct{}

func (s GreaterThan) Token() {}

func (s GreaterThan) String() string { return ">" }

type LessThanOrEqualTo struct{}

func (s LessThanOrEqualTo) Token() {}

func (s LessThanOrEqualTo) String() string { return "<=" }

type GreaterThanOrEqualTo struct{}

func (s GreaterThanOrEqualTo) Token() {}

func (s GreaterThanOrEqualTo) String() string { return ">=" }
