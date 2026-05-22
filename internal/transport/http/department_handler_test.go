package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"github.com/en7ka/hitalent_testovoe/internal/service"
)

type fakeDepartmentService struct{}

func (fakeDepartmentService) Create(ctx context.Context, input service.CreateDepartmentInput) (domain.Department, error) {
	return domain.Department{
		ID:        1,
		Name:      "Backend",
		ParentID:  input.ParentID,
		CreatedAt: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
	}, nil
}

func (fakeDepartmentService) GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (service.DepartmentTree, error) {
	return service.DepartmentTree{}, nil
}

func (fakeDepartmentService) Update(ctx context.Context, id uint, input service.UpdateDepartmentInput) (domain.Department, error) {
	return domain.Department{}, nil
}

func (fakeDepartmentService) Delete(ctx context.Context, id uint, mode string, reassignToDepartmentID *uint) error {
	return nil
}

type fakeEmployeeService struct{}

func (fakeEmployeeService) Create(ctx context.Context, departmentID uint, input service.CreateEmployeeInput) (domain.Employee, error) {
	return domain.Employee{}, nil
}

func TestCreateDepartment(t *testing.T) {
	router := NewRouter(fakeDepartmentService{}, fakeEmployeeService{})
	request := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"Backend"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	if !bytes.Contains(response.Body.Bytes(), []byte(`"name":"Backend"`)) {
		t.Fatalf("expected response to contain department name, got %s", response.Body.String())
	}
}
