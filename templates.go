package main

import (
	"encoding/json"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed prints/**
var printsFS embed.FS

type TemplateStore struct {
	fs fs.FS
}

type TemplateInfo struct {
	Name        string `json:"name"`
	Language    string `json:"language"`
	Description string `json:"description"`
}

func NewTemplateStore() *TemplateStore {
	return &TemplateStore{fs: printsFS}
}

func (t *TemplateStore) List() ([]string, error) {
	entries, err := fs.ReadDir(t.fs, "prints")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	sort.Strings(names)
	return names, nil
}

func (t *TemplateStore) Has(name string) (bool, error) {
	root, err := t.templateRoot(name)
	if err != nil {
		return false, err
	}

	_, err = fs.Stat(t.fs, root)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (t *TemplateStore) Templates() ([]TemplateInfo, error) {
	names, err := t.List()
	if err != nil {
		return nil, err
	}

	templates := make([]TemplateInfo, 0, len(names))
	for _, name := range names {
		info := TemplateInfo{Name: name}
		metaPath := path.Join("prints", name, "template.json")
		if data, err := fs.ReadFile(t.fs, metaPath); err == nil {
			var parsed TemplateInfo
			if err := json.Unmarshal(data, &parsed); err == nil {
				if parsed.Name != "" {
					info.Name = parsed.Name
				}
				info.Language = parsed.Language
				info.Description = parsed.Description
			}
		}
		templates = append(templates, info)
	}

	return templates, nil
}

func (t *TemplateStore) Extract(name, dest string) error {
	root, err := t.templateRoot(name)
	if err != nil {
		return err
	}

	if err := ensureDirReady(dest); err != nil {
		return err
	}

	return fs.WalkDir(t.fs, root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == root {
			return nil
		}

		rel := strings.TrimPrefix(p, root)
		rel = strings.TrimPrefix(rel, "/")
		target := rewriteTemplatePath(filepath.Join(dest, filepath.FromSlash(rel)))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		src, err := t.fs.Open(p)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("file already exists: %s", target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	})
}

func (t *TemplateStore) templateRoot(name string) (string, error) {
	if name == "" {
		return "", errors.New("template name is required")
	}
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid template name: %s", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("template name must be a single folder: %s", name)
	}

	return path.Join("prints", name), nil
}

func ensureDirReady(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.MkdirAll(dir, 0o755)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("destination is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("destination directory is not empty: %s", dir)
	}

	return nil
}

func rewriteTemplatePath(target string) string {
	switch filepath.Base(target) {
	case "go.mod.txt":
		return filepath.Join(filepath.Dir(target), "go.mod")
	case "go.sum.txt":
		return filepath.Join(filepath.Dir(target), "go.sum")
	default:
		return target
	}
}
