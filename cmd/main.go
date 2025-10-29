package main

import (
	"flag"
	"linkding-pdf-archiver/internal/job"
	"linkding-pdf-archiver/internal/linkding"
	"linkding-pdf-archiver/internal/logging"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	isDryRun := flag.Bool("n", false, "Dry run: download PDFs but do not actually upload them to Linkding")
	isSingleRun := flag.Bool("s", false, "Single run: exit after processing bookmarks once")
	flag.Parse()

	logger := logging.NewLogger()
	slog.SetDefault(logger)

	client, err := linkding.NewClient(os.Getenv("LDPA_BASEURL"), os.Getenv("LDPA_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	tempdir, err := os.MkdirTemp(os.TempDir(), "pdfs")
	if err != nil {
		log.Fatal(err)
	}

	cleanupAndExit := func(code int) {
		os.RemoveAll(tempdir)
		os.Exit(code)
	}

	onInterrupt(cleanupAndExit)

	tags := getLinkdingTags()
	bundleId := getLinkdingBundleId()
	interval := getScanInterval()
	sleep := time.NewTicker(time.Duration(interval) * time.Second)

	var lastScan time.Time

	// Run immediately and then on every tick
	for ; true; <-sleep.C {
		timeBeforeRun := time.Now()

		config := job.JobConfiguration{Tags: tags, BundleId: bundleId, IsDryRun: *isDryRun, LastScan: lastScan}
		err := job.ProcessBookmarks(client, config)

		if err == nil {
			lastScan = timeBeforeRun // Only update last scan time when bookmarks were actually processed
		} else {
			logger.Error("Error processing bookmarks", "error", err)
		}

		if *isSingleRun {
			cleanupAndExit(0)
		}

		logger.Info("Waiting for next scan", "scanInterval", interval)
	}
}

func getLinkdingTags() []string {
	tagsEnv := os.Getenv("LDPA_TAGS")
	return strings.Fields(tagsEnv)
}

func getLinkdingBundleId() int {
	bundleId, err := strconv.Atoi(os.Getenv("LDPA_BUNDLE_ID"))

	if bundleId <= 0 || err != nil {
		bundleId = 0
	}

	return bundleId
}

func getScanInterval() int {
	interval, err := strconv.Atoi(os.Getenv("LDPA_SCAN_INTERVAL"))

	if interval <= 0 || err != nil {
		interval = 3600
	}

	return interval
}

func onInterrupt(cleanup func(int)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cleanup(1)
	}()
}
