package models

type Teacher struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Classroom string `json:"classroom,omitempty"`
	Subject   string `json:"subject,omitempty"`
}
