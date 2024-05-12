//go:generate go run versiongen/gen-version.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MetricConfig represents the YAML configuration file structure.
type MetricConfig struct {
	Metrics      []Metric        `yaml:"metrics"`
	PackageName  string          `yaml:"package_name"`
	UniqueLabels map[string]bool `yaml:"-"`
}

type Metric struct {
	Name    string    `yaml:"name"`
	Type    string    `yaml:"type"`
	Labels  []string  `yaml:"labels,omitempty"`
	Help    string    `yaml:"help,omitempty"`
	Buckets []float64 `yaml:"buckets,omitempty"`
}

// Convert snake_case to CamelCase
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	c := cases.Title(language.English)
	for i, part := range parts {
		parts[i] = c.String(part)
	}
	return strings.Join(parts, "")
}

func main() {
	var configPath, outputPath, packageName string

	var rootCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generates Prometheus metrics based on a JSON configuration",
		Long: `A tool to generate Prometheus metrics Go code from a JSON configuration file.
Complete documentation is available at http://example.com`,
		Run: func(cmd *cobra.Command, args []string) {
			// Load and parse the YAML configuration file.
			content, err := os.ReadFile(configPath)
			if err != nil {
				fmt.Printf("error reading config file: %v\n", err)
				os.Exit(1)
			}

			// Validate the JSON config
			err = validateConfig(content)
			if err != nil {
				fmt.Printf("config validation failed: %v\n", err)
				os.Exit(1)
			}

			var config MetricConfig
			err = json.Unmarshal(content, &config)
			if err != nil {
				fmt.Printf("error parsing config file: %v\n", err)
				os.Exit(1)
			}

			// Populate unique labels
			config.UniqueLabels = make(map[string]bool)
			for _, metric := range config.Metrics {
				for _, label := range metric.Labels {
					config.UniqueLabels[label] = true
				}
			}

			// Define a custom function map
			funcMap := template.FuncMap{
				"snakeToCamel": snakeToCamel,
			}

			// Generate Go code from the template with the custom function map.
			t, err := template.New("metrics").Funcs(funcMap).Parse(metricsTemplate)
			if err != nil {
				fmt.Printf("error parsing template: %v\n", err)
				os.Exit(1)
			}

			// Create a buffer to hold the executed template before formatting.
			var buf bytes.Buffer

			// Set package name in the config passed for template execution
			config.PackageName = packageName

			err = t.Execute(&buf, config)
			if err != nil {
				fmt.Printf("error executing template: %v\n", err)
				os.Exit(1)
			}

			// Format the source code in the buffer.
			formattedSource, err := format.Source(buf.Bytes())
			if err != nil {
				fmt.Printf("error formatting source: %v\n", err)
				os.Exit(1)
			}

			// Create the output file.
			outputFile, err := os.Create(outputPath)
			if err != nil {
				fmt.Printf("error creating output file: %v\n", err)
				os.Exit(1)
			}
			defer outputFile.Close()

			// Write the formatted source code to the output file.
			_, err = outputFile.Write(formattedSource)
			if err != nil {
				fmt.Printf("error writing to output file: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file (required)")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to the output file (required)")
	rootCmd.Flags().StringVarP(&packageName, "package", "p", "", "Package name for the output file (required)")

	rootCmd.MarkFlagRequired("config")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("package")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\n", version, commit)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateConfig(content []byte) error {
	// Load the JSON schema
	schemaLoader := gojsonschema.NewStringLoader(metricConfigSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("error parsing schema: %v", err)
	}

	// Load the JSON config
	documentLoader := gojsonschema.NewBytesLoader(content)

	// Validate the JSON config against the schema
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("error validating config: %v", err)
	}

	if !result.Valid() {
		var errMessages []string
		for _, err := range result.Errors() {
			errMessages = append(errMessages, fmt.Sprintf("- %s", err))
		}
		return fmt.Errorf("invalid config:\n%s", strings.Join(errMessages, "\n"))
	}

	return nil
}
