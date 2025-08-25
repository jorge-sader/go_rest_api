package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/jorge-sader/go-rest-api/internal/api/repositories/sqlconnect"
	"github.com/jorge-sader/go-rest-api/internal/models"
)

var teachers = make(map[int]models.Teacher)
var mutex = &sync.Mutex{}
var nextID = 1

// Initialize/mock some data
func init() {
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "John",
		LastName:  "Doe",
		Classroom: "9A",
		Subject:   "Math",
	}
	nextID++

	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Doe",
		Classroom: "10A",
		Subject:   "Algebra",
	}
	nextID++

	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Smith",
		Classroom: "11A",
		Subject:   "Calculus",
	}
	nextID++
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		addTeacherHandler(w, r)

	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		getTeachersHandler(w, r)

	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	}
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	if idStr == "" {
		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		// teacherList := []Teacher{}
		teacherList := make([]models.Teacher, 0, len(teachers))

		for _, teacher := range teachers {
			// Apply filters
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
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
