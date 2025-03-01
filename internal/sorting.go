package internal

type SortOption struct {
	// name of the field in the
	FieldName   string
	Desc        bool
	AsNumber    bool
	NumberRadix int
	AsTime      bool
	TimeFormat  string
}
