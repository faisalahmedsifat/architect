package parser

import (
	"os"

	"github.com/faisalahmedsifat/architect/internal/models"
	"gopkg.in/yaml.v3"
)

func ParseAPIYAML(filepath string) (*models.API, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var api models.API
	if err := yaml.Unmarshal(data, &api); err != nil {
		return nil, err
	}

	return &api, nil
}

func ParseProjectYAML(filepath string) (*models.Project, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	return &project, nil
}
