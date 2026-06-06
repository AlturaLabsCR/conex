package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const interReleaseURL = "https://github.com/rsms/inter/releases/download/v4.1/Inter-4.1.zip"

var siteCardFonts = map[string]string{
	"extras/ttf/Inter-Bold.ttf":             "Inter-Bold.ttf",
	"extras/ttf/Inter-Regular.ttf":          "Inter-Regular.ttf",
	"extras/ttf/InterDisplay-ExtraBold.ttf": "InterDisplay-ExtraBold.ttf",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	res, err := http.Get(interReleaseURL)
	if err != nil {
		return fmt.Errorf("download Inter fonts: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download Inter fonts: unexpected status %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read Inter font archive: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("open Inter font archive: %w", err)
	}

	outDir := filepath.Join("assets", "fonts")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create font output directory: %w", err)
	}

	found := make(map[string]bool, len(siteCardFonts))
	for _, file := range zr.File {
		outName, ok := siteCardFonts[file.Name]
		if !ok {
			continue
		}
		if err := extractZipFile(file, filepath.Join(outDir, outName)); err != nil {
			return err
		}
		found[file.Name] = true
	}

	for source := range siteCardFonts {
		if !found[source] {
			return fmt.Errorf("Inter font archive did not contain %s", source)
		}
	}

	return nil
}

func extractZipFile(file *zip.File, outPath string) error {
	in, err := file.Open()
	if err != nil {
		return fmt.Errorf("open %s from Inter archive: %w", file.Name, err)
	}
	defer in.Close()

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	return nil
}
