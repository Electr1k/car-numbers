package valuation

// Result - результат предсказания цены
type Result struct {
	Number     string
	Region     *Region
	Price      Price
	Confidence string
	Breakdown  Breakdown
}

type Region struct {
	ID   int
	Code string
	Name string
}

type Price struct {
	P25 int
	P50 int
	P75 int
}

type Breakdown struct {
	Base  int
	Items []BreakdownItem
}

type BreakdownItem struct {
	Code       string
	Value      string
	Title      string
	Multiplier float64
	Exact      bool
}
