package sqlconnect

import (
	"database/sql"
	"log"

	"github.com/jorge-sader/go-rest-api/internal/models"
)

// GetStudentById fetches a student by ID
func GetStudentById(db *sql.DB, id int) (models.Student, error) {
	var student models.Student
	err := db.QueryRow(`SELECT id, first_name, last_name, email, classroom_id FROM students WHERE id = ?`, id).
		Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.ClassroomID)
	if err != nil {
		log.Printf("Error fetching student with id %d: %v", id, err)
		return models.Student{}, err
	}
	return student, nil
}

// GetStudentByEmail fetches a student by email
func GetStudentByEmail(db *sql.DB, email string) (models.Student, error) {
	var student models.Student
	err := db.QueryRow(`SELECT id, first_name, last_name, email, classroom_id FROM students WHERE email = ?`, email).
		Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.ClassroomID)
	if err != nil {
		log.Printf("Error fetching student with email %s: %v", email, err)
		return models.Student{}, err
	}
	return student, nil
}
