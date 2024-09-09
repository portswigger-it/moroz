package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"gopkg.in/yaml.v2"
)

// Struct to hold the YAML format
type Rule struct {
	RuleType   string `yaml:"rule_type"`
	Policy     string `yaml:"policy"`
	Identifier string `yaml:"identifier"`
	CustomMsg  string `yaml:"custom_msg"`
}

func main() {
	// Check if the file path is passed as an argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run script.go /path/to/application")
		os.Exit(1)
	}

	// Get the file path from the argument
	filePath := os.Args[1]

	// Run the santactl command
	cmd := exec.Command("santactl", "fileinfo", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		os.Exit(1)
	}

	// Extract Organization and Organizational Unit using regex
	output := out.String()
	orgRegex := regexp.MustCompile(`Organization\s+:\s+(.+)`)
	orgUnitRegex := regexp.MustCompile(`Organizational Unit\s+:\s+(.+)`)

	orgMatch := orgRegex.FindStringSubmatch(output)
	orgUnitMatch := orgUnitRegex.FindStringSubmatch(output)

	if len(orgMatch) > 1 && len(orgUnitMatch) > 1 {
		org := orgMatch[1]
		orgUnit := orgUnitMatch[1]

		// Create the struct for YAML
		rule := Rule{
			RuleType:   "TEAMID",
			Policy:     "ALLOWLIST",
			Identifier: orgUnit,
			CustomMsg:  org,
		}

		// Convert struct to YAML
		yamlData, err := yaml.Marshal([]Rule{rule})
		if err != nil {
			fmt.Printf("Error converting to YAML: %v\n", err)
			os.Exit(1)
		}

		// Print the YAML output
		fmt.Println(string(yamlData))
	} else {
		fmt.Println("Organization or Organizational Unit not found.")
	}
}
