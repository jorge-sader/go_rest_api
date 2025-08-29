package models

// Model defines methods for providing sortable and filterable fields.
type Model interface {
	// SortableFields returns a map of query parameter names to database column names for sorting.
	SortableFields() map[string]string

	// FilterableFields returns a map of query parameter names to database column names for filtering.
	FilterableFields() map[string]string
}
