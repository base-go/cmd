package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/base-go/cmd/utils"
	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:     "scheduler",
	Short:   "Scheduler management commands",
	Long:    `Generate and manage scheduled tasks for your Base Framework application.`,
	Aliases: []string{"sc"},
}

var schedulerGenerateCmd = &cobra.Command{
	Use:   "generate [module] [task-name]",
	Short: "Generate a new scheduled task",
	Long: `Generate a new scheduled task file in the specified module.

Examples:
  base scheduler generate posts cleanup-old-posts
  base scheduler g users send-weekly-digest
  base scheduler g core backup-database`,
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(2),
	Run:     generateTask,
}

var schedulerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all registered tasks",
	Long:    `List all registered tasks in the application with their status.`,
	Aliases: []string{"ls"},
	Run:     listTasks,
}

var schedulerRunCmd = &cobra.Command{
	Use:   "run [task-name]",
	Short: "Run a specific task immediately",
	Long:  `Execute a specific task immediately, bypassing its schedule.`,
	Args:  cobra.ExactArgs(1),
	Run:   runTask,
}

var schedulerEnableCmd = &cobra.Command{
	Use:   "enable [task-name]",
	Short: "Enable a specific task",
	Long:  `Enable a specific task to run according to its schedule.`,
	Args:  cobra.ExactArgs(1),
	Run:   enableTask,
}

var schedulerDisableCmd = &cobra.Command{
	Use:   "disable [task-name]",
	Short: "Disable a specific task",
	Long:  `Disable a specific task to prevent it from running.`,
	Args:  cobra.ExactArgs(1),
	Run:   disableTask,
}

var schedulerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get scheduler status",
	Long:  `Get the current status and statistics of the scheduler.`,
	Run:   getSchedulerStatus,
}

// Global flags
var (
	apiKey  string
	baseURL string
)

func init() {
	schedulerCmd.AddCommand(schedulerGenerateCmd)
	schedulerCmd.AddCommand(schedulerListCmd)
	schedulerCmd.AddCommand(schedulerRunCmd)
	schedulerCmd.AddCommand(schedulerEnableCmd)
	schedulerCmd.AddCommand(schedulerDisableCmd)
	schedulerCmd.AddCommand(schedulerStatusCmd)

	// Add global flags
	schedulerCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication")
	schedulerCmd.PersistentFlags().StringVar(&baseURL, "url", "http://localhost:8100", "Base URL of the application")

	rootCmd.AddCommand(schedulerCmd)
}

func generateTask(cmd *cobra.Command, args []string) {
	moduleName := args[0]
	taskName := args[1]

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: Could not get current directory: %v\n", err)
		return
	}

	// Check if we're in a Base Framework project (look for main.go)
	mainGoPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		fmt.Println("Error: Not in a Base Framework project directory")
		fmt.Println("Please run this command from the root of your Base Framework project")
		return
	}

	// Determine module path - try multiple variations
	var modulePath string

	if moduleName == "core" {
		modulePath = filepath.Join(cwd, "core", moduleName)
	} else {
		// Try different variations of the module name
		variations := []string{
			moduleName,                  // As provided (e.g., "Post")
			strings.ToLower(moduleName), // Lowercase (e.g., "post")
			utils.PluralizeClient.Plural(strings.ToLower(moduleName)), // Pluralized lowercase (e.g., "posts")
		}

		for _, variation := range variations {
			testPath := filepath.Join(cwd, "app", variation)
			if _, err := os.Stat(testPath); err == nil {
				modulePath = testPath
				break
			}
		}

		// If no variation found, use the original
		if modulePath == "" {
			modulePath = filepath.Join(cwd, "app", moduleName)
		}
	}

	// Check if module exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		fmt.Printf("Error: Module '%s' does not exist\n", moduleName)
		fmt.Printf("Expected path: %s\n", modulePath)

		// Suggest possible variations
		if moduleName != "core" {
			fmt.Printf("\nTried variations:\n")
			fmt.Printf("  - %s (as provided)\n", filepath.Join(cwd, "app", moduleName))
			fmt.Printf("  - %s (lowercase)\n", filepath.Join(cwd, "app", strings.ToLower(moduleName)))
			fmt.Printf("  - %s (pluralized)\n", filepath.Join(cwd, "app", utils.PluralizeClient.Plural(strings.ToLower(moduleName))))
		}
		return
	}

	// Create task file
	taskFileName := fmt.Sprintf("%s_task.go", utils.ToSnakeCase(taskName))
	taskFilePath := filepath.Join(modulePath, taskFileName)

	// Check if task file already exists
	if _, err := os.Stat(taskFilePath); err == nil {
		fmt.Printf("Error: Task file already exists: %s\n", taskFilePath)
		return
	}

	// Generate task content using the actual module name found
	actualPackageName := filepath.Base(modulePath)
	taskContent := generateTaskContent(actualPackageName, taskName)

	// Write task file
	err = os.WriteFile(taskFilePath, []byte(taskContent), 0644)
	if err != nil {
		fmt.Printf("Error: Could not create task file: %v\n", err)
		return
	}

	fmt.Printf("✅ Task generated successfully!\n")
	fmt.Printf("   📁 Module: %s\n", moduleName)
	fmt.Printf("   📄 File: %s\n", taskFilePath)
	fmt.Printf("   🔧 Task: %s\n", taskName)
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("   1. Edit the task file to implement your logic\n")
	fmt.Printf("   2. Register the task in your module's initialization\n")
	fmt.Printf("   3. Configure the schedule (daily, monthly, interval, or cron)\n")
	fmt.Printf("   4. Test with: base scheduler run %s\n", utils.ToKebabCase(taskName))
}

