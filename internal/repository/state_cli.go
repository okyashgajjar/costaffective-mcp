package repository

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func PromptIndex(info *StateInfo) bool {
	fmt.Println()
	fmt.Printf("Repository detected: %s\n", getRepoRootDisplay())
	fmt.Println()
	fmt.Println("This repository has not been indexed.")
	fmt.Println()
	fmt.Println("Indexing will:")
	fmt.Println("✓ Build Symbol Database")
	fmt.Println("✓ Build Reference Graph")
	fmt.Println("✓ Build Call Graph")
	fmt.Println("✓ Enable Knowledge Memory")
	fmt.Println("✓ Improve retrieval accuracy")
	fmt.Println("✓ Reduce token usage")
	fmt.Print("\nIndex repository now? [Y/n]: ")

	return promptYesNo(true)
}

func PromptReindex(info *StateInfo) bool {
	fmt.Println()
	fmt.Println("Repository index is out of date.")
	fmt.Printf("\nChanged Files: %d\n", info.ChangedFiles)
	fmt.Printf("Deleted Files: %d\n", info.DeletedFiles)
	fmt.Print("\nRefresh repository index? [Y/n]: ")

	return promptYesNo(true)
}

func ShowIndexingStart() {
	fmt.Println("\nIndexing repository...")
}

func ShowIndexingComplete(changed, skipped, deleted, total int) {
	fmt.Printf("\n✓ %d files indexed\n", total)
	fmt.Printf("✓ Symbol database created\n")
	fmt.Printf("✓ Reference graph created\n")
	fmt.Printf("✓ Call graph created\n")
	fmt.Printf("✓ Knowledge store initialized\n")
	fmt.Println("\nRepository ready.")
}

func ShowReindexStart() {
	fmt.Println("\nRunning incremental reindex...")
}

func ShowReindexComplete(changed, skipped, deleted int) {
	fmt.Printf("\n✓ Changed: %d\n", changed)
	fmt.Printf("✓ Skipped: %d\n", skipped)
	fmt.Printf("✓ Deleted: %d\n", deleted)
	fmt.Println("\nRepository ready.")
}

func ShowRepositoryReady(info *StateInfo) {
	fmt.Println("\nRepository ready.")
	fmt.Printf("Indexed Files: %d\n", info.IndexedFiles)
	ageStr := "just now"
	if info.IndexAge > time.Minute {
		ageStr = fmt.Sprintf("%.0f minutes", info.IndexAge.Minutes())
	}
	fmt.Printf("Index Age: %s\n", ageStr)
}

func ShowSkippingIndex() {
	fmt.Println("\nRunning without repository index.")
	fmt.Println("Retrieval quality may be reduced.")
	fmt.Println()
}

func ShowStaleSkipped() {
	fmt.Println("\nContinuing with stale index.")
	fmt.Println("Results may not reflect latest code.")
	fmt.Println()
}

func ShowAgentAutoIndex() {
	fmt.Println("\nRepository not indexed.")
	fmt.Println()
	fmt.Println("Building repository index for Agent mode...")
}

func ShowAgentReady() {
	fmt.Println("\n✓ Index complete")
	fmt.Println("\nStarting agent...")
}

var getRepoRootDisplay = func() string {
	cwd, _ := os.Getwd()
	return cwd
}

func promptYesNo(defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes
	}
	return input == "y" || input == "yes"
}
