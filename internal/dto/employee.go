package dto

import (
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
)

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

type EmployeeResponse struct {
	ID           uint      `json:"id"`
	DepartmentID uint      `json:"department_id"`
	FullName     string    `json:"full_name"`
	Position     string    `json:"position"`
	HiredAt      *string   `json:"hired_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func EmployeeFromDomain(employee domain.Employee) EmployeeResponse {
	var hiredAt *string
	if employee.HiredAt != nil {
		value := employee.HiredAt.Format(time.DateOnly)
		hiredAt = &value
	}

	return EmployeeResponse{
		ID:           employee.ID,
		DepartmentID: employee.DepartmentID,
		FullName:     employee.FullName,
		Position:     employee.Position,
		HiredAt:      hiredAt,
		CreatedAt:    employee.CreatedAt,
	}
}

func EmployeesFromDomain(employees []domain.Employee) []EmployeeResponse {
	response := make([]EmployeeResponse, 0, len(employees))
	for _, employee := range employees {
		response = append(response, EmployeeFromDomain(employee))
	}

	return response
}
