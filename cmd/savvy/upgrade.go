package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/heyajulia/savvy/internal"
	"github.com/minio/selfupdate"
	"github.com/urfave/cli/v3"
)

const (
	repoOwner = "heyajulia"
	repoName  = "savvy"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func upgradeCommand() *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "Upgrade Savvy to the latest version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "check",
				Usage: "Check for updates without applying",
			},
			&cli.StringFlag{
				Name:  "target",
				Usage: "Upgrade to a specific version tag (default: latest)",
			},
		},
		Action: runUpgrade,
	}
}

func runUpgrade(ctx context.Context, c *cli.Command) error {
	checkOnly := c.Bool("check")
	targetVersion := c.String("target")

	release, err := getRelease(targetVersion)
	if err != nil {
		return fmt.Errorf("fetch release: %w", err)
	}

	if release.TagName == internal.Version {
		fmt.Printf("Already running latest version %s\n", internal.Version)
		return nil
	}

	fmt.Printf("Current version: %s\n", internal.Version)
	fmt.Printf("Available version: %s\n", release.TagName)

	if checkOnly {
		return nil
	}

	assetName := fmt.Sprintf("savvy_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	// Find and download checksums file
	var checksumsURL string
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			checksumsURL = asset.BrowserDownloadURL
			break
		}
	}

	if checksumsURL == "" {
		return fmt.Errorf("checksums.txt not found in release")
	}

	expectedChecksum, err := getExpectedChecksum(checksumsURL, assetName)
	if err != nil {
		return fmt.Errorf("get checksum: %w", err)
	}

	fmt.Printf("Downloading %s...\n", release.TagName)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Read archive into memory for checksum verification
	archiveData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(archiveData)
	actualChecksumHex := hex.EncodeToString(actualChecksum[:])
	if actualChecksumHex != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksumHex)
	}

	fmt.Println("Checksum verified.")

	binary, err := extractBinaryFromTarGz(bytes.NewReader(archiveData))
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	if err := selfupdate.Apply(binary, selfupdate.Options{}); err != nil {
		return fmt.Errorf("apply update: %w", err)
	}

	fmt.Printf("Successfully upgraded to %s\n", release.TagName)
	return nil
}

func getRelease(version string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	if version != "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", repoOwner, repoName, version)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getExpectedChecksum(checksumsURL, assetName string) (string, error) {
	resp, err := http.Get(checksumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksums: %s", resp.Status)
	}

	// Parse checksums.txt format: "sha256sum  filename"
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			return parts[0], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("checksum not found for %s", assetName)
}

func extractBinaryFromTarGz(r io.Reader) (io.Reader, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "savvy") {
			content, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read binary: %w", err)
			}
			return bytes.NewReader(content), nil
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}
