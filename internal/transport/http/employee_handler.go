package http

import (
	"net/http"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/dto"
	"github.com/en7ka/hitalent_testovoe/internal/service"
)

type EmployeeHandler struct {
	employees service.EmployeeService
}

func NewEmployeeHandler(employees service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employees: employees}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	departmentID, ok := parsePathID(w, r)
	if !ok {
		return
	}

	var request dto.CreateEmployeeRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	hiredAt, ok := parseOptionalDate(w, request.HiredAt)
	if !ok {
		return
	}

	employee, err := h.employees.Create(r.Context(), departmentID, service.CreateEmployeeInput{
		FullName: request.FullName,
		Position: request.Position,
		HiredAt:  hiredAt,
	})
	if handleServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusCreated, dto.EmployeeFromDomain(employee))
}

func parseOptionalDate(w http.ResponseWriter, raw *string) (*time.Time, bool) {
	if raw == nil || *raw == "" {
		return nil, true
	}

	value, err := time.Parse(time.DateOnly, *raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hired_at")
		return nil, false
	}

	return &value, true
}
