package service

import (
	"context"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"github.com/en7ka/hitalent_testovoe/internal/repository"
)

type CreateEmployeeInput struct {
	FullName string
	Position string
	HiredAt  *time.Time
}

type EmployeeService interface {
	Create(ctx context.Context, departmentID uint, input CreateEmployeeInput) (domain.Employee, error)
}

type employeeService struct {
	employees   repository.EmployeeRepository
	departments repository.DepartmentRepository
}

func NewEmployeeService(employees repository.EmployeeRepository, departments repository.DepartmentRepository) EmployeeService {
	return &employeeService{
		employees:   employees,
		departments: departments,
	}
}

func (s *employeeService) Create(ctx context.Context, departmentID uint, input CreateEmployeeInput) (domain.Employee, error) {
	if departmentID == 0 {
		return domain.Employee{}, ErrInvalidInput
	}

	fullName, err := normalizeRequiredString(input.FullName)
	if err != nil {
		return domain.Employee{}, err
	}

	position, err := normalizeRequiredString(input.Position)
	if err != nil {
		return domain.Employee{}, err
	}

	exists, err := s.departments.Exists(ctx, departmentID)
	if err != nil {
		return domain.Employee{}, err
	}
	if !exists {
		return domain.Employee{}, ErrNotFound
	}

	employee := domain.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      input.HiredAt,
	}
	if err := s.employees.Create(ctx, &employee); err != nil {
		return domain.Employee{}, err
	}

	return employee, nil
}
