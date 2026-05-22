package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"github.com/en7ka/hitalent_testovoe/internal/service"
)

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateDepartmentRequest struct {
	Name     *string      `json:"name"`
	ParentID NullableUint `json:"parent_id"`
}

type NullableUint struct {
	Set   bool
	Value *uint
}

func (n *NullableUint) UnmarshalJSON(data []byte) error {
	n.Set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var value uint
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("parent_id must be integer or null")
	}

	n.Value = &value
	return nil
}

type DepartmentResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ParentID  *uint     `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type DepartmentTreeResponse struct {
	Department DepartmentResponse       `json:"department"`
	Employees  []EmployeeResponse       `json:"employees,omitempty"`
	Children   []DepartmentTreeResponse `json:"children"`
}

func DepartmentFromDomain(department domain.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:        department.ID,
		Name:      department.Name,
		ParentID:  department.ParentID,
		CreatedAt: department.CreatedAt,
	}
}

func DepartmentTreeFromService(tree service.DepartmentTree) DepartmentTreeResponse {
	children := make([]DepartmentTreeResponse, 0, len(tree.Children))
	for _, child := range tree.Children {
		children = append(children, DepartmentTreeFromService(child))
	}

	return DepartmentTreeResponse{
		Department: DepartmentFromDomain(tree.Department),
		Employees:  EmployeesFromDomain(tree.Employees),
		Children:   children,
	}
}
