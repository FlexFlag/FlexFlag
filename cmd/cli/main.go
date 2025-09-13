package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL      string
	environment string
)

var rootCmd = &cobra.Command{
	Use:   "flexflag",
	Short: "FlexFlag CLI - Feature flag management",
	Long:  `FlexFlag CLI is a command-line tool for managing feature flags in the FlexFlag system.`,
}

var createCmd = &cobra.Command{
	Use:   "create [flag-key]",
	Short: "Create a new feature flag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagKey := args[0]
		
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		flagType, _ := cmd.Flags().GetString("type")
		defaultValue, _ := cmd.Flags().GetString("default")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if name == "" {
			name = flagKey
		}

		var defVal interface{}
		switch flagType {
		case "boolean":
			defVal = defaultValue == "true"
		case "number":
			defVal = 0
			_, _ = fmt.Sscanf(defaultValue, "%f", &defVal)
		case "json":
			_ = json.Unmarshal([]byte(defaultValue), &defVal)
		default:
			defVal = defaultValue
		}

		payload := map[string]interface{}{
			"key":         flagKey,
			"name":        name,
			"description": description,
			"type":        flagType,
			"default":     defVal,
			"enabled":     enabled,
			"environment": environment,
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/flags", apiURL), "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating flag: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			fmt.Printf("✅ Flag '%s' created successfully\n", flagKey)
			
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			
			fmt.Printf("  Type: %s\n", flagType)
			fmt.Printf("  Default: %v\n", defVal)
			fmt.Printf("  Enabled: %v\n", enabled)
			fmt.Printf("  Environment: %s\n", environment)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: %s\n", string(body))
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all feature flags",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/flags?environment=%s", apiURL, environment))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing flags: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result struct {
			Flags []map[string]interface{} `json:"flags"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&result)

		if len(result.Flags) == 0 {
			fmt.Println("No flags found")
			return
		}

		fmt.Printf("%-20s %-10s %-10s %-30s\n", "KEY", "TYPE", "ENABLED", "DESCRIPTION")
		fmt.Println(string(bytes.Repeat([]byte("-"), 70)))
		
		for _, flag := range result.Flags {
			key := flag["key"].(string)
			flagType := flag["type"].(string)
			enabled := flag["enabled"].(bool)
			description := ""
			if desc, ok := flag["description"].(string); ok {
				if len(desc) > 27 {
					description = desc[:27] + "..."
				} else {
					description = desc
				}
			}
			
			fmt.Printf("%-20s %-10s %-10v %-30s\n", key, flagType, enabled, description)
		}
	},
}

var toggleCmd = &cobra.Command{
	Use:   "toggle [flag-key]",
	Short: "Toggle a feature flag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagKey := args[0]
		
		client := &http.Client{Timeout: 10 * time.Second}
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/flags/%s/toggle?environment=%s", apiURL, flagKey, environment), nil)
		
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error toggling flag: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			fmt.Printf("✅ Flag '%s' toggled successfully. Enabled: %v\n", flagKey, result["enabled"])
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: %s\n", string(body))
			os.Exit(1)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get [flag-key]",
	Short: "Get details of a feature flag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagKey := args[0]
		
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/flags/%s?environment=%s", apiURL, flagKey, environment))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting flag: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var flag map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&flag)
			
			fmt.Printf("Flag: %s\n", flag["key"])
			fmt.Printf("  Name: %s\n", flag["name"])
			fmt.Printf("  Type: %s\n", flag["type"])
			fmt.Printf("  Enabled: %v\n", flag["enabled"])
			fmt.Printf("  Description: %s\n", flag["description"])
			fmt.Printf("  Default: %s\n", string(flag["default"].([]byte)))
			fmt.Printf("  Environment: %s\n", flag["environment"])
			fmt.Printf("  Created: %s\n", flag["created_at"])
		} else {
			fmt.Fprintf(os.Stderr, "Flag not found: %s\n", flagKey)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().StringVar(&environment, "env", "production", "Environment")

	createCmd.Flags().String("name", "", "Flag name")
	createCmd.Flags().String("description", "", "Flag description")
	createCmd.Flags().String("type", "boolean", "Flag type (boolean, string, number, json)")
	createCmd.Flags().String("default", "false", "Default value")
	createCmd.Flags().Bool("enabled", false, "Enable flag immediately")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(toggleCmd)
	rootCmd.AddCommand(getCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}