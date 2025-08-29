package models

type Student struct {
	ID          int    `json:"id,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	ClassroomID int    `json:"classroom_id,omitempty"`
}

func (Student) SortableFields() map[string]string {
	return map[string]string{
		"id":           "id",
		"first_name":   "first_name",
		"last_name":    "last_name",
		"email":        "email",
		"classroom_id": "classroom_id",
	}
}

func (Student) FilterableFields() map[string]string {
	return map[string]string{
		"id":           "id",
		"first_name":   "first_name",
		"last_name":    "last_name",
		"email":        "email",
		"classroom_id": "classroom_id",
	}
}
