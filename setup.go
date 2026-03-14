package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Setup struct {
	Templates *TemplateStore
}

type Language int

const (
	LangUnknown Language = iota
	LangRust
	LangPython
	LangGo
)

type SetupOptions struct {
	ProjectDir string
	Git        bool
	Language   Language
	GitUser    string
	GitRepo    string
	Template   string
}

func NewSetup() *Setup {
	return &Setup{Templates: NewTemplateStore()}
}

func (s *Setup) Setup(options SetupOptions) error {
	if options.ProjectDir == "" {
		return errors.New("project directory is required")
	}

	if s.Templates == nil {
		s.Templates = NewTemplateStore()
	}

	if err := os.MkdirAll(options.ProjectDir, 0o755); err != nil {
		return fmt.Errorf("create project directory: %w", err)
	}

	if options.Template != "" {
		if err := s.Templates.Extract(options.Template, options.ProjectDir); err != nil {
			return fmt.Errorf("extract template: %w", err)
		}
	}

	if options.Git {
		if err := initGit(options.ProjectDir, options.GitUser, options.GitRepo); err != nil {
			return fmt.Errorf("git setup: %w", err)
		}
	}

	return nil
}

func initGit(dir, user, repo string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return nil
	}

	if err := runCmd(dir, "git", "init"); err != nil {
		return err
	}

	if user != "" && repo != "" {
		remote := fmt.Sprintf("https://github.com/%s/%s.git", user, repo)
		if err := runCmd(dir, "git", "remote", "add", "origin", remote); err != nil {
			return err
		}
	}

	return nil
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
