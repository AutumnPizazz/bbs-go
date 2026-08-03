package services

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"bbs-go/internal/pkg/config"
)

// buildTarGzWithDump creates a tar.gz containing dump.sql with the given bytes.
func buildTarGzWithDump(t *testing.T, dump []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bbs-go-test.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	header := &tar.Header{Name: backupArchiveSqlName, Mode: 0o644, Size: int64(len(dump))}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(dump); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestDatabaseBackupVerifyTarGzAcceptsSplitUTF8Boundary is a regression test
// for an intermittent backup failure: verification truncated dump.sql at the
// 64KB read limit and a multi-byte UTF-8 character split at the boundary made
// utf8.Valid fail, rejecting perfectly good archives.
func TestDatabaseBackupVerifyTarGzAcceptsSplitUTF8Boundary(t *testing.T) {
	service := &databaseBackupService{}

	// A real dump is ASCII for the header/DDL, then INSERT data. Build one
	// whose byte 65535 is the first byte of a 3-byte UTF-8 char ("测"), so
	// the 64KB truncation cuts the rune in half.
	ascii := make([]byte, 65515)
	for i := range ascii {
		ascii[i] = 'a'
	}
	dump := append([]byte("-- MySQL dump 10.13\n"), ascii...)
	dump = append(dump, 0xE6, 0xB5, 0x8B) // 测, 0xE6 lands at index 65535
	if len(dump) <= 64*1024 {
		t.Fatalf("test dump must exceed the 64KB verification window, got %d", len(dump))
	}
	// Sanity: the first 64KB really ends mid-rune.
	if utf8.Valid(dump[:64*1024]) {
		t.Fatal("test setup error: expected first 64KB to end with an incomplete rune")
	}

	path := buildTarGzWithDump(t, dump)
	checksum, err := fileChecksum(path)
	if err != nil {
		t.Fatalf("checksum archive: %v", err)
	}

	if err := service.verifyTarGz(path, config.DbTypeMySQL); err != nil {
		t.Fatalf("valid archive rejected: %v", err)
	}
	if err := service.verifyFile(path, config.DbTypeMySQL, checksum); err != nil {
		t.Fatalf("valid archive rejected by verifyFile: %v", err)
	}
}

// TestDatabaseBackupVerifyTarGzRejectsBinaryDump ensures genuinely corrupt
// dump content (NUL bytes / invalid UTF-8 in the middle) is still rejected.
func TestDatabaseBackupVerifyTarGzRejectsBinaryDump(t *testing.T) {
	service := &databaseBackupService{}

	header := []byte("-- MySQL dump 10.13\n")
	mid := make([]byte, 32*1024)
	for i := range mid {
		mid[i] = 'x'
	}
	for _, corrupt := range [][]byte{
		append(append(append([]byte{}, header...), mid...), 0x00, 0x01, 0x02), // NUL bytes
		append(append(append([]byte{}, header...), mid...), 0xFF, 0xFE),       // invalid UTF-8
	} {
		path := buildTarGzWithDump(t, corrupt)
		if err := service.verifyTarGz(path, config.DbTypeMySQL); err == nil {
			t.Fatalf("expected corrupt dump to be rejected")
		}
	}
}

// TestDatabaseBackupVerifyMySQLDumpAcceptsSplitUTF8Boundary is the plain
// .sql equivalent of the tar.gz regression test above.
func TestDatabaseBackupVerifyMySQLDumpAcceptsSplitUTF8Boundary(t *testing.T) {
	ascii := make([]byte, 65515)
	for i := range ascii {
		ascii[i] = 'a'
	}
	dump := append([]byte("-- MySQL dump 10.13\n"), ascii...)
	dump = append(dump, 0xE6, 0xB5, 0x8B)
	if !utf8.Valid(dump) || utf8.Valid(dump[:64*1024]) {
		t.Fatal("test setup error: expected truncated window to end mid-rune")
	}

	path := filepath.Join(t.TempDir(), "backup.sql")
	if err := os.WriteFile(path, dump, 0o600); err != nil {
		t.Fatalf("write dump: %v", err)
	}
	if err := verifyMySQLDump(path); err != nil {
		t.Fatalf("valid sql dump rejected: %v", err)
	}
}

// TestDatabaseBackupVerifyMySQLDumpRejectsCorruptContent ensures mid-content
// NUL bytes and invalid UTF-8 are still rejected for plain .sql dumps.
func TestDatabaseBackupVerifyMySQLDumpRejectsCorruptContent(t *testing.T) {
	header := []byte("-- MySQL dump 10.13\n")
	mid := make([]byte, 1024)
	for i := range mid {
		mid[i] = 'x'
	}
	for name, corrupt := range map[string][]byte{
		"nul":     append(append(append([]byte{}, header...), mid...), 0x00),
		"invalid": append(append(append([]byte{}, header...), mid...), 0xFF),
	} {
		path := filepath.Join(t.TempDir(), name+".sql")
		if err := os.WriteFile(path, corrupt, 0o600); err != nil {
			t.Fatalf("write dump: %v", err)
		}
		if err := verifyMySQLDump(path); err == nil {
			t.Fatalf("expected corrupt dump %q to be rejected", name)
		}
	}
}

// TestDatabaseBackupVerifyRejectsPlainTextWithoutDumpHeader keeps the
// recognized-header check working after the trimming change.
func TestDatabaseBackupVerifyRejectsPlainTextWithoutDumpHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.sql")
	content := []byte("this is not a SQL dump")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := verifyMySQLDump(path); err == nil || !strings.Contains(err.Error(), "recognized") {
		t.Fatalf("expected unrecognized dump header error, got %v", err)
	}
}

// TestUTF8ValidWindowEdgeCases pins down the tolerance boundary of
// utf8ValidWindow: incomplete runes that could still be completed are
// accepted, everything else is rejected.
func TestUTF8ValidWindowEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"ascii", []byte("-- MySQL dump 10.13\n"), true},
		{"complete chinese", []byte("abc\xE6\xB5\x8B"), true},
		{"truncated 3-byte start", []byte("abc\xE6"), true},
		{"truncated 3-byte mid", []byte("abc\xE6\xB5"), true},
		{"truncated 4-byte", []byte("abc\xF0\x9F\x98"), true},
		{"complete 4-byte", []byte("abc\xF0\x9F\x98\x80"), true},
		{"bare invalid byte", []byte("abc\xFF"), false},
		{"overlong start C0", []byte("abc\xC0"), false},
		{"overlong start C1", []byte("abc\xC1\x80"), false},
		{"E0 below range", []byte("abc\xE0\x80"), false},
		{"E0 in range", []byte("abc\xE0\xA0"), true},
		{"ED surrogate", []byte("abc\xED\xA0"), false},
		{"ED in range", []byte("abc\xED\x9F"), true},
		{"F0 below range", []byte("abc\xF0\x80"), false},
		{"F0 in range", []byte("abc\xF0\x90"), true},
		{"F4 above range", []byte("abc\xF4\x90"), false},
		{"F4 in range", []byte("abc\xF4\x8F"), true},
		{"F5 never valid", []byte("abc\xF5"), false},
		{"corrupt in middle", []byte("ab\xFF\xE6\xB5"), false},
		{"lone continuation", []byte("abc\x80"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := utf8ValidWindow(tc.data); got != tc.want {
				t.Fatalf("utf8ValidWindow(% x) = %v, want %v", tc.data, got, tc.want)
			}
		})
	}
}
