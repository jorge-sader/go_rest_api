package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jorge-sader/go-rest-api/internal/api/repositories/sqlconnect"
	"github.com/jorge-sader/go-rest-api/internal/models"
)

// GetManyStudentsHandler retrieves multiple students with filtering and sorting.
func GetManyStudentsHandler(w http.ResponseWriter, r *http.Request) {
	scanFunc := func(s *models.Student, rows *sql.Rows) error {
		return rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.ClassroomID)
	}
	GetManyHandler(w, r, "students", scanFunc)
}

// GetOneStudentHandler retrieves a single student by ID.
func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		log.Printf("Invalid student ID: %v", err)
		http.Error(w, "Invalid student id", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id FROM students WHERE id = ?`, id).
		Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.ClassroomID)
	if err == sql.ErrNoRows {
		log.Printf("Student not found: id=%d", id)
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying student: %v", err)
		http.Error(w, "Error querying student", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string         `json:"status"`
		Data   models.Student `json:"data"`
	}{
		Status: "success",
		Data:   student,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// AddManyStudentsHandler creates multiple students with transaction support.
func AddManyStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newStudents []models.Student
	err = json.NewDecoder(r.Body).Decode(&newStudents)
	if err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(newStudents) == 0 {
		log.Printf("Empty student list")
		http.Error(w, "Empty student list", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(r.Context(), `INSERT INTO students (first_name, last_name, email, classroom_id) VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Error preparing SQL statement: %v", err)
		http.Error(w, "Error preparing SQL statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, 0, len(newStudents))
	for _, newStudent := range newStudents {
		if newStudent.FirstName == "" || newStudent.LastName == "" || newStudent.Email == "" || newStudent.ClassroomID == 0 {
			log.Printf("Missing required fields for student: %+v", newStudent)
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		res, err := stmt.ExecContext(r.Context(), newStudent.FirstName, newStudent.LastName, newStudent.Email, newStudent.ClassroomID)
		if err != nil {
			log.Printf("Error executing SQL statement: %v", err)
			http.Error(w, "Error executing SQL statement", http.StatusInternalServerError)
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			log.Printf("Error retrieving last inserted ID: %v", err)
			http.Error(w, "Error retrieving last inserted ID", http.StatusInternalServerError)
			return
		}
		newStudent.ID = int(lastID)
		addedStudents = append(addedStudents, newStudent)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// PatchOneStudentHandler partially updates a single student by ID with dynamic UPDATE query.
// Uses defer tx.Rollback() to ensure transaction cleanup on errors,
// with tx.Commit() called only on success. The rollback after a commit fails silently, which is safe.
func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		log.Printf("Invalid student ID: %v", err)
		http.Error(w, "Invalid student id", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate and build dynamic SET clause
	validFields := models.Student{}.FilterableFields()
	var setClauses []string
	var args []any
	for k, v := range updates {
		if field, ok := validFields[k]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
			args = append(args, v)
		}
	}
	if len(setClauses) == 0 {
		log.Printf("No valid fields to update for student ID %d", id)
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Start transaction
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify student exists
	var existingStudent models.Student
	err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id FROM students WHERE id = ?`, id).
		Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.ClassroomID)
	if err == sql.ErrNoRows {
		log.Printf("Student not found: id=%d", id)
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying student: %v", err)
		http.Error(w, "Error querying student", http.StatusInternalServerError)
		return
	}

	// Build and execute dynamic UPDATE query
	query := fmt.Sprintf(`UPDATE students SET %s WHERE id = ?`, strings.Join(setClauses, ", "))
	args = append(args, id)
	_, err = tx.ExecContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error updating student ID %d: %v", id, err)
		http.Error(w, "Error updating student", http.StatusInternalServerError)
		return
	}

	// Fetch updated student
	updatedStudent := models.Student{}
	err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id FROM students WHERE id = ?`, id).
		Scan(&updatedStudent.ID, &updatedStudent.FirstName, &updatedStudent.LastName, &updatedStudent.Email, &updatedStudent.ClassroomID)
	if err != nil {
		log.Printf("Error querying updated student: %v", err)
		http.Error(w, "Error querying updated student", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string         `json:"status"`
		Data   models.Student `json:"data"`
	}{
		Status: "success",
		Data:   updatedStudent,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// PatchManyStudentsHandler partially updates multiple students based on filters with dynamic UPDATE query.
// Uses defer tx.Rollback() to ensure transaction cleanup on errors,
// with tx.Commit() called only on success. The rollback after a commit fails silently, which is safe.
func PatchManyStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate and build dynamic SET clause
	validFields := models.Student{}.FilterableFields()
	var setClauses []string
	var args []any
	for k, v := range updates {
		if field, ok := validFields[k]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
			args = append(args, v)
		}
	}
	if len(setClauses) == 0 {
		log.Printf("No valid fields to update")
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Start transaction
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Build WHERE clause with filters
	query := `SELECT id, first_name, last_name, email, classroom_id FROM students`
	var selectArgs []any
	query, selectArgs, err = addFilters(r, query, selectArgs, models.Student{})
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}

	rows, err := tx.QueryContext(r.Context(), query, selectArgs...)
	if err != nil {
		log.Printf("Error querying students: %v", err)
		if err == sql.ErrNoRows {
			http.Error(w, "No students found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error querying students", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	studentList := make([]models.Student, 0)
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.ClassroomID)
		if err != nil {
			log.Printf("Error scanning database results: %v", err)
			http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return
		}
		studentList = append(studentList, student)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating database results: %v", err)
		http.Error(w, "Error iterating database results", http.StatusInternalServerError)
		return
	}

	if len(studentList) == 0 {
		log.Printf("No students found to update")
		http.Error(w, "No students found to update", http.StatusNotFound)
		return
	}

	// Build and execute dynamic UPDATE query for all matching students
	updateQuery := fmt.Sprintf(`UPDATE students SET %s WHERE id = ?`, strings.Join(setClauses, ", "))
	updatedStudents := make([]models.Student, 0, len(studentList))
	for _, student := range studentList {
		updateArgs := append(args, student.ID)
		_, err = tx.ExecContext(r.Context(), updateQuery, updateArgs...)
		if err != nil {
			log.Printf("Error updating student ID %d: %v", student.ID, err)
			http.Error(w, "Error updating students", http.StatusInternalServerError)
			return
		}

		// Fetch updated student
		updatedStudent := models.Student{}
		err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id FROM students WHERE id = ?`, student.ID).
			Scan(&updatedStudent.ID, &updatedStudent.FirstName, &updatedStudent.LastName, &updatedStudent.Email, &updatedStudent.ClassroomID)
		if err != nil {
			log.Printf("Error querying updated student ID %d: %v", student.ID, err)
			http.Error(w, "Error querying updated students", http.StatusInternalServerError)
			return
		}
		updatedStudents = append(updatedStudents, updatedStudent)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(updatedStudents),
		Data:   updatedStudents,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// DeleteOneStudentHandler deletes a single student by ID.
func DeleteOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		log.Printf("Invalid student ID: %v", err)
		http.Error(w, "Invalid student id", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Start transaction
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(r.Context(), `DELETE FROM students WHERE id = ?`, id)
	if err != nil {
		log.Printf("Error deleting student: %v", err)
		http.Error(w, "Error deleting student", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving delete result: %v", err)
		http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("Student not found: id=%d", id)
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "success",
		ID:     id,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// DeleteManyStudentsHandler deletes multiple students based on filters.
func DeleteManyStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Start transaction
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Build DELETE query with filters
	query := `DELETE FROM students`
	var args []any
	query, args, err = addFilters(r, query, args, models.Student{})
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}

	result, err := tx.ExecContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error deleting students: %v", err)
		http.Error(w, "Error deleting students", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving delete result: %v", err)
		http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No students found to delete")
		http.Error(w, "No students found to delete", http.StatusNotFound)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status       string `json:"status"`
		CountDeleted int64  `json:"count_deleted"`
	}{
		Status:       "success",
		CountDeleted: rowsAffected,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
