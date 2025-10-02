package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Flow represents the structure of the JSON log entries
type Flow struct {
	Source struct {
		Identity int `json:"identity"`
	} `json:"source"`
	Destination struct {
		Identity int `json:"identity"`
	} `json:"destination"`
	L4 struct {
		TCP struct {
			DestinationPort int `json:"destination_port"`
		} `json:"TCP"`
	} `json:"l4"`
}

func main() {
	// Map to track unique flows
	uniqueFlows := make(map[string]string)
	
	// Get all log files in the current directory
	logFiles, err := filepath.Glob("*.log")
	if err != nil {
		fmt.Printf("Error finding log files: %v\n", err)
		os.Exit(1)
	}
	
	if len(logFiles) == 0 {
		fmt.Println("No log files found in the current directory")
		os.Exit(1)
	}
	
	// Process each log file
	for _, logFile := range logFiles {
		fmt.Fprintf(os.Stderr, "Processing file: %s\n", logFile)
		
		file, err := os.Open(logFile)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", logFile, err)
			continue
		}
		
		scanner := bufio.NewScanner(file)

		// Set a larger buffer size for potentially large lines
		const maxCapacity = 512 * 1024 // 512KB
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			line := scanner.Text()
			
			// Skip lines that don't contain flow data
			if !strings.Contains(line, "\"flow\":") {
				continue
			}

			var flowData struct {
				Flow Flow `json:"flow"`
			}

			if err := json.Unmarshal([]byte(line), &flowData); err != nil {
				// Skip lines that can't be parsed
				continue
			}

			// Create a unique key based on the specified fields
			key := fmt.Sprintf("%d-%d-%d", 
				flowData.Flow.Source.Identity, 
				flowData.Flow.Destination.Identity, 
				flowData.Flow.L4.TCP.DestinationPort)

			// Store only unique flows
			if _, exists := uniqueFlows[key]; !exists {
				uniqueFlows[key] = line
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading file %s: %v\n", logFile, err)
		}
		
		file.Close()
	}

	// Generate output filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	outputFile := fmt.Sprintf("unique-flows-%s.json", timestamp)
	
	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write unique flows to the output file
	for _, line := range uniqueFlows {
		fmt.Fprintln(outFile, line)
	}

	fmt.Fprintf(os.Stderr, "Total unique flows: %d\n", len(uniqueFlows))
	fmt.Fprintf(os.Stderr, "Unique flows written to: %s\n", outputFile)
}
