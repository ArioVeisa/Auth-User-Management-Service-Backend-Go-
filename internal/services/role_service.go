package services

import (
	"errors"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
}

func NewRoleService(roleRepo *repository.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (s *RoleService) GetRole(id int) (*models.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *RoleService) ListRoles() ([]models.Role, error) {
	return s.roleRepo.List()
}

func (s *RoleService) CreateRole(req models.CreateRoleRequest) (*models.Role, error) {
	existing, _ := s.roleRepo.GetByName(req.Name)
	if existing != nil {
		return nil, errors.New("role name already exists")
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *RoleService) UpdateRole(id int, req models.CreateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	if req.Name != role.Name {
		existing, _ := s.roleRepo.GetByName(req.Name)
		if existing != nil {
			return nil, errors.New("role name already exists")
		}
	}

	role.Name = req.Name
	role.Description = req.Description

	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *RoleService) DeleteRole(id int) error {
	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return errors.New("role not found")
	}

	return s.roleRepo.Delete(id)
}
