package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionPattern(t *testing.T) {
	for _, value := range []string{"v1.0.0", "20260722-120000", "release_candidate"} {
		if !versionPattern.MatchString(value) {
			t.Fatalf("expected valid version: %s", value)
		}
	}
	for _, value := range []string{"", "release latest", "/v1"} {
		if versionPattern.MatchString(value) {
			t.Fatalf("expected invalid version: %s", value)
		}
	}
}

func TestAddBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	if err := addBytes(writer, "release/test.txt", []byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if len(reader.File) != 1 || reader.File[0].Name != "release/test.txt" {
		t.Fatalf("unexpected archive entries: %+v", reader.File)
	}
}
