package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type DepartmentRepository interface {
	Create(ctx context.Context, department *domain.Department) error
	GetByID(ctx context.Context, id uint) (domain.Department, error)
	Update(ctx context.Context, department *domain.Department) error
	Delete(ctx context.Context, id uint) error
	Exists(ctx context.Context, id uint) (bool, error)
	NameExists(ctx context.Context, name string, parentID *uint, excludeID *uint) (bool, error)
	ListChildren(ctx context.Context, parentID uint) ([]domain.Department, error)
	GetDescendantIDs(ctx context.Context, id uint) ([]uint, error)
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, departments DepartmentRepository, employees EmployeeRepository) error) error
}

type departmentRepository struct {
	db *gorm.DB
}

type gormTransactor struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func NewTransactor(db *gorm.DB) Transactor {
	return &gormTransactor{db: db}
}

func (t *gormTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context, departments DepartmentRepository, employees EmployeeRepository) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, NewDepartmentRepository(tx), NewEmployeeRepository(tx))
	})
}

func (r *departmentRepository) Create(ctx context.Context, department *domain.Department) error {
	if err := r.db.WithContext(ctx).Create(department).Error; err != nil {
		return fmt.Errorf("create department: %w", err)
	}

	return nil
}

func (r *departmentRepository) GetByID(ctx context.Context, id uint) (domain.Department, error) {
	var department domain.Department
	err := r.db.WithContext(ctx).First(&department, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Department{}, ErrNotFound
	}
	if err != nil {
		return domain.Department{}, fmt.Errorf("get department: %w", err)
	}

	return department, nil
}

func (r *departmentRepository) Update(ctx context.Context, department *domain.Department) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Department{}).
		Where("id = ?", department.ID).
		Updates(map[string]any{
			"name":      department.Name,
			"parent_id": department.ParentID,
		})
	if result.Error != nil {
		return fmt.Errorf("update department: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return r.db.WithContext(ctx).First(department, department.ID).Error
}

func (r *departmentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.Department{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete department: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *departmentRepository) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Department{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check department exists: %w", err)
	}

	return count > 0, nil
}

func (r *departmentRepository) NameExists(ctx context.Context, name string, parentID *uint, excludeID *uint) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domain.Department{}).Where("lower(name) = lower(?)", name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check department name: %w", err)
	}

	return count > 0, nil
}

func (r *departmentRepository) ListChildren(ctx context.Context, parentID uint) ([]domain.Department, error) {
	var departments []domain.Department
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC, id ASC").
		Find(&departments).Error
	if err != nil {
		return nil, fmt.Errorf("list child departments: %w", err)
	}

	return departments, nil
}

func (r *departmentRepository) GetDescendantIDs(ctx context.Context, id uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id FROM departments WHERE parent_id = ?
			UNION ALL
			SELECT departments.id
			FROM departments
			JOIN subtree ON departments.parent_id = subtree.id
		)
		SELECT id FROM subtree
	`, id).Scan(&ids).Error
	if err != nil {
		return nil, fmt.Errorf("get descendant ids: %w", err)
	}

	return ids, nil
}
