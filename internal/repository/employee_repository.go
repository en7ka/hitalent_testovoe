package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"gorm.io/gorm"
)

type EmployeeRepository interface {
	Create(ctx context.Context, employee *domain.Employee) error
	ListByDepartmentID(ctx context.Context, departmentID uint) ([]domain.Employee, error)
	ReassignDepartments(ctx context.Context, departmentIDs []uint, targetDepartmentID uint) error
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *domain.Employee) error {
	if err := r.db.WithContext(ctx).Create(employee).Error; err != nil {
		return fmt.Errorf("create employee: %w", err)
	}

	return nil
}

func (r *employeeRepository) ListByDepartmentID(ctx context.Context, departmentID uint) ([]domain.Employee, error) {
	var employees []domain.Employee
	err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("created_at ASC, full_name ASC, id ASC").
		Find(&employees).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []domain.Employee{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}

	return employees, nil
}

func (r *employeeRepository) ReassignDepartments(ctx context.Context, departmentIDs []uint, targetDepartmentID uint) error {
	if len(departmentIDs) == 0 {
		return nil
	}

	err := r.db.WithContext(ctx).
		Model(&domain.Employee{}).
		Where("department_id IN ?", departmentIDs).
		Update("department_id", targetDepartmentID).Error
	if err != nil {
		return fmt.Errorf("reassign employees: %w", err)
	}

	return nil
}
