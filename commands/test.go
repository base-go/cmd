package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	htmlFlag bool
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// colorizeTestOutput adds colors to Go test output
func colorizeTestOutput(line string) string {
	switch {
	case strings.Contains(line, "--- PASS:"):
		return colorGreen + line + colorReset
	case strings.Contains(line, "--- FAIL:"):
		return colorRed + line + colorReset
	case strings.Contains(line, "=== RUN"):
		return colorBlue + line + colorReset
	case strings.Contains(line, "PASS") && !strings.Contains(line, "---"):
		return colorGreen + line + colorReset
	case strings.Contains(line, "FAIL") && !strings.Contains(line, "---"):
		return colorRed + line + colorReset
	case strings.Contains(line, "Error Trace:"):
		return colorRed + line + colorReset
	case strings.Contains(line, "expected:"):
		return colorYellow + line + colorReset
	case strings.Contains(line, "actual:"):
		return colorPurple + line + colorReset
	case strings.Contains(line, "Test:"):
		return colorCyan + line + colorReset
	case strings.HasPrefix(line, "?"):
		return colorGray + line + colorReset
	case strings.HasPrefix(line, "ok"):
		return colorGreen + line + colorReset
	case strings.Contains(line, "coverage:"):
		return colorCyan + line + colorReset
	default:
		return line
	}
}

var testCmd = &cobra.Command{
	Use:   "test [module]",
	Short: "Run tests for the application",
	Long: `Run tests for different modules of the application.

Examples:
  base test           # Run all tests
  base test app       # Run app module tests
  base test core      # Run core module tests
  base test coverage  # Run all tests with coverage
  base test coverage app      # Run app tests with coverage
  base test coverage core     # Run core tests with coverage
  base test coverage --html   # Run all tests with HTML coverage report`,
	Args: cobra.MaximumNArgs(2),
	RunE: runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	// Determine if we're running coverage
	isCoverage := false
	module := ""

	if len(args) > 0 {
		if args[0] == "coverage" {
			isCoverage = true
			if len(args) > 1 {
				module = args[1]
			}
		} else {
			module = args[0]
		}
	}

	// Check if --html flag is used with coverage
	if htmlFlag && !isCoverage {
		return fmt.Errorf("--html flag can only be used with coverage command")
	}

	// Build the go test command
	var testArgs []string
	var testPath string

	switch module {
	case "app":
		testPath = "./test/app_test/..."
	case "core":
		testPath = "./test/core_test/..."
	case "":
		testPath = "./test/..."
	default:
		return fmt.Errorf("unknown module: %s. Available modules: app, core", module)
	}

	if isCoverage {
		// Coverage command
		coverageFile := "coverage.out"
		if module != "" {
			coverageFile = fmt.Sprintf("%s_coverage.out", module)
		}

		testArgs = []string{
			"test",
			"-v",
			"-race",
			"-coverprofile=" + coverageFile,
			"-coverpkg=./...",
			testPath,
		}

		// Print beautiful header for coverage
		fmt.Printf("%s🧪 Running Base Framework Tests with Coverage%s\n", colorBold+colorWhite, colorReset)
		fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("=", 50), colorReset)
		fmt.Printf("📋 Module: %s%s%s\n", colorCyan, getModuleDescription(module), colorReset)
		fmt.Printf("📊 Coverage: %sEnabled%s\n\n", colorGreen, colorReset)

		// Run the test command
		if err := runGoCommand(testArgs); err != nil {
			return fmt.Errorf("tests failed: %v", err)
		}

		// Generate coverage report
		if err := generateCoverageReport(coverageFile, module); err != nil {
			return fmt.Errorf("failed to generate coverage report: %v", err)
		}

		// Generate HTML report if requested
		if htmlFlag {
			if err := generateHTMLCoverageReport(coverageFile, module); err != nil {
				return fmt.Errorf("failed to generate HTML coverage report: %v", err)
			}
		}

	} else {
		// Regular test command
		testArgs = []string{
			"test",
			"-v",
			"-race",
			testPath,
		}

		// Print beautiful header
		fmt.Printf("%s🧪 Running Base Framework Tests%s\n", colorBold+colorWhite, colorReset)
		fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("=", 50), colorReset)
		fmt.Printf("📋 Module: %s%s%s\n\n", colorCyan, getModuleDescription(module), colorReset)

		if err := runGoCommand(testArgs); err != nil {
			fmt.Printf("\n%s🚫 Tests failed%s\n", colorRed, colorReset)
			return fmt.Errorf("tests failed: %v", err)
		}
	}

	fmt.Printf("\n%s✅ Tests completed successfully!%s\n", colorGreen+colorBold, colorReset)
	return nil
}

func getModuleDescription(module string) string {
	switch module {
	case "app":
		return "app module"
	case "core":
		return "core module"
	case "":
		return "all modules"
	default:
		return module
	}
}

func runGoCommand(args []string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = "."

	// Create pipes to capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create channels to handle output from both pipes
	done := make(chan bool, 2)

	// Handle stdout
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(colorizeTestOutput(line))
		}
		done <- true
	}()

	// Handle stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(os.Stderr, colorizeTestOutput(line))
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done

	// Wait for the command to finish
	return cmd.Wait()
}

func generateCoverageReport(coverageFile, module string) error {
	// Generate text coverage report
	args := []string{"tool", "cover", "-func=" + coverageFile}

	fmt.Printf("\n📊 Coverage Report for %s:\n", getModuleDescription(module))
	fmt.Println(strings.Repeat("=", 50))

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func generateHTMLCoverageReport(coverageFile, module string) error {
	// Create test/coverage directory if it doesn't exist
	coverageDir := "test/coverage"
	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		return fmt.Errorf("failed to create coverage directory: %v", err)
	}

	// Generate HTML coverage report
	htmlFile := filepath.Join(coverageDir, "coverage.html")
	if module != "" {
		htmlFile = filepath.Join(coverageDir, fmt.Sprintf("%s_coverage.html", module))
	}

	args := []string{"tool", "cover", "-html=" + coverageFile, "-o", htmlFile}

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("\n🌐 HTML Coverage Report generated: %s\n", htmlFile)

	// Try to get absolute path for better user experience
	if absPath, err := filepath.Abs(htmlFile); err == nil {
		fmt.Printf("📂 Open in browser: file://%s\n", absPath)
	}

	return nil
}

func init() {
	// Add flags
	testCmd.Flags().BoolVar(&htmlFlag, "html", false, "Generate HTML coverage report (only with coverage)")

	// Add the test command to root
	rootCmd.AddCommand(testCmd)
}