func generateTaskContent(moduleName, taskName string) string {
	packageName := moduleName
	if moduleName == "core" {
		packageName = "core"
	}

	taskStructName := utils.ToPascalCase(taskName) + "Task"
	_ = utils.ToCamelCase(taskName) // Reserved for future use
	taskID := utils.ToKebabCase(taskName)

	return fmt.Sprintf(`package %s

import (
	"context"
	"time"

	"base/core/logger"
	"base/core/scheduler"
)

// %s handles %s
type %s struct {
	logger logger.Logger
}

// New%s creates a new %s instance
func New%s(log logger.Logger) *%s {
	return &%s{
		logger: log,
	}
}

// RegisterTask registers the %s task with the scheduler
func (t *%s) RegisterTask(s *scheduler.Scheduler) error {
	task := &scheduler.Task{
		Name:        "%s",
		Description: "%s task for %s module",
		Schedule:    &scheduler.DailySchedule{Hour: 2, Minute: 0}, // 2:00 AM daily
		Handler:     t.execute,
		Enabled:     true,
	}

	return s.RegisterTask(task)
}

// RegisterCronTask registers the %s task with cron scheduler (alternative)
func (t *%s) RegisterCronTask(cs *scheduler.CronScheduler) error {
	task := &scheduler.CronTask{
		Name:        "%s",
		Description: "%s task for %s module",
		CronExpr:    "0 0 2 * * *", // 2:00 AM daily
		Handler:     t.execute,
		Enabled:     true,
	}

	return cs.RegisterTask(task)
}

// execute is the main task execution function
func (t *%s) execute(ctx context.Context) error {
	t.logger.Info("Starting %s task")

	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// TODO: Implement your task logic here
	// Example:
	// - Clean up old records
	// - Send notifications
	// - Generate reports
	// - Backup data
	// - Process queued items

	// Simulate work (remove this in your implementation)
	time.Sleep(1 * time.Second)

	t.logger.Info("%s task completed successfully")
	return nil
}

// GetTaskInfo returns information about this task
func (t *%s) GetTaskInfo() map[string]any {
	return map[string]any{
		"name":        "%s",
		"description": "%s task for %s module",
		"module":      "%s",
		"type":        "scheduled_task",
	}
}
`,
		packageName,
		taskStructName, taskName, taskStructName,
		taskStructName, taskStructName, taskStructName, taskStructName, taskStructName,
		taskName, taskStructName,
		taskID, taskName, moduleName,
		taskName, taskStructName,
		taskID, taskName, moduleName,
		taskStructName, taskName,
		taskName,
		taskStructName,
		taskID, taskName, moduleName, moduleName,
	)
}

