package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	nodeImage  = "node:24-alpine"
	goImage    = "golang:1.26-alpine"
	mysqlImage = "mysql:8.4"
)

var versionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

type manifest struct {
	FormatVersion        int    `json:"formatVersion"`
	Version              string `json:"version"`
	Platform             string `json:"platform"`
	AppImage             string `json:"appImage"`
	MySQLImage           string `json:"mysqlImage"`
	CreatedAt            string `json:"createdAt"`
	SourceRevision       string `json:"sourceRevision"`
	SourceDirty          bool   `json:"sourceDirty"`
	ContainsInstanceData bool   `json:"containsInstanceData"`
	ImagesSHA256         string `json:"imagesSha256"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	defaultVersion := time.Now().Format("20060102-150405")
	version := flag.String("version", defaultVersion, "release version")
	platform := flag.String("platform", "linux/amd64", "target platform: linux/amd64 or linux/arm64")
	output := flag.String("output", "", "output ZIP path")
	flag.Parse()

	if !versionPattern.MatchString(*version) {
		return errors.New("invalid version; use letters, digits, dots, underscores, and hyphens")
	}
	if *platform != "linux/amd64" && *platform != "linux/arm64" {
		return fmt.Errorf("unsupported platform: %s", *platform)
	}
	root, err := projectRoot()
	if err != nil {
		return err
	}
	if err := requireCommands("docker", "git"); err != nil {
		return err
	}
	if err := commandQuiet(root, "docker", "info"); err != nil {
		return errors.New("Docker is not running")
	}

	arch := strings.ReplaceAll(*platform, "/", "-")
	outputPath := *output
	if outputPath == "" {
		outputPath = filepath.Join(root, "dist", fmt.Sprintf("bbs-go-%s-%s.zip", *version, arch))
	} else if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(root, outputPath)
	}
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("output already exists: %s", outputPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	appImage := "bbs-go:" + strings.ToLower(*version)
	for _, image := range []string{nodeImage, goImage, mysqlImage} {
		if err := command(root, "docker", "pull", "--platform", *platform, image); err != nil {
			return err
		}
	}
	if err := command(root, "docker", "build", "--platform", *platform, "--tag", appImage, "."); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp(filepath.Dir(outputPath), ".bbs-go-package-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	imagesPath := filepath.Join(tempDir, "docker-images.tar")
	if err := command(root, "docker", "image", "save", "--output", imagesPath, appImage, mysqlImage); err != nil {
		return err
	}

	if err := writeRelease(root, outputPath, imagesPath, *version, *platform, appImage); err != nil {
		_ = os.Remove(outputPath)
		return err
	}
	digest, err := fileSHA256(outputPath)
	if err != nil {
		return err
	}
	fmt.Printf("\nRelease package created:\n%s\nSHA-256: %s\n", outputPath, digest)
	return nil
}

func projectRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for current := workingDir; ; current = filepath.Dir(current) {
		if fileExists(filepath.Join(current, "go.mod")) && fileExists(filepath.Join(current, "Dockerfile")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("run from the BBS-GO project tree")
		}
	}
}

func requireCommands(names ...string) error {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("required command was not found: %s", name)
		}
	}
	return nil
}

func command(dir, name string, args ...string) error {
	fmt.Printf("$ %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func commandOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

func commandQuiet(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func writeRelease(root, outputPath, imagesPath, version, platform, appImage string) error {
	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(output)
	closed := false
	defer func() {
		if !closed {
			_ = archive.Close()
			_ = output.Close()
		}
	}()
	closeArchive := func() error {
		if err := archive.Close(); err != nil {
			_ = output.Close()
			return err
		}
		if err := output.Close(); err != nil {
			return err
		}
		closed = true
		return nil
	}

	prefix := fmt.Sprintf("bbs-go-%s/", version)
	if err := addFile(archive, prefix+"docker-images.tar", imagesPath, zip.Store); err != nil {
		return err
	}
	compose, err := os.ReadFile(filepath.Join(root, "deploy", "docker-compose.yml"))
	if err != nil {
		return err
	}
	envTemplate, err := os.ReadFile(filepath.Join(root, "deploy", ".env.example"))
	if err != nil {
		return err
	}
	envTemplate = bytes.ReplaceAll(envTemplate, []byte("replace-with-image:v1.0.0"), []byte(appImage))
	readme, err := os.ReadFile(filepath.Join(root, "deploy", "README.md"))
	if err != nil {
		return err
	}
	for name, data := range map[string][]byte{
		prefix + "docker-compose.yml": compose,
		prefix + ".env.example":       envTemplate,
		prefix + "README.md":          readme,
	} {
		if err := addBytes(archive, name, data); err != nil {
			return err
		}
	}

	sourcePaths, err := gitSourceFiles(root)
	if err != nil {
		return err
	}
	for _, relative := range sourcePaths {
		if err := addFile(archive, prefix+"source/"+filepath.ToSlash(relative), filepath.Join(root, relative), zip.Deflate); err != nil {
			return err
		}
	}

	revision, _ := commandOutput(root, "git", "rev-parse", "HEAD")
	status, statusErr := commandOutput(root, "git", "status", "--porcelain")
	imagesDigest, err := fileSHA256(imagesPath)
	if err != nil {
		return err
	}
	metadata := manifest{
		FormatVersion: 1, Version: version, Platform: platform,
		AppImage: appImage, MySQLImage: mysqlImage,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), SourceRevision: revision,
		SourceDirty: statusErr != nil || status != "", ContainsInstanceData: false,
		ImagesSHA256: imagesDigest,
	}
	manifestData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	if err := addBytes(archive, prefix+"manifest.json", append(manifestData, '\n')); err != nil {
		return err
	}
	return closeArchive()
}

func gitSourceFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(splitNull)
	for scanner.Scan() {
		relative := scanner.Text()
		info, err := os.Stat(filepath.Join(root, relative))
		if err == nil && info.Mode().IsRegular() {
			paths = append(paths, relative)
		}
	}
	sort.Strings(paths)
	return paths, scanner.Err()
}

func splitNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if index := bytes.IndexByte(data, 0); index >= 0 {
		return index + 1, data[:index], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func addFile(archive *zip.Writer, name, path string, method uint16) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = method
	header.SetMode(0o644)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, input)
	return err
}

func addBytes(archive *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(0o644)
	header.Modified = time.Now()
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
