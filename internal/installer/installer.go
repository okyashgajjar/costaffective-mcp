package installer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Installer struct {
	Build     bool
	All       bool
	DryRun    bool
	TargetID  string
	Location  Location
	Uninstall bool
	Yes       bool
	Repair    bool
}

func (inst *Installer) Run() error {
	if inst.Uninstall {
		return inst.runUninstall()
	}
	if inst.Repair {
		return inst.runRepair()
	}
	return inst.runInstall()
}

func (inst *Installer) runInstall() error {
	// 1. Install (build + copy) binary to PATH
	if inst.Build && !inst.DryRun {
		fmt.Println("Building and installing CostAffective binary...")
		installedPath, err := InstallBinary()
		if err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
		fmt.Printf("✓ Installed to %s\n", Tildify(installedPath))
	} else if inst.DryRun {
		fmt.Printf("  [DRY RUN] Would build and install to %s\n", Tildify(DefaultBinaryPath()))
	} else {
		// Binary already installed; verify
		result := CheckBinary()
		if !result.Exists {
			return ActionableError{
				Message: "CostAffective binary not found.",
				Action:  "Run: costaffective install",
			}
		}
		SetBinaryPath(result.Path)
		fmt.Printf("✓ Found binary at %s\n", Tildify(result.Path))
	}

	// 2. Resolve targets
	targets, err := inst.resolveTargets()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Println("No agent targets selected — nothing to do.")
		return nil
	}

	// 3. Resolve location (global vs local)
	loc, err := inst.resolveLocation(targets)
	if err != nil {
		return err
	}

	// 4. Auto-allow (only relevant for Claude)
	autoAllow := inst.resolveAutoAllow(targets)

	// 5. Install for each target
	for _, t := range targets {
		if !t.SupportsLocation(loc) {
			fmt.Printf("  %s: skipped — does not support %s location\n", t.DisplayName(), loc)
			continue
		}

		if inst.DryRun {
			fmt.Printf("  [DRY RUN] Would configure %s\n", t.DisplayName())
			for _, p := range t.DescribePaths(loc) {
				fmt.Printf("    → %s\n", Tildify(p))
			}
			continue
		}

		results := t.Install(loc, InstallOptions{AutoAllow: autoAllow})
		for _, r := range results {
			verb := map[string]string{
				"created":   "Created",
				"updated":   "Updated",
				"unchanged": "Unchanged",
				"removed":   "Removed",
				"not-found": "Not found",
			}[r.Action]
			if verb == "" {
				verb = r.Action
			}
			fmt.Printf("  %s: %s %s\n", t.DisplayName(), verb, Tildify(r.Path))
		}
	}

	return nil
}

func (inst *Installer) runRepair() error {
	fmt.Println("CostAffective Repair Mode")
	fmt.Println()

	// 1. Build and install binary
	fmt.Println("1. Installing binary...")
	installedPath, err := InstallBinary()
	if err != nil {
		return fmt.Errorf("binary installation failed: %w", err)
	}
	fmt.Printf("   ✓ %s\n", Tildify(installedPath))

	// 2. Verify binary
	fmt.Println("2. Verifying binary...")
	if err := VerifyBinary(installedPath); err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}
	fmt.Println("   ✓ Executable OK")
	fmt.Println("   ✓ Launches OK")
	fmt.Println()

	// 3. Detect all clients and replace stale configs
	fmt.Println("3. Checking MCP configurations...")
	allTargets := AllTargets()
	fixed := 0
	skipped := 0
	for _, t := range allTargets {
		d := t.Detect(LocationGlobal)
		if d.AlreadyConfigured {
			results := t.Install(LocationGlobal, InstallOptions{AutoAllow: true})
			for _, r := range results {
				if r.Action == "updated" {
					fmt.Printf("   ✓ %s: repaired %s\n", t.DisplayName(), Tildify(r.Path))
					fixed++
				} else if r.Action == "unchanged" {
					fmt.Printf("   ✓ %s: OK (%s)\n", t.DisplayName(), Tildify(r.Path))
					skipped++
				} else {
					fmt.Printf("   ✓ %s: %s %s\n", t.DisplayName(), r.Action, Tildify(r.Path))
					fixed++
				}
			}
		} else if d.Installed {
			fmt.Printf("   - %s: detected but not configured. Skipping.\n", t.DisplayName())
			fmt.Printf("     Run: costaffective install --target %s\n", t.ID())
		}
	}
	if fixed+skipped == 0 {
		fmt.Println("   No MCP configurations found. Nothing to repair.")
	}
	fmt.Println()

	fmt.Println("Repair complete.")
	return nil
}

func (inst *Installer) resolveTargets() ([]Target, error) {
	if inst.TargetID != "" {
		t := GetTarget(inst.TargetID)
		if t == nil {
			return nil, fmt.Errorf("unknown target: %s", inst.TargetID)
		}
		return []Target{t}, nil
	}

	if inst.All {
		return AllTargets(), nil
	}

	if inst.Yes {
		detected := DetectAll(LocationGlobal)
		var result []Target
		for t, d := range detected {
			if d.Installed {
				result = append(result, t)
			}
		}
		if len(result) == 0 {
			fmt.Println("No supported AI coding clients detected.")
			return nil, nil
		}
		return result, nil
	}

	return inst.promptTargets()
}

