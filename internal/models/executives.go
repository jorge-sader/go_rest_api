package models

type Executive struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Role      string `json:"role,omitempty"`
}

func (Executive) SortableFields() map[string]string {
	return map[string]string{
		"id":         "id",
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"username":   "username",
		"password":   "password",
		"role":       "role",
	}
}

func (Executive) FilterableFields() map[string]string {
	return map[string]string{
		"id":         "id",
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"username":   "username",
		"password":   "password",
		"role":       "role",
	}
}
