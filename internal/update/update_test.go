package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNewerVersion(t *testing.T) {
	assert.True(t, IsNewer("v1.2.0", "1.1.0"))
	assert.True(t, IsNewer("1.0.1", "v1.0.0"))
	assert.False(t, IsNewer("1.0.0", "1.0.0"))
	assert.False(t, IsNewer("1.0.0", "dev"))
	assert.False(t, IsNewer("0.9.0", "1.0.0"))
}

func TestGetLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","assets":[]}`))
	}))
	t.Cleanup(srv.Close)
	u := NewUpdaterWithBaseURL("1.0.0", srv.URL)
	rel, err := u.GetLatestRelease(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", rel.TagName)
}

func TestCheckAndUpdate_DevBuildIsNoOp(t *testing.T) {
	u := NewUpdater("dev")
	res := u.CheckAndUpdate(context.Background())
	assert.False(t, res.Updated)
	assert.NoError(t, res.Error)
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello")
	sum := sha256.Sum256(data)
	sums := hex.EncodeToString(sum[:]) + "  archive.tar.gz\n"
	assert.True(t, verifyChecksum(data, "archive.tar.gz", []byte(sums)))
	assert.False(t, verifyChecksum(data, "missing.tar.gz", []byte(sums)))
	assert.False(t, verifyChecksum([]byte("other"), "archive.tar.gz", []byte(sums)))
}

// makeTarGz builds a minimal gzip'd tar containing the binary, as GoReleaser would.
func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tar path only")
	}
	blob := makeTarGz(t, binaryName, []byte("BINARY"))
	got, err := extractBinary(blob, "remoteok_1.0.0_linux_amd64.tar.gz")
	require.NoError(t, err)
	assert.Equal(t, []byte("BINARY"), got)

	_, err = extractBinary(makeTarGz(t, "other", []byte("X")), "a.tar.gz")
	assert.Error(t, err)
}

func TestFindAssets(t *testing.T) {
	u := NewUpdater("1.0.0")
	osTok := runtime.GOOS
	archTok := runtime.GOARCH
	rel := &Release{Assets: []Asset{
		{Name: "checksums.txt", BrowserDownloadURL: "u/checksums"},
		{Name: fmt.Sprintf("remoteok_1.0.0_%s_%s.tar.gz", osTok, archTok), BrowserDownloadURL: "u/archive"},
		{Name: "remoteok_1.0.0_linux_amd64.deb", BrowserDownloadURL: "u/deb"},
	}}
	archive, checksums := u.findAssets(rel)
	require.NotNil(t, checksums)
	if osTok != "windows" { // the .tar.gz asset matches on non-windows runners
		require.NotNil(t, archive)
		assert.Equal(t, "u/archive", archive.BrowserDownloadURL)
	}
}

// TestCheckAndUpdate_FullFlow exercises download → checksum → extract → apply against a
// mock GitHub, replacing a throwaway temp "binary" (never the test binary itself).
func TestCheckAndUpdate_FullFlow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("tar/exec-path assumptions are unix-only")
	}
	archiveName := fmt.Sprintf("remoteok_1.1.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	archive := makeTarGz(t, binaryName, []byte("NEWBINARY"))
	sum := sha256.Sum256(archive)
	checksums := hex.EncodeToString(sum[:]) + "  " + archiveName + "\n"

	mux := http.NewServeMux()
	var base string
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"tag_name":"v1.1.0","assets":[
			{"name":%q,"browser_download_url":"%s/dl/archive"},
			{"name":"checksums.txt","browser_download_url":"%s/dl/sums"}]}`, archiveName, base, base)
	})
	mux.HandleFunc("/dl/archive", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/dl/sums", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksums)) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	base = srv.URL

	target := filepath.Join(t.TempDir(), "remoteok")
	require.NoError(t, os.WriteFile(target, []byte("OLD"), 0o755)) //nolint:gosec
	u := NewUpdaterWithBaseURL("1.0.0", srv.URL)
	u.ExecutablePath = target

	res := u.CheckAndUpdate(context.Background())
	require.NoError(t, res.Error)
	assert.True(t, res.Updated)
	got, err := os.ReadFile(target) //nolint:gosec
	require.NoError(t, err)
	assert.Equal(t, "NEWBINARY", string(got))
}

func TestParseVersion(t *testing.T) {
	assert.Equal(t, [3]int{1, 2, 3}, parseVersion("v1.2.3"))
	assert.Equal(t, [3]int{1, 0, 0}, parseVersion("1.0.0-rc1"))
	assert.Equal(t, [3]int{0, 0, 0}, parseVersion("garbage"))
}

func TestCheckAndUpdate_NoCompatibleArchive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0","assets":[{"name":"checksums.txt","browser_download_url":"x"}]}`))
	}))
	t.Cleanup(srv.Close)
	u := NewUpdaterWithBaseURL("1.0.0", srv.URL)
	res := u.CheckAndUpdate(context.Background())
	require.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "no compatible archive")
}

func TestCheckAndUpdate_AlreadyCurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0","assets":[]}`))
	}))
	t.Cleanup(srv.Close)
	u := NewUpdaterWithBaseURL("1.0.0", srv.URL)
	res := u.CheckAndUpdate(context.Background())
	require.NoError(t, res.Error)
	assert.False(t, res.Updated)
}

func TestGetLatestRelease_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	u := NewUpdaterWithBaseURL("1.0.0", srv.URL)
	_, err := u.GetLatestRelease(context.Background())
	require.Error(t, err)
}