func (inst *Installer) promptTargets() ([]Target, error) {
	detected := DetectAll(LocationGlobal)
	allTargets := AllTargets()

	fmt.Println()
	fmt.Println("Detected AI coding clients:")

	type item struct {
		target Target
		detect DetectionResult
	}
	var items []item
	installedCount := 0

	for _, t := range allTargets {
		d := detected[t]
		items = append(items, item{t, d})
		if d.Installed {
			installedCount++
		}
	}

	if installedCount == 0 {
		fmt.Println("  (none detected)")
		fmt.Println()
		fmt.Println("You can still install for specific clients:")
		fmt.Println("  costaffective install --target claude")
		fmt.Println("  costaffective install --all")
		return nil, nil
	}

	for i, it := range items {
		status := ""
		if it.detect.Installed {
			status = "detected"
		}
		if it.detect.AlreadyConfigured {
			status = "configured"
		}
		if status != "" {
			fmt.Printf("  %d. %-20s (%s)\n", i+1, it.target.DisplayName(), status)
		} else {
			fmt.Printf("  %d. %-20s (not found)\n", i+1, it.target.DisplayName())
		}
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Install for which clients? (comma-separated numbers, 'all', 'none') [default: all]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			input = "all"
		}

		if input == "none" || input == "n" {
			return nil, nil
		}

		if input == "all" || input == "a" {
			var result []Target
			for _, it := range items {
				if it.detect.Installed {
					result = append(result, it.target)
				}
			}
			if len(result) == 0 {
				fmt.Println("No detected clients to install for.")
				return nil, nil
			}
			fmt.Printf("Selected: ")
			for i, t := range result {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(t.DisplayName())
			}
			fmt.Println()
			return result, nil
		}

		parts := strings.Split(input, ",")
		var selected []Target
		valid := true
		for _, p := range parts {
			p = strings.TrimSpace(p)
			num, err := strconv.Atoi(p)
			if err != nil || num < 1 || num > len(items) {
				fmt.Printf("  Invalid number: %s\n", p)
				valid = false
				break
			}
			selected = append(selected, items[num-1].target)
		}
		if valid && len(selected) > 0 {
			return selected, nil
		}
	}
}

func (inst *Installer) resolveLocation(targets []Target) (Location, error) {
	if inst.Location != "" {
		return inst.Location, nil
	}
	if inst.DryRun || inst.Yes {
		return LocationGlobal, nil
	}

	allGlobalOnly := true
	for _, t := range targets {
		if t.SupportsLocation(LocationLocal) {
			allGlobalOnly = false
			break
		}
	}
	if allGlobalOnly {
		fmt.Println("  (writing user-wide configs — selected clients have no project-local config)")
		return LocationGlobal, nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Apply configs to all your projects, or just this one? (global/local) [default: global]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return LocationGlobal, nil
		}
		if input == "global" || input == "g" || input == "all" {
			return LocationGlobal, nil
		}
		if input == "local" || input == "l" || input == "project" {
			return LocationLocal, nil
		}
		fmt.Println("  Please enter 'global' or 'local'.")
	}
}

func (inst *Installer) resolveAutoAllow(targets []Target) bool {
	if inst.Yes || inst.DryRun {
		return true
	}

	hasClaude := false
	for _, t := range targets {
		if t.ID() == "claude" {
			hasClaude = true
			break
		}
	}
	if !hasClaude {
		return false
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Auto-allow CostAffective commands in Claude Code? (Y/n) [default: Y]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" || input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
	}
}

func (inst *Installer) runUninstall() error {
	var targets []Target
	if inst.TargetID != "" {
		t := GetTarget(inst.TargetID)
		if t == nil {
			return fmt.Errorf("unknown target: %s", inst.TargetID)
		}
		targets = []Target{t}
	} else if inst.All {
		targets = AllTargets()
	} else if inst.Yes {
		detected := DetectAll(inst.Location)
		for t, d := range detected {
			if d.AlreadyConfigured {
				targets = append(targets, t)
			}
		}
	} else {
		detected := DetectAll(inst.Location)
		for t, d := range detected {
			if d.AlreadyConfigured {
				targets = append(targets, t)
			}
		}
		if len(targets) == 0 {
			fmt.Println("No configured costaffective MCP entries found in any client.")
			return nil
		}
		fmt.Println("Found costaffective MCP configs in:")
		for _, t := range targets {
			fmt.Printf("  - %s\n", t.DisplayName())
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Remove these configs? (Y/n) [default: Y]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "n" || input == "no" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	for _, t := range targets {
		if !t.SupportsLocation(inst.Location) {
			fmt.Printf("  %s: skipped — does not support %s location\n", t.DisplayName(), inst.Location)
			continue
		}

		if inst.DryRun {
			fmt.Printf("  [DRY RUN] Would uninstall %s from %s\n", t.DisplayName(), inst.Location)
			for _, p := range t.DescribePaths(inst.Location) {
				fmt.Printf("    → %s\n", Tildify(p))
			}
			continue
		}

		results := t.Uninstall(inst.Location)
		for _, r := range results {
			verb := map[string]string{
				"removed":   "Removed",
				"not-found": "Not found",
				"unchanged": "Unchanged",
				"kept":      "Kept",
			}[r.Action]
			if verb == "" {
				verb = r.Action
			}
			if r.Action != "not-found" && r.Action != "unchanged" {
				fmt.Printf("  %s: %s %s\n", t.DisplayName(), verb, Tildify(r.Path))
			}
		}
	}

	return nil
}
