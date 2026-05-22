package http

import (
	"net/http"

	"github.com/en7ka/hitalent_testovoe/internal/service"
)

func NewRouter(departments service.DepartmentService, employees service.EmployeeService) http.Handler {
	departmentHandler := NewDepartmentHandler(departments)
	employeeHandler := NewEmployeeHandler(employees)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":    "hitalent_testovoe",
			"status":  "ok",
			"message": "Organization structure API",
			"endpoints": []string{
				"GET /health",
				"POST /departments/",
				"POST /departments/{id}/employees/",
				"GET /departments/{id}?depth=1&include_employees=true",
				"PATCH /departments/{id}",
				"DELETE /departments/{id}?mode=cascade",
				"DELETE /departments/{id}?mode=reassign&reassign_to_department_id=2",
			},
		})
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /departments", departmentHandler.Create)
	mux.HandleFunc("POST /departments/", departmentHandler.Create)
	mux.HandleFunc("GET /departments/{id}", departmentHandler.GetByID)
	mux.HandleFunc("PATCH /departments/{id}", departmentHandler.Update)
	mux.HandleFunc("DELETE /departments/{id}", departmentHandler.Delete)

	mux.HandleFunc("POST /departments/{id}/employees", employeeHandler.Create)
	mux.HandleFunc("POST /departments/{id}/employees/", employeeHandler.Create)

	return recoverMiddleware(loggingMiddleware(mux))
}