func listTasks(cmd *cobra.Command, args []string) {
	if apiKey != "" {
		// Make API call to get tasks
		tasks, err := makeAPIRequest("GET", "/api/scheduler/tasks", nil)
		if err != nil {
			fmt.Printf("❌ Cannot connect to Base Framework server at %s\n", baseURL)
			fmt.Printf("Error: %v\n\n", err)
			
			fmt.Println("💡 Possible solutions:")
			fmt.Println("   1. Start your Base Framework server:")
			fmt.Println("      base start")
			fmt.Println("   2. Check if server is running on a different port:")
			fmt.Println("      base scheduler list --url=http://localhost:8080 --api-key=your-key")
			fmt.Println("   3. For Docker deployments:")
			fmt.Println("      base scheduler list --url=http://your-container:8100 --api-key=your-key")
			fmt.Println("   4. For remote deployments:")
			fmt.Println("      base scheduler list --url=https://your-domain.com --api-key=your-key")
			return
		}
		displayTasks(tasks)
	} else {
		fmt.Println("📋 Listing registered tasks...")
		fmt.Println("💡 To get real-time task status, provide an API key:")
		fmt.Printf("   base scheduler list --api-key=your-key\n")
		fmt.Printf("   base scheduler list --url=%s --api-key=your-key\n", baseURL)
		fmt.Println("\n🌐 For different environments:")
		fmt.Println("   Local:  --url=http://localhost:8100")
		fmt.Println("   Docker: --url=http://container-name:8100")
		fmt.Println("   Remote: --url=https://your-domain.com")
		fmt.Println("\n🔧 Available task management commands:")
		fmt.Println("   base scheduler run <task-name>     - Run task immediately")
		fmt.Println("   base scheduler enable <task-name>  - Enable task")
		fmt.Println("   base scheduler disable <task-name> - Disable task")
	}
}

func runTask(cmd *cobra.Command, args []string) {
	taskName := args[0]
	if apiKey != "" {
		// Make API call to run task
		response, err := makeAPIRequest("POST", fmt.Sprintf("/api/scheduler/tasks/%s/run", taskName), nil)
		if err != nil {
			fmt.Printf("❌ Cannot connect to Base Framework server at %s\n", baseURL)
			fmt.Printf("Error: %v\n\n", err)
			showConnectionHelp()
			return
		}
		fmt.Printf("✅ Task '%s' executed successfully\n", taskName)
		if response != nil {
			if msg, ok := response["message"]; ok {
				fmt.Printf("Response: %v\n", msg)
			}
		}
	} else {
		fmt.Printf("🚀 Running task '%s' immediately...\n", taskName)
		fmt.Println("💡 Provide API key and server URL to run tasks:")
		fmt.Printf("   base scheduler run %s --api-key=your-key --url=%s\n", taskName, baseURL)
		showEnvironmentExamples()
	}
}

func enableTask(cmd *cobra.Command, args []string) {
	taskName := args[0]
	if apiKey != "" {
		// Make API call to enable task
		response, err := makeAPIRequest("PUT", fmt.Sprintf("/api/scheduler/tasks/%s/enable", taskName), nil)
		if err != nil {
			fmt.Printf("❌ Cannot connect to Base Framework server at %s\n", baseURL)
			fmt.Printf("Error: %v\n\n", err)
			showConnectionHelp()
			return
		}
		fmt.Printf("✅ Task '%s' enabled successfully\n", taskName)
		if response != nil {
			if msg, ok := response["message"]; ok {
				fmt.Printf("Response: %v\n", msg)
			}
		}
	} else {
		fmt.Printf("✅ Enabling task '%s'...\n", taskName)
		fmt.Println("💡 Provide API key and server URL to enable tasks:")
		fmt.Printf("   base scheduler enable %s --api-key=your-key --url=%s\n", taskName, baseURL)
		showEnvironmentExamples()
	}
}

func disableTask(cmd *cobra.Command, args []string) {
	taskName := args[0]
	if apiKey != "" {
		// Make API call to disable task
		response, err := makeAPIRequest("PUT", fmt.Sprintf("/api/scheduler/tasks/%s/disable", taskName), nil)
		if err != nil {
			fmt.Printf("❌ Cannot connect to Base Framework server at %s\n", baseURL)
			fmt.Printf("Error: %v\n\n", err)
			showConnectionHelp()
			return
		}
		fmt.Printf("⏸️  Task '%s' disabled successfully\n", taskName)
		if response != nil {
			if msg, ok := response["message"]; ok {
				fmt.Printf("Response: %v\n", msg)
			}
		}
	} else {
		fmt.Printf("⏸️  Disabling task '%s'...\n", taskName)
		fmt.Println("💡 Provide API key and server URL to disable tasks:")
		fmt.Printf("   base scheduler disable %s --api-key=your-key --url=%s\n", taskName, baseURL)
		showEnvironmentExamples()
	}
}

