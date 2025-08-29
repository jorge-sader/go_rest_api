package models

type Subject struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	TotalHours  string `json:"total_hours,omitempty"`
}

func (Subject) SortableFields() map[string]string {
	return map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"total_hours": "total_hours",
	}
}

func (Subject) FilterableFields() map[string]string {
	return map[string]string{
		"id":          "id",
		"name":        "name",
		"description": "description",
		"total_hours": "total_hours",
	}
}
