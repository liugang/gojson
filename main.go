package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	name        string
	pkg         string
	inputFile   string
	outputFile  string
	tags        string
	fileType    string
	forceFloats bool
	subStruct   bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gojson",
	Short: "Generate Go struct definitions from JSON/YAML",
	Long: `gojson generates Go struct definitions from JSON or YAML documents.
It reads from stdin or a file and outputs the generated structs to stdout or a file.

Example:
  curl -s https://api.github.com/repos/chimeracoder/gojson | gojson -name=Repository`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate fileType
		if fileType == "" && inputFile != "" {
			if strings.HasSuffix(strings.ToLower(inputFile), ".json") {
				fileType = "json"
			} else if strings.HasSuffix(strings.ToLower(inputFile), ".yaml") || strings.HasSuffix(strings.ToLower(inputFile), ".yml") {
				fileType = "yaml"
			}
		}
		if fileType == "" {
			fileType = "json"
		}
		if fileType != "json" && fileType != "yaml" {
			return fmt.Errorf("fileType must be json or yaml")
		}

		// Handle tags
		tagList := make([]string, 0)
		if tags == "" || tags == "fmt" {
			tagList = append(tagList, fileType)
		} else {
			tagList = strings.Split(tags, ",")
		}

		// Handle input
		var input io.Reader = os.Stdin
		if inputFile != "" {
			f, err := os.Open(inputFile)
			if err != nil {
				return fmt.Errorf("reading input file: %s", err)
			}
			defer f.Close()
			input = f
		} else if isInteractive() {
			return fmt.Errorf("expects input on stdin when no input file specified")
		}

		// Select parser based on fileType
		var parser Parser
		var convertFloats bool
		switch fileType {
		case "json":
			parser = ParseJson
			convertFloats = true
		case "yaml":
			parser = ParseYaml
		}

		// Generate struct
		output, err := Generate(input, parser, name, pkg, tagList, subStruct, convertFloats)
		if err != nil {
			return fmt.Errorf("error generating struct: %s", err)
		}

		// Handle output
		if outputFile != "" {
			if err := os.WriteFile(outputFile, output, 0644); err != nil {
				return fmt.Errorf("writing output file: %s", err)
			}
		} else {
			fmt.Print(string(output))
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", "Foo", "the name of the struct")
	rootCmd.PersistentFlags().StringVarP(&pkg, "pkg", "p", "", "the name of the package for the generated code")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "the name of the input file containing JSON (if input not provided via STDIN)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "the name of the file to write the output to (outputs to STDOUT by default)")
	rootCmd.PersistentFlags().StringVarP(&fileType, "fmt", "f", "", "the fileType of the input data (json or yaml)")
	rootCmd.PersistentFlags().StringVarP(&tags, "tags", "t", "fmt", "comma separated list of the tags to put on the struct")
	rootCmd.PersistentFlags().BoolVar(&forceFloats, "forcefloats", false, "[experimental] force float64 type for integral values")
	rootCmd.PersistentFlags().BoolVarP(&subStruct, "substruct", "s", false, "create types for sub-structs")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Return true if os.Stdin appears to be interactive
func isInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fileInfo.Mode()&(os.ModeCharDevice|os.ModeCharDevice) != 0
}
