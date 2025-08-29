package router

import (
	"net/http"

	"github.com/jorge-sader/go-rest-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", handlers.RootHandler)

	// TEACHERS
	// INFO: I'm knowingly using pre Go 1.22 routing method for teachers as lots of legacy code still uses it.
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)

	//STUDENTS
	mux.HandleFunc("GET /students/", handlers.GetManyStudentsHandler)
	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentHandler)
	mux.HandleFunc("POST /students/", handlers.AddManyStudentsHandler)
	mux.HandleFunc("PATCH /students/", handlers.PatchManyStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentHandler)
	mux.HandleFunc("DELETE /students/", handlers.DeleteManyStudentsHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudentHandler)

	//EXECS
	// mux.HandleFunc("GET /executives/", handlers.GetManyExecutivesHandler)
	// mux.HandleFunc("GET /executives/{id}", handlers.GetOneExecutiveHandler)
	// mux.HandleFunc("POST /executives/", handlers.AddManyExecutivesHandler)
	// mux.HandleFunc("PATCH /executives/", handlers.PatchManyExecutivesHandler)
	// mux.HandleFunc("PATCH /executives/{id}", handlers.PatchOneExecutiveHandler)
	// mux.HandleFunc("DELETE /executives/", handlers.DeleteManyExecutivesHandler)
	// mux.HandleFunc("DELETE /executives/{id}", handlers.DeleteOneExecutiveHandler)

	// CLASSROOMS
	// mux.HandleFunc("GET /classrooms/", handlers.GetManyClassroomsHandler)
	// mux.HandleFunc("GET /classrooms/{id}", handlers.GetOneClassroomHandler)
	// mux.HandleFunc("POST /classrooms/", handlers.AddManyClassroomsHandler)
	// mux.HandleFunc("PATCH /classrooms/", handlers.PatchManyClassroomsHandler)
	// mux.HandleFunc("PATCH /classrooms/{id}", handlers.PatchOneClassroomHandler)
	// mux.HandleFunc("DELETE /classrooms/", handlers.DeleteManyClassroomHandler)
	// mux.HandleFunc("DELETE /classrooms/{id}", handlers.DeleteOneClassroomHandler)

	// SUBJECTS
	// mux.HandleFunc("GET /subjects/", handlers.GetManySubjectsHandler)
	// mux.HandleFunc("GET /subjects/{id}", handlers.GetOneSubjectHandler)
	// mux.HandleFunc("POST /subjects/", handlers.AddManySubjectsHandler)
	// mux.HandleFunc("PATCH /subjects/", handlers.PatchManySubjectsHandler)
	// mux.HandleFunc("PATCH /subjects/{id}", handlers.PatchOneSubjectHandler)
	// mux.HandleFunc("DELETE /subjects/", handlers.DeleteManySubjectHandler)
	// mux.HandleFunc("DELETE /subjects/{id}", handlers.DeleteOneSubjectHandler)
	return mux
}
