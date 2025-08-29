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
	"github.com/jorge-sader/go-rest-api/pkg/responder"
)

// addSorting appends an ORDER BY clause based on the model's sortable fields.
func addSorting(r *http.Request, query string, model models.Model) string {
	validFields := model.SortableFields()
	validOrders := map[string]bool{"asc": true, "desc": true}

	sortParams := r.URL.Query()["sort_by"]
	if len(sortParams) == 0 {
		return query
	}

	query += " ORDER BY"
	for i, param := range sortParams {
		parts := strings.Split(param, ":")
		if len(parts) != 2 {
			continue
		}
		field, order := parts[0], strings.ToLower(parts[1])
		dbField, fieldValid := validFields[field]
		if !fieldValid || !validOrders[order] {
			continue
		}
		if i > 0 {
			query += ","
		}
		query += " " + dbField + " " + order
	}
	return query
}

// addFilters appends WHERE clauses to the query based on query parameters and the model's FilterableFields.
// It adds a WHERE clause for the first filter and AND for subsequent filters.
// Returns an error if no valid filters are provided to prevent unintended operations.
func addFilters(r *http.Request, query string, args []any, model interface{}) (string, []any, error) {
	validFields := model.(models.Model).FilterableFields()
	firstFilter := true

	for key, values := range r.URL.Query() {
		if dbField, ok := validFields[key]; ok {
			if key == "id" && len(values) > 1 {
				// Handle multiple ID values with IN clause
				validIDs := make([]string, 0, len(values))
				idArgs := make([]any, 0, len(values))
				for _, v := range values {
					if id, err := strconv.Atoi(v); err == nil && id > 0 {
						validIDs = append(validIDs, "?")
						idArgs = append(idArgs, id)
					}
				}
				if len(validIDs) > 0 {
					if firstFilter {
						query += fmt.Sprintf(" WHERE %s IN (%s)", dbField, strings.Join(validIDs, ", "))
						firstFilter = false
					} else {
						query += fmt.Sprintf(" AND %s IN (%s)", dbField, strings.Join(validIDs, ", "))
					}
					args = append(args, idArgs...)
				}
			} else if len(values) > 0 {
				// Handle single value for other fields
				if firstFilter {
					query += fmt.Sprintf(" WHERE %s = ?", dbField)
					firstFilter = false
				} else {
					query += fmt.Sprintf(" AND %s = ?", dbField)
				}
				args = append(args, values[0])
			}
		}
	}

	if firstFilter {
		// No valid filters provided
		return query, args, fmt.Errorf("no valid filters provided")
	}

	return query, args, nil
}

// GetManyHandler is a generic handler for retrieving multiple records of any model.
func GetManyHandler[T models.Model](w http.ResponseWriter, r *http.Request, table string, scanFunc func(*T, *sql.Rows) error) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `SELECT * FROM ` + table
	var args []any

	var model T
	query, args, err = addFilters(r, query, args, model)
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}
	query = addSorting(r, query, model)

	rows, err := db.QueryContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error querying %s: %v", table, err)
		http.Error(w, "Error querying data.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := make([]T, 0)
	for rows.Next() {
		var item T
		err := scanFunc(&item, rows)
		if err != nil {
			log.Printf("Error scanning database results: %v", err)
			http.Error(w, "Error scanning database results.", http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over database results: %v", err)
		http.Error(w, "Error iterating over database results.", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
		Data   []T    `json:"data"`
	}{
		Status: "success",
		Count:  len(list),
		Data:   list,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOneHandler is a generic handler for retrieving a single record of any model.
func GetOneHandler[T models.Model](w http.ResponseWriter, r *http.Request, table string, scanFunc func(*T, *sql.Row) error) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `SELECT * FROM ` + table
	var args []any

	var model T
	query, args, err = addFilters(r, query, args, model)
	if err != nil {
		log.Printf("Invalid request: %v", err)
		http.Error(w, "At least one valid filter is required", http.StatusBadRequest)
		return
	}

	// Limit to one record
	query += " LIMIT 1"

	row := db.QueryRowContext(r.Context(), query, args...)
	var item T
	err = scanFunc(&item, row)
	if err == sql.ErrNoRows {
		responder.RespondNoRecordFound(w)
		return
	} else if err != nil {
		log.Printf("Error scanning %s: %v", table, err)
		http.Error(w, "Error retrieving data.", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string `json:"status"`
		Data   T      `json:"data"`
	}{
		Status: "success",
		Data:   item,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// func addSorting(r *http.Request, query string) string {
// 	sortParams := r.URL.Query()["sort_by"]
// 	if len(sortParams) > 0 {
// 		query += " ORDER BY"
// 		for i, param := range sortParams {
// 			parts := strings.Split(param, ":")
// 			if len(parts) != 2 {
// 				continue
// 			}
// 			field, order := parts[0], parts[1]
// 			if !isValidSortFiled(field) || !isValidSortOrder(order) {
// 				continue
// 			}
// 			if i > 0 {
// 				query += ","
// 			}
// 			query += " " + field + " " + order
// 		}
// 	}
// 	return query
// }
//
// // TODO: Could we not make this generic? (accept a Model/struct with JSON and DB tags and infer params via reflection?)
// func addFilters(r *http.Request, query string, args []any) (string, []any) {
// 	params := map[string]string{
// 		"first_name":   "first_name",
// 		"last_name":    "last_name",
// 		"email":        "email",
// 		"classroom_id": "classroom_id",
// 		"subject_id":   "subject_id",
// 	}
//
// 	for param, dbField := range params {
// 		value := r.URL.Query().Get(param)
// 		if value != "" {
// 			query += " AND " + dbField + " = ?"
// 			args = append(args, value)
//
// 		}
// 	}
// 	return query, args
// }
