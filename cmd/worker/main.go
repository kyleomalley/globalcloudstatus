package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/globalcloudstatus/globalcloudstatus/internal/providers/aws"
	"github.com/globalcloudstatus/globalcloudstatus/internal/providers/coreweave"
	"github.com/globalcloudstatus/globalcloudstatus/internal/types"
)

func run(outputPath string) error {
	log.Println("Fetching cloud provider statuses...")
	start := time.Now()

	var (
		awsRegions []types.RegionStatusData
		cwRegions  []types.RegionStatusData
		wg         sync.WaitGroup
	)

	wg.Add(2)
	go func() { defer wg.Done(); awsRegions = aws.FetchAllRegions() }()
	go func() { defer wg.Done(); cwRegions = coreweave.FetchAllRegions() }()
	wg.Wait()

	now := time.Now().UTC()
	out := types.StatusOutput{
		GeneratedAt: now,
		Providers: []types.ProviderOutput{
			{Provider: "aws", UpdatedAt: now, Regions: awsRegions},
			{Provider: "coreweave", UpdatedAt: now, Regions: cwRegions},
		},
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	log.Printf("Wrote %d AWS + %d CoreWeave regions to %s in %s",
		len(awsRegions), len(cwRegions), outputPath, time.Since(start).Round(time.Millisecond))
	return nil
}

func main() {
	outputPath := flag.String("output", "web/public/data/status.json", "Output JSON file path")
	watch := flag.Bool("watch", false, "Run continuously on an interval")
	interval := flag.Duration("interval", 10*time.Minute, "Polling interval (watch mode)")
	flag.Parse()

	if err := run(*outputPath); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if *watch {
		log.Printf("Watch mode: refreshing every %s", *interval)
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := run(*outputPath); err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}
}
