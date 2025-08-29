package models

type Classroom struct {
	ID         int    `json:"id,omitempty"`
	RoomNumber string `json:"room_number,omitempty"`
	Building   string `json:"building,omitempty"`
	Capacity   int    `json:"capacity,omitempty"`
}

func (Classroom) SortableFields() map[string]string {
	return map[string]string{
		"id":          "id",
		"room_number": "room_number",
		"building":    "building",
		"capacity":    "capacity",
	}
}

func (Classroom) FilterableFields() map[string]string {
	return map[string]string{
		"id":          "id",
		"room_number": "room_number",
		"building":    "building",
		"capacity":    "capacity",
	}
}
