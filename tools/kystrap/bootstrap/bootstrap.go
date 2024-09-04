package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/KYVENetwork/kyve-rdk/tools/kystrap/types"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var funcMap = template.FuncMap{
	"ToUpper": strings.ToUpper,                      // THIS-here is an example -> THIS-HERE IS AN EXAMPLE
	"ToLower": strings.ToLower,                      // THIS-here is an example -> this-here is an example
	"ToTitle": cases.Title(language.English).String, // THIS-here is an example -> This-Here Is An Example
	"ToPascal": func(s string) string { // THIS-here is an example -> ThisHereIsAnExample
		// remove dashes and underscores
		s = strings.ReplaceAll(s, "-", " ")
		s = strings.ReplaceAll(s, "_", " ")

		// convert to title case
		s = cases.Title(language.English).String(s)

		// remove spaces
		return strings.ReplaceAll(s, " ", "")
	},
}

func readConfig(name string) error {
	viper.SetConfigName(types.TemplateStringFile)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	viper.Set("name", name)
	return nil
}

func createFile(path string, outputPath string, data map[string]any, dirEntry os.DirEntry) error {
	// Check if the file is a directory
	if dirEntry.IsDir() {
		// Create the directory in the output path
		return os.MkdirAll(outputPath, os.ModePerm)
	}

	// Read the template file
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse the template
	tmpl, err := template.New("").Funcs(funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template for file %s with error:\n%s", dirEntry.Name(), err.Error())
	}

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s with error:\n%s", dirEntry.Name(), err.Error())
	}
	//goland:noinspection GoUnhandledErrorResult
	defer outputFile.Close()

	fileInfo, err := dirEntry.Info()
	if err != nil {
		return fmt.Errorf("failed to get file info for file %s with error:\n%s", dirEntry.Name(), err.Error())
	}

	// Set the file permissions
	err = os.Chmod(outputPath, fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set file permissions for file %s with error:\n%s", dirEntry.Name(), err.Error())
	}

	// Execute the template
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return fmt.Errorf("failed to create template for file %s with error:\n%s", dirEntry.Name(), err.Error())
	}
	return nil
}

func CreateRuntime(outputDir string, language types.Language, name string) error {
	// Read the config file
	if err := readConfig(name); err != nil {
		return err
	}

	// Assemble paths
	templateDir := filepath.Join(types.TemplatesDir, strings.ToLower(language.StringValue()))
	outputPath := filepath.Join(outputDir, name)

	// Check if the output directory already exists
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		return errors.New("runtime already exists")
	}

	// Create the output directory
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Data for the templates
	data := viper.GetViper().AllSettings()

	// Walk through the template directory to get all files
	return filepath.WalkDir(templateDir, func(path string, fileInfo os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Remove the template directory from the path
		filePath := strings.Replace(path, templateDir, "", 1)
		newPath := filepath.Join(outputPath, filePath)

		// Create file
		return createFile(path, newPath, data, fileInfo)
	})
}

func UpdateReleasePleaseConfig(language types.Language, name string) error {
	configPath := "release-please-config.json"

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Parse the JSON content into a map
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	releaseType := ""
	switch language.StringValue() {
	case "go":
		releaseType = "go"
	case "typescript":
		releaseType = "node"
	case "python":
		releaseType = "python"
	default:
		return fmt.Errorf("unsupported language: %s", language.StringValue())
	}

	// Create a new package
	packageName := "runtime/" + name
	newPackage := map[string]interface{}{
		"release-type": releaseType,
		"package-name": packageName,
	}

	// Get the packages map from the config
	packages, ok := config["packages"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to parse packages")
	}

	// Add the new package to the packages map
	packages[packageName] = newPackage

	// Convert the updated Config back to JSON
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write the updated JSON back to the config file
	err = os.WriteFile(configPath, updatedData, 0644)
	if err != nil {
		return err
	}

	return nil
}
