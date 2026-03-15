package main

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name     string
	Path     string `gorm:"uniqueIndex"`
	Language string
	Template string
	Git      bool
	GitUser  string
	GitRepo  string
}

type Store struct {
	db *gorm.DB
}

func NewStore() *Store {
	db, err := gorm.Open(sqlite.Open("store.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return &Store{db: db}
}

func (s *Store) Init() error {
	return s.db.AutoMigrate(&Project{})
}

func (s *Store) AddProject(project Project) (Project, error) {
	if project.Path == "" {
		return project, errors.New("project path is required")
	}

	var existing Project
	err := s.db.Where("path = ?", project.Path).First(&existing).Error
	if err == nil {
		project.ID = existing.ID
		updates := map[string]any{
			"name":     project.Name,
			"path":     project.Path,
			"language": project.Language,
			"template": project.Template,
			"git":      project.Git,
			"git_user": project.GitUser,
			"git_repo": project.GitRepo,
		}
		if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
			return project, err
		}
		var updated Project
		if err := s.db.First(&updated, existing.ID).Error; err != nil {
			return project, err
		}
		return updated, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return project, err
	}

	if err := s.db.Create(&project).Error; err != nil {
		return project, err
	}
	return project, nil
}

func (s *Store) DeleteProject(id uint) error {
	return s.db.Delete(&Project{}, id).Error
}

func (s *Store) GetProjects() ([]Project, error) {
	var projects []Project
	if err := s.db.Order("name asc").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}
