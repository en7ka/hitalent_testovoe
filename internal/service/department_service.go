package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/en7ka/hitalent_testovoe/internal/domain"
	"github.com/en7ka/hitalent_testovoe/internal/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
)

type CreateDepartmentInput struct {
	Name     string
	ParentID *uint
}

type UpdateDepartmentInput struct {
	Name        *string
	ParentIDSet bool
	ParentID    *uint
}

type DepartmentTree struct {
	Department domain.Department
	Employees  []domain.Employee
	Children   []DepartmentTree
}

type DepartmentService interface {
	Create(ctx context.Context, input CreateDepartmentInput) (domain.Department, error)
	GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (DepartmentTree, error)
	Update(ctx context.Context, id uint, input UpdateDepartmentInput) (domain.Department, error)
	Delete(ctx context.Context, id uint, mode string, reassignToDepartmentID *uint) error
}

type departmentService struct {
	departments repository.DepartmentRepository
	employees   repository.EmployeeRepository
	transactor  repository.Transactor
}

func NewDepartmentService(
	departments repository.DepartmentRepository,
	employees repository.EmployeeRepository,
	transactor repository.Transactor,
) DepartmentService {
	return &departmentService{
		departments: departments,
		employees:   employees,
		transactor:  transactor,
	}
}

func (s *departmentService) Create(ctx context.Context, input CreateDepartmentInput) (domain.Department, error) {
	name, err := normalizeRequiredString(input.Name)
	if err != nil {
		return domain.Department{}, err
	}

	if input.ParentID != nil {
		exists, err := s.departments.Exists(ctx, *input.ParentID)
		if err != nil {
			return domain.Department{}, err
		}
		if !exists {
			return domain.Department{}, ErrNotFound
		}
	}

	exists, err := s.departments.NameExists(ctx, name, input.ParentID, nil)
	if err != nil {
		return domain.Department{}, err
	}
	if exists {
		return domain.Department{}, ErrConflict
	}

	department := domain.Department{
		Name:     name,
		ParentID: input.ParentID,
	}
	if err := s.departments.Create(ctx, &department); err != nil {
		return domain.Department{}, err
	}

	return department, nil
}

func (s *departmentService) GetTree(ctx context.Context, id uint, depth int, includeEmployees bool) (DepartmentTree, error) {
	if id == 0 || depth < 0 || depth > 5 {
		return DepartmentTree{}, ErrInvalidInput
	}

	return s.buildTree(ctx, id, depth, includeEmployees)
}

func (s *departmentService) Update(ctx context.Context, id uint, input UpdateDepartmentInput) (domain.Department, error) {
	if id == 0 {
		return domain.Department{}, ErrInvalidInput
	}

	current, err := s.departments.GetByID(ctx, id)
	if err != nil {
		return domain.Department{}, mapRepositoryError(err)
	}

	name := current.Name
	if input.Name != nil {
		name, err = normalizeRequiredString(*input.Name)
		if err != nil {
			return domain.Department{}, err
		}
	}

	parentID := current.ParentID
	if input.ParentIDSet {
		parentID = input.ParentID
		if parentID != nil {
			if *parentID == id {
				return domain.Department{}, ErrConflict
			}

			parentExists, err := s.departments.Exists(ctx, *parentID)
			if err != nil {
				return domain.Department{}, err
			}
			if !parentExists {
				return domain.Department{}, ErrNotFound
			}

			descendants, err := s.departments.GetDescendantIDs(ctx, id)
			if err != nil {
				return domain.Department{}, err
			}
			if containsID(descendants, *parentID) {
				return domain.Department{}, ErrConflict
			}
		}
	}

	nameExists, err := s.departments.NameExists(ctx, name, parentID, &id)
	if err != nil {
		return domain.Department{}, err
	}
	if nameExists {
		return domain.Department{}, ErrConflict
	}

	updated := domain.Department{
		ID:       id,
		Name:     name,
		ParentID: parentID,
	}
	if err := s.departments.Update(ctx, &updated); err != nil {
		return domain.Department{}, mapRepositoryError(err)
	}

	return updated, nil
}

func (s *departmentService) Delete(ctx context.Context, id uint, mode string, reassignToDepartmentID *uint) error {
	if id == 0 {
		return ErrInvalidInput
	}

	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "cascade"
	}

	switch mode {
	case "cascade":
		if err := s.departments.Delete(ctx, id); err != nil {
			return mapRepositoryError(err)
		}
		return nil
	case "reassign":
		if reassignToDepartmentID == nil || *reassignToDepartmentID == 0 {
			return ErrInvalidInput
		}
		if *reassignToDepartmentID == id {
			return ErrConflict
		}

		return s.transactor.WithinTransaction(ctx, func(txCtx context.Context, departments repository.DepartmentRepository, employees repository.EmployeeRepository) error {
			if _, err := departments.GetByID(txCtx, id); err != nil {
				return mapRepositoryError(err)
			}

			targetExists, err := departments.Exists(txCtx, *reassignToDepartmentID)
			if err != nil {
				return err
			}
			if !targetExists {
				return ErrNotFound
			}

			descendants, err := departments.GetDescendantIDs(txCtx, id)
			if err != nil {
				return err
			}
			if containsID(descendants, *reassignToDepartmentID) {
				return ErrConflict
			}

			departmentIDs := append([]uint{id}, descendants...)
			if err := employees.ReassignDepartments(txCtx, departmentIDs, *reassignToDepartmentID); err != nil {
				return err
			}

			return mapRepositoryError(departments.Delete(txCtx, id))
		})
	default:
		return ErrInvalidInput
	}
}

func (s *departmentService) buildTree(ctx context.Context, id uint, depth int, includeEmployees bool) (DepartmentTree, error) {
	department, err := s.departments.GetByID(ctx, id)
	if err != nil {
		return DepartmentTree{}, mapRepositoryError(err)
	}

	tree := DepartmentTree{Department: department}
	if includeEmployees {
		tree.Employees, err = s.employees.ListByDepartmentID(ctx, id)
		if err != nil {
			return DepartmentTree{}, err
		}
	}

	if depth == 0 {
		return tree, nil
	}

	children, err := s.departments.ListChildren(ctx, id)
	if err != nil {
		return DepartmentTree{}, err
	}

	tree.Children = make([]DepartmentTree, 0, len(children))
	for _, child := range children {
		childTree, err := s.buildTree(ctx, child.ID, depth-1, includeEmployees)
		if err != nil {
			return DepartmentTree{}, err
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

func normalizeRequiredString(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 200 {
		return "", ErrInvalidInput
	}

	return value, nil
}

func containsID(ids []uint, id uint) bool {
	for _, current := range ids {
		if current == id {
			return true
		}
	}

	return false
}

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}

	return err
}
