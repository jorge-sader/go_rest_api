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

// TeachersHandler handles all requests to /teachers/ using pre-Go 1.22 routing.
// This demonstrates compatibility with legacy codebases by manually handling HTTP methods and path parsing.
func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on '%s'", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		idStr := extractID(r)
		if idStr == "" {
			GetManyTeachersHandler(w, r)
		} else {
			GetOneTeacherHandler(w, r)
		}
	case http.MethodPost:
		AddManyTeachersHandler(w, r)
	case http.MethodPut:
		PutOneTeacherHandler(w, r)
	case http.MethodPatch:
		idStr := extractID(r)
		if idStr == "" {
			PatchManyTeachersHandler(w, r)
		} else {
			PatchOneTeacherHandler(w, r)
		}
	case http.MethodDelete:
		idStr := extractID(r)
		if idStr == "" {
			DeleteManyTeachersHandler(w, r)
		} else {
			DeleteOneTeacherHandler(w, r)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// extractID extracts the ID from the path (e.g., "/teachers/123" -> "123").
func extractID(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/teachers")
	path = strings.Trim(path, "/")
	return path
}

// GetManyTeachersHandler retrieves multiple teachers with filtering and sorting.
func GetManyTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers`
	var args []any

	query, args, err = addFilters(r, query, args, models.Teacher{})
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}
	query = addSorting(r, query, models.Teacher{})

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error querying teachers: %v", err)
		if err == sql.ErrNoRows {
			response := struct {
				Status string           `json:"status"`
				Count  int              `json:"count"`
				Data   []models.Teacher `json:"data"`
			}{
				Status: "success",
				Count:  0,
				Data:   []models.Teacher{},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		http.Error(w, "Error querying teachers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.ClassroomID, &teacher.SubjectID)
		if err != nil {
			log.Printf("Error scanning database results: %v", err)
			http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return
		}
		teacherList = append(teacherList, teacher)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating database results: %v", err)
		http.Error(w, "Error iterating database results", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teacherList),
		Data:   teacherList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOneTeacherHandler retrieves a single teacher by ID.
func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid teacher ID: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var teacher models.Teacher
	err = db.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers WHERE id = ?`, id).
		Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.ClassroomID, &teacher.SubjectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying teacher: %v", err)
		http.Error(w, "Error querying teacher", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string         `json:"status"`
		Data   models.Teacher `json:"data"`
	}{
		Status: "success",
		Data:   teacher,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AddManyTeachersHandler creates multiple teachers with transaction support.
func AddManyTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(newTeachers) == 0 {
		http.Error(w, "Empty teacher list", http.StatusBadRequest)
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

	stmt, err := tx.PrepareContext(r.Context(), `INSERT INTO teachers (first_name, last_name, email, classroom_id, subject_id) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Error preparing SQL statement: %v", err)
		http.Error(w, "Error preparing SQL statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, 0, len(newTeachers))
	for _, newTeacher := range newTeachers {
		if newTeacher.FirstName == "" || newTeacher.LastName == "" || newTeacher.Email == "" || newTeacher.ClassroomID == 0 || newTeacher.SubjectID == 0 {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		res, err := stmt.ExecContext(r.Context(), newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.ClassroomID, newTeacher.SubjectID)
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
		newTeacher.ID = int(lastID)
		addedTeachers = append(addedTeachers, newTeacher)
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
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

// PutOneTeacherHandler updates a teacher (full update) with transaction support.
func PutOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid teacher ID: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&updatedTeacher); err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
	// Uses defer tx.Rollback() to ensure transaction cleanup on errors,
	// with tx.Commit() called only on success. The rollback after a commit fails silently, which is safe.
	defer tx.Rollback()

	var existingTeacher models.Teacher
	err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers WHERE id = ?`, id).
		Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.ClassroomID, &existingTeacher.SubjectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying teacher: %v", err)
		http.Error(w, "Error querying teacher", http.StatusInternalServerError)
		return
	}

	updatedTeacher.ID = id
	_, err = tx.ExecContext(r.Context(), `UPDATE teachers SET first_name = ?, last_name = ?, email = ?, classroom_id = ?, subject_id = ? WHERE id = ?`,
		updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.ClassroomID, updatedTeacher.SubjectID, id)
	if err != nil {
		log.Printf("Error updating teacher: %v", err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
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
	if err := json.NewEncoder(w).Encode(updatedTeacher); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// PatchOneTeacherHandler partially updates a single teacher by ID with dynamic UPDATE query.
// Uses defer tx.Rollback() to ensure transaction cleanup on errors,
// with tx.Commit() called only on success. The rollback after a commit fails silently, which is safe.
func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid teacher ID: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate and build dynamic SET clause
	validFields := models.Teacher{}.FilterableFields()
	var setClauses []string
	var args []any
	for k, v := range updates {
		if field, ok := validFields[k]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
			args = append(args, v)
		}
	}
	if len(setClauses) == 0 {
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

	// Verify teacher exists
	var existingTeacher models.Teacher
	err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers WHERE id = ?`, id).
		Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.ClassroomID, &existingTeacher.SubjectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying teacher: %v", err)
		http.Error(w, "Error querying teacher", http.StatusInternalServerError)
		return
	}

	// Build and execute dynamic UPDATE query
	query := fmt.Sprintf(`UPDATE teachers SET %s WHERE id = ?`, strings.Join(setClauses, ", "))
	args = append(args, id)
	_, err = tx.ExecContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error updating teacher ID %d: %v", id, err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	// Fetch updated teacher
	updatedTeacher := models.Teacher{}
	err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers WHERE id = ?`, id).
		Scan(&updatedTeacher.ID, &updatedTeacher.FirstName, &updatedTeacher.LastName, &updatedTeacher.Email, &updatedTeacher.ClassroomID, &updatedTeacher.SubjectID)
	if err != nil {
		log.Printf("Error querying updated teacher: %v", err)
		http.Error(w, "Error querying updated teacher", http.StatusInternalServerError)
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
		Data   models.Teacher `json:"data"`
	}{
		Status: "success",
		Data:   updatedTeacher,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// PatchManyTeachersHandler partially updates multiple teachers based on filters with dynamic UPDATE query.
// Uses defer tx.Rollback() to ensure transaction cleanup on errors,
// with tx.Commit() called only on success. The rollback after a commit fails silently, which is safe.
func PatchManyTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate and build dynamic SET clause
	validFields := models.Teacher{}.FilterableFields()
	var setClauses []string
	var args []any
	for k, v := range updates {
		if field, ok := validFields[k]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
			args = append(args, v)
		}
	}
	if len(setClauses) == 0 {
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
	query := `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers`
	var selectArgs []any
	query, selectArgs, err = addFilters(r, query, selectArgs, models.Teacher{})
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}

	rows, err := tx.QueryContext(r.Context(), query, selectArgs...)
	if err != nil {
		log.Printf("Error querying teachers: %v", err)
		if err == sql.ErrNoRows {
			http.Error(w, "No teachers found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error querying teachers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.ClassroomID, &teacher.SubjectID)
		if err != nil {
			log.Printf("Error scanning database results: %v", err)
			http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return
		}
		teacherList = append(teacherList, teacher)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating database results: %v", err)
		http.Error(w, "Error iterating database results", http.StatusInternalServerError)
		return
	}

	if len(teacherList) == 0 {
		http.Error(w, "No teachers found to update", http.StatusNotFound)
		return
	}

	// Build and execute dynamic UPDATE query for all matching teachers
	updateQuery := fmt.Sprintf(`UPDATE teachers SET %s WHERE id = ?`, strings.Join(setClauses, ", "))
	updatedTeachers := make([]models.Teacher, 0, len(teacherList))
	for _, teacher := range teacherList {
		updateArgs := append(args, teacher.ID)
		_, err = tx.ExecContext(r.Context(), updateQuery, updateArgs...)
		if err != nil {
			log.Printf("Error updating teacher ID %d: %v", teacher.ID, err)
			http.Error(w, "Error updating teachers", http.StatusInternalServerError)
			return
		}

		// Fetch updated teacher
		updatedTeacher := models.Teacher{}
		err = tx.QueryRowContext(r.Context(), `SELECT id, first_name, last_name, email, classroom_id, subject_id FROM teachers WHERE id = ?`, teacher.ID).
			Scan(&updatedTeacher.ID, &updatedTeacher.FirstName, &updatedTeacher.LastName, &updatedTeacher.Email, &updatedTeacher.ClassroomID, &updatedTeacher.SubjectID)
		if err != nil {
			log.Printf("Error querying updated teacher ID %d: %v", teacher.ID, err)
			http.Error(w, "Error querying updated teachers", http.StatusInternalServerError)
			return
		}
		updatedTeachers = append(updatedTeachers, updatedTeacher)
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
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(updatedTeachers),
		Data:   updatedTeachers,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// DeleteOneTeacherHandler deletes a single teacher by ID.
func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid teacher ID: %v", err)
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
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

	result, err := tx.ExecContext(r.Context(), `DELETE FROM teachers WHERE id = ?`, id)
	if err != nil {
		log.Printf("Error deleting teacher: %v", err)
		http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving delete result: %v", err)
		http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Teacher not found", http.StatusNotFound)
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
		Status: "Teacher successfully deleted",
		ID:     id,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// DeleteManyTeachersHandler deletes multiple teachers based on filters.
func DeleteManyTeachersHandler(w http.ResponseWriter, r *http.Request) {
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
	query := `DELETE FROM teachers`
	var args []any
	query, args, err = addFilters(r, query, args, models.Teacher{})
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}

	result, err := tx.ExecContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error deleting teachers: %v", err)
		http.Error(w, "Error deleting teachers", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving delete result: %v", err)
		http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No teachers found to delete", http.StatusNotFound)
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
