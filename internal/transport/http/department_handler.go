package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/en7ka/hitalent_testovoe/internal/dto"
	"github.com/en7ka/hitalent_testovoe/internal/service"
)

type DepartmentHandler struct {
	departments service.DepartmentService
}

func NewDepartmentHandler(departments service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{departments: departments}
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateDepartmentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	department, err := h.departments.Create(r.Context(), service.CreateDepartmentInput{
		Name:     request.Name,
		ParentID: request.ParentID,
	})
	if handleServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusCreated, dto.DepartmentFromDomain(department))
}

func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r)
	if !ok {
		return
	}

	depth, ok := parseDepth(w, r)
	if !ok {
		return
	}

	includeEmployees, ok := parseIncludeEmployees(w, r)
	if !ok {
		return
	}

	tree, err := h.departments.GetTree(r.Context(), id, depth, includeEmployees)
	if handleServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusOK, dto.DepartmentTreeFromService(tree))
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r)
	if !ok {
		return
	}

	var request dto.UpdateDepartmentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	department, err := h.departments.Update(r.Context(), id, service.UpdateDepartmentInput{
		Name:        request.Name,
		ParentIDSet: request.ParentID.Set,
		ParentID:    request.ParentID.Value,
	})
	if handleServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusOK, dto.DepartmentFromDomain(department))
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(w, r)
	if !ok {
		return
	}

	mode := r.URL.Query().Get("mode")
	reassignToDepartmentID, ok := parseOptionalUintQuery(w, r, "reassign_to_department_id")
	if !ok {
		return
	}

	if err := h.departments.Delete(r.Context(), id, mode, reassignToDepartmentID); handleServiceError(w, err) {
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func parsePathID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}

	return uint(id), true
}

func parseDepth(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("depth")
	if raw == "" {
		return 1, true
	}

	depth, err := strconv.Atoi(raw)
	if err != nil || depth < 0 || depth > 5 {
		writeError(w, http.StatusBadRequest, "invalid depth")
		return 0, false
	}

	return depth, true
}

func parseIncludeEmployees(w http.ResponseWriter, r *http.Request) (bool, bool) {
	raw := r.URL.Query().Get("include_employees")
	if raw == "" {
		return true, true
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid include_employees")
		return false, false
	}

	return value, true
}

func parseOptionalUintQuery(w http.ResponseWriter, r *http.Request, key string) (*uint, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, true
	}

	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		writeError(w, http.StatusBadRequest, "invalid "+key)
		return nil, false
	}

	id := uint(value)
	return &id, true
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}
