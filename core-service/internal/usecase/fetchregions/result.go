package fetchregions

// Result - результат выборки регионов
type Result struct {
	Items []Region
}

type Region struct {
	ID    int
	Name  string
	Codes []string
}
