package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// IaCType represents the type of Infrastructure as Code
type IaCType string

// Supported IaC types
const (
	TypeTerraform      IaCType = "terraform"
	TypePulumi         IaCType = "pulumi"
	TypeCloudFormation IaCType = "cloudformation"
	TypeAzureARM       IaCType = "azure_arm"
	TypeAnsible        IaCType = "ansible"
	TypeUnknown        IaCType = "unknown"
)

// DetectIaCType tries to determine the IaC type from a directory or file
func DetectIaCType(path string) (IaCType, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return TypeUnknown, err
	}

	// If it's a file, check its extension
	if !info.IsDir() {
		return detectFromFile(path)
	}

	// Check for terraform files
	tfFiles, err := filepath.Glob(filepath.Join(path, "*.tf"))
	if err == nil && len(tfFiles) > 0 {
		return TypeTerraform, nil
	}

	// Check for Pulumi files
	if fileExists(filepath.Join(path, "Pulumi.yaml")) || fileExists(filepath.Join(path, "Pulumi.yml")) {
		return TypePulumi, nil
	}

	// Check for CloudFormation templates
	cfFiles, err := filepath.Glob(filepath.Join(path, "*.template"))
	cfJsonFiles, err2 := filepath.Glob(filepath.Join(path, "*.template.json"))
	cfYamlFiles, err3 := filepath.Glob(filepath.Join(path, "*.template.yaml"))
	if (err == nil && len(cfFiles) > 0) ||
		(err2 == nil && len(cfJsonFiles) > 0) ||
		(err3 == nil && len(cfYamlFiles) > 0) {
		return TypeCloudFormation, nil
	}

	// Check for Azure ARM templates
	armFiles, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err == nil && len(armFiles) > 0 {
		// Check content to confirm it's an ARM template
		for _, file := range armFiles {
			content, err := os.ReadFile(file)
			if err == nil && strings.Contains(string(content), "\"$schema\": \"https://schema.management.azure.com/schemas/") {
				return TypeAzureARM, nil
			}
		}
	}

	// Check for Ansible playbooks
	ansibleFiles, err := filepath.Glob(filepath.Join(path, "*.yml"))
	ansibleYamlFiles, err2 := filepath.Glob(filepath.Join(path, "*.yaml"))
	if (err == nil && len(ansibleFiles) > 0) || (err2 == nil && len(ansibleYamlFiles) > 0) {
		// Check content to confirm it's an Ansible playbook
		for _, file := range append(ansibleFiles, ansibleYamlFiles...) {
			content, err := os.ReadFile(file)
			if err == nil && (strings.Contains(string(content), "hosts:") ||
				strings.Contains(string(content), "tasks:")) {
				return TypeAnsible, nil
			}
		}
	}

	// Could not determine the type
	return TypeUnknown, nil
}

// detectFromFile tries to determine IaC type from a single file
func detectFromFile(path string) (IaCType, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".tf":
		return TypeTerraform, nil
	case ".json":
		// Could be CloudFormation or ARM
		content, err := os.ReadFile(path)
		if err != nil {
			return TypeUnknown, err
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "\"$schema\": \"https://schema.management.azure.com/schemas/") {
			return TypeAzureARM, nil
		}
		if strings.Contains(contentStr, "\"AWSTemplateFormatVersion\"") ||
			strings.Contains(contentStr, "\"Resources\"") {
			return TypeCloudFormation, nil
		}

		return TypeUnknown, nil
	case ".yaml", ".yml":
		// Could be CloudFormation, Ansible, or Pulumi
		content, err := os.ReadFile(path)
		if err != nil {
			return TypeUnknown, err
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "AWSTemplateFormatVersion") {
			return TypeCloudFormation, nil
		}
		if strings.Contains(contentStr, "hosts:") || strings.Contains(contentStr, "tasks:") {
			return TypeAnsible, nil
		}

		return TypeUnknown, nil
	default:
		return TypeUnknown, nil
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
