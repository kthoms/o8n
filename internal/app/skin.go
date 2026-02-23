package app

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Skin struct {
	O8n struct {
		Body struct {
			FgColor   string `yaml:"fgColor"`
			BgColor   string `yaml:"bgColor"`
			LogoColor string `yaml:"logoColor"`
		} `yaml:"body"`
		Frame struct {
			Border struct {
				FgColor    string `yaml:"fgColor"`
				FocusColor string `yaml:"focusColor"`
			} `yaml:"border"`
		} `yaml:"frame"`
	} `yaml:"o8n"`
}

func loadSkin(skinName string) (*Skin, error) {
	if skinName == "" {
		skinName = "stock.yaml"
	}
	path := filepath.Join("skins", skinName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skin file %s: %w", path, err)
	}

	var skin Skin
	if err := yaml.Unmarshal(data, &skin); err != nil {
		return nil, fmt.Errorf("failed to parse skin file %s: %w", path, err)
	}

	return &skin, nil
}