// HTTP client functions
func makeAPIRequest(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	url := strings.TrimSuffix(baseURL, "/") + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("scheduler API endpoint not found (404) - the scheduler module may not be enabled or the API path is incorrect")
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func displayTasks(response map[string]interface{}) {
	if data, ok := response["data"]; ok {
		if tasks, ok := data.([]interface{}); ok {
			fmt.Printf("\n📋 Found %d tasks:\n\n", len(tasks))
			for _, taskInterface := range tasks {
				if task, ok := taskInterface.(map[string]interface{}); ok {
					name := getStringValue(task, "name")
					description := getStringValue(task, "description")
					enabled := getBoolValue(task, "enabled")
					schedule := getStringValue(task, "schedule")
					runCount := getIntValue(task, "run_count")
					errorCount := getIntValue(task, "error_count")
					lastRun := getStringValue(task, "last_run")
					nextRun := getStringValue(task, "next_run")

					status := "🔴 Disabled"
					if enabled {
						status = "🟢 Enabled"
					}

					fmt.Printf("📌 %s %s\n", name, status)
					fmt.Printf("   📝 %s\n", description)
					fmt.Printf("   ⏰ Schedule: %s\n", schedule)
					fmt.Printf("   📊 Runs: %d | Errors: %d\n", runCount, errorCount)
					if lastRun != "" {
						fmt.Printf("   🕐 Last run: %s\n", lastRun)
					}
					if nextRun != "" {
						fmt.Printf("   ⏭️  Next run: %s\n", nextRun)
					}
					fmt.Println()
				}
			}
		} else {
			fmt.Println("No tasks found")
		}
	} else {
		fmt.Println("Invalid response format")
	}
}

func getSchedulerStatus(cmd *cobra.Command, args []string) {
	if apiKey != "" {
		// Make API call to get status
		status, err := makeAPIRequest("GET", "/api/scheduler/status", nil)
		if err != nil {
			fmt.Printf("❌ Cannot connect to Base Framework server at %s\n", baseURL)
			fmt.Printf("Error: %v\n\n", err)
			showConnectionHelp()
			return
		}
		displayStatus(status)
	} else {
		fmt.Println("📊 Scheduler Status")
		fmt.Println("💡 Provide API key and server URL to get real-time status:")
		fmt.Printf("   base scheduler status --api-key=your-key --url=%s\n", baseURL)
		showEnvironmentExamples()
	}
}

func displayStatus(response map[string]interface{}) {
	if data, ok := response["data"]; ok {
		if status, ok := data.(map[string]interface{}); ok {
			running := getBoolValue(status, "running")
			totalTasks := getIntValue(status, "total_tasks")
			enabledTasks := getIntValue(status, "enabled_tasks")
			disabledTasks := getIntValue(status, "disabled_tasks")
			checkInterval := getStringValue(status, "check_interval")

			runningStatus := "🔴 Stopped"
			if running {
				runningStatus = "🟢 Running"
			}

			fmt.Printf("\n📊 Scheduler Status: %s\n\n", runningStatus)
			fmt.Printf("📈 Statistics:\n")
			fmt.Printf("   📋 Total tasks: %d\n", totalTasks)
			fmt.Printf("   🟢 Enabled: %d\n", enabledTasks)
			fmt.Printf("   🔴 Disabled: %d\n", disabledTasks)
			if checkInterval != "" {
				fmt.Printf("   ⏱️  Check interval: %s\n", checkInterval)
			}
			fmt.Println()
		}
	}
}

// Helper functions for type assertions
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// Helper functions for better error handling and user guidance
func showConnectionHelp() {
	fmt.Println("💡 Possible solutions:")
	fmt.Println("   1. Check if the scheduler module is enabled in your Base Framework")
	fmt.Println("   2. Verify the API endpoint exists:")
	fmt.Printf("      curl -H \"X-Api-Key: api\" %s/api/scheduler/status\n", baseURL)
	fmt.Println("   3. Check if server is running on a different port:")
	fmt.Printf("      base scheduler status --url=http://localhost:8080 --api-key=your-key\n")
	fmt.Println("   4. For Docker deployments:")
	fmt.Printf("      base scheduler status --url=http://your-container:8100 --api-key=your-key\n")
	fmt.Println("   5. For remote deployments:")
	fmt.Printf("      base scheduler status --url=https://your-domain.com --api-key=your-key\n")
}

func showEnvironmentExamples() {
	fmt.Println("\n🌐 Environment examples:")
	fmt.Println("   Local:  --url=http://localhost:8100")
	fmt.Println("   Docker: --url=http://container-name:8100")
	fmt.Println("   Remote: --url=https://your-domain.com")
}
