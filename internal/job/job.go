package job

import (
	"linkding-pdf-archiver/internal/linkding"
	"linkding-pdf-archiver/internal/pdf"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

func ProcessBookmarks(client *linkding.Client, config JobConfiguration) (err error) {
	logger := slog.With("tags", config.Tags, "bundleId", config.BundleId, "isDryRun", config.IsDryRun)

	bookmarks, err := getBookmarks(client, config)
	if err != nil {
		return
	}

	if len(bookmarks) == 0 {
		logger.Info("No bookmarks to process")
		return
	}

	logger.Info("Processing bookmarks", "count", len(bookmarks))

	var wg sync.WaitGroup
	succeeded := make(chan linkding.Bookmark, len(bookmarks))
	failed := make(chan linkding.Bookmark, len(bookmarks))

	for _, bookmark := range bookmarks {
		if !pdf.IsPDF(bookmark.Url) {
			logger.Debug("Skipping non-PDF URL", "url", bookmark.Url)
			continue
		}
		path, err := downloadPDF(client, bookmark)

		if err != nil {
			failed <- bookmark
			continue
		}
		if path == "" {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := uploadPDF(client, bookmark, path, config.IsDryRun); err != nil {
				failed <- bookmark
				return
			}

			succeeded <- bookmark
		}()
	}

	wg.Wait()

	logger.Info("Done processing bookmarks", "succeeded", len(succeeded), "failed", len(failed))

	return
}

func getBookmarks(client *linkding.Client, config JobConfiguration) ([]linkding.Bookmark, error) {
	bookmarks := make([]linkding.Bookmark, 0, 100)

	tags := config.Tags
	if len(tags) == 0 {
		tags = []string{""}
	}

	for _, tag := range tags {
		query := linkding.BookmarksQuery{Tag: tag, BundleId: config.BundleId, ModifiedSince: config.LastScan}
		bookmarksForTag, err := client.GetBookmarks(query)

		if err != nil {
			return nil, err
		}

		for _, bookmarkForTag := range bookmarksForTag {
			exists := slices.ContainsFunc(bookmarks, func(b linkding.Bookmark) bool { return b.Id == bookmarkForTag.Id })

			if !exists {
				bookmarks = append(bookmarks, bookmarkForTag)
			}
		}
	}

	return bookmarks, nil
}

func downloadPDF(client *linkding.Client, bookmark linkding.Bookmark) (string, error) {
	logger := slog.With("bookmarkId", bookmark.Id)

	assets, err := client.GetBookmarkAssets(bookmark.Id)
	if err != nil {
		logger.Error("Failed to fetch bookmark assets")
		return "", err
	}

	assetIndex := slices.IndexFunc(assets, func(asset linkding.Asset) bool {
		return asset.AssetType == "upload" && linkding.IsKnownMimeType(asset.ContentType)
	})
	if assetIndex > -1 {
		logger.Info("PDF asset already exists", "assetId", assets[assetIndex].Id)
		return "", nil
	}

	logger.Info("Downloading PDF")
	path, err := pdf.Download(bookmark.Url)

	if err != nil {
		logger.Error("Failed to download PDF", "error", err)
		return "", err
	}

	logger.Info("PDF downloaded successfully", "path", path)
	return path, nil
}

func uploadPDF(client *linkding.Client, bookmark linkding.Bookmark, path string, isDryRun bool) error {
	logger := slog.With("bookmarkId", bookmark.Id, "isDryRun", isDryRun)

	logger.Info("Adding asset", "path", path)

	// Always clean up temp directory when done
	defer os.RemoveAll(filepath.Dir(path))

	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open PDF file", "path", path, "error", err)
		return err
	}
	defer file.Close()

	asset, err := uploadAsset(client, bookmark, file, isDryRun)
	if err != nil {
		logger.Error("Failed to add asset", "path", path, "error", err)
		return err
	}
	logger.Info("Asset added successfully", "path", path, "assetId", asset.Id)

	return nil
}

func uploadAsset(client *linkding.Client, bookmark linkding.Bookmark, file *os.File, isDryRun bool) (*linkding.Asset, error) {
	if isDryRun {
		mimeType, err := linkding.GetMimeType(file.Name())
		if err != nil {
			return nil, err
		}

		asset := &linkding.Asset{Id: -1, AssetType: "upload", ContentType: mimeType, DisplayName: "Simulated Asset" + filepath.Ext(file.Name())}
		return asset, nil
	}

	return client.AddBookmarkAsset(bookmark.Id, file)
}
