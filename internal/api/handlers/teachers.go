package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/jorge-sader/go-rest-api/internal/api/repositories/sqlconnect"
	"github.com/jorge-sader/go-rest-api/internal/models"
)

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		// POST /teachers/
		addTeacherHandler(w, r)

	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		// GET /teachers/{id?}
		getTeachersHandler(w, r)

	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		// PUT /teachers/{id}
		updateTeacherHandler(w, r)

	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		// PATCH /teachers/{id}
		patchTeacherHandler(w, r)

	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		// DELETE /teachers/{id}
	}
}

func isValidSortOrder(sortOrder string) bool {
	return sortOrder == "asc" || sortOrder == "desc"
}

// TODO: what happens if sortField is not on the map?
func isValidSortFiled(sortField string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"classroom":  true,
		"subject":    true,
	}
	return validFields[sortField]
}

// GET /teachers/{id?}
func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	idStr := extractID(r)

	if idStr == "" {
		query := `SELECT id, first_name, last_name, email, classroom, subject FROM teachers WHERE 1 = 1`
		var args []any

		query, args = addFilters(r, query, args)

		query = addSorting(r, query)

		rows, err := db.Query(query, args...)
		if err == sql.ErrNoRows {
			http.Error(w, "No matches found.", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Error querying teachers.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// teacherList := []Teacher{}
		teacherList := make([]models.Teacher, 0)
		for rows.Next() {
			var teacher models.Teacher
			err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Classroom, &teacher.Subject)
			if err != nil {
				http.Error(w, "Error scanning database results.", http.StatusInternalServerError)
				return
			}
			teacherList = append(teacherList, teacher)
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
		json.NewEncoder(w).Encode(response)
	}

	// Handle path Parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	err = db.QueryRow(`SELECT first_name, last_name, email, classroom, subject FROM teachers WHERE id = ?`, id).Scan(&teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Classroom, &teacher.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error Querying teachers.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func extractID(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	return idStr
}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sort_by"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortFiled(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

// TODO: Could we not make this generic? (accept a Model/struct with JSON and DB tags and infer params via reflection?)
func addFilters(r *http.Request, query string, args []any) (string, []any) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"classroom":  "classroom",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)

		}
	}
	return query, args
}

// POST /teachers/
func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare(`INSERT INTO teachers (first_name, last_name, email, classroom, subject) VALUES(?,?,?,?,?);`)
	if err != nil {
		http.Error(w, "error preparing SQL statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Classroom, newTeacher.Subject)
		if err != nil {
			http.Error(w, "error executing the SQL statement", http.StatusInternalServerError)
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "error retrieving last inserted ID", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
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

// PUT /teachers/{id}
func updateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher id", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow(`SELECT id, first_name, last_name, email, classroom, subject FROM teachers WHERE id = ?`, id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Classroom, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	updatedTeacher.ID = existingTeacher.ID
	_, err = db.Exec(`UPDATE teachers SET first_name = ?, last_name = ?, email = ?, classroom = ?, subject = ? WHERE id = ?`, updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Classroom, updatedTeacher.Subject, updatedTeacher.ID)
	if err != nil {
		http.Error(w, "Error updating Teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updatedTeacher)
	if err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
		return
	}
}

// PATCH /teachers/{id}
func patchTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := extractID(r)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher id", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow(`SELECT id, first_name, last_name, email, classroom, subject FROM teachers WHERE id = ?`, id).Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Classroom, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	// TODO: Could we not make this generic? (accept a Model/struct with JSON and DB tags and infer params via reflection?)
	updatedTeacher := existingTeacher
	for k, v := range updates {
		switch k {
		case "first_name":
			updatedTeacher.FirstName = v.(string)
		case "last_name":
			updatedTeacher.LastName = v.(string)
		case "email":
			updatedTeacher.Email = v.(string)
		case "classroom":
			updatedTeacher.Classroom = v.(string)
		case "subject":
			updatedTeacher.Subject = v.(string)
		}
	}

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			jsonTag := teacherType.Field(i).Tag.Get("json")
			field := strings.Split(jsonTag, ",")[0]
			if field == k && teacherVal.Field(i).CanSet() {
				teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
			}
		}
	}

	_, err = db.Exec(`UPDATE teachers SET first_name = ?, last_name = ?, email = ?, classroom = ?, subject = ? WHERE id = ?`, updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Classroom, updatedTeacher.Subject, updatedTeacher.ID)
	if err != nil {
		http.Error(w, "Error updating Teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updatedTeacher)
	if err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
		return
	}
}
