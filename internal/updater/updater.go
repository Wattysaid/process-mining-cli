package updater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	BaseURL          string
	Version          string
	VerifySignatures bool
	PublicKeyPath    string
	PublicKeyURL     string
}

func Update(opts Options) error {
	osName, arch, err := detectPlatform()
	if err != nil {
		return err
	}
	asset := fmt.Sprintf("pm-assist_%s_%s.tar.gz", osName, arch)
	checksums := "checksums.txt"
	baseURL := strings.TrimSuffix(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://github.com/pm-assist/pm-assist/releases"
	}
	version := opts.Version
	if version == "" {
		version = "latest"
	}
	var downloadBase string
	if version == "latest" {
		downloadBase = baseURL + "/latest/download"
	} else {
		downloadBase = baseURL + "/download/" + version
	}

	assetPath, err := downloadFile(downloadBase+"/"+asset, asset)
	if err != nil {
		return err
	}
	checksumsPath, err := downloadFile(downloadBase+"/"+checksums, checksums)
	if err != nil {
		return err
	}
	if opts.VerifySignatures {
		if err := verifyChecksumsSignature(downloadBase, checksumsPath, opts.PublicKeyPath, opts.PublicKeyURL); err != nil {
			return err
		}
	}

	expected, err := lookupChecksum(checksumsPath, asset)
	if err != nil {
		return err
	}
	actual, err := sha256File(assetPath)
	if err != nil {
		return err
	}
	if expected != actual {
		return fmt.Errorf("checksum mismatch")
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return replaceBinary(assetPath, exePath)
}

func detectPlatform() (string, string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	if osName != "linux" && osName != "darwin" {
		return "", "", fmt.Errorf("unsupported os: %s", osName)
	}
	switch arch {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "arm64"
	default:
		return "", "", fmt.Errorf("unsupported arch: %s", arch)
	}
	return osName, arch, nil
}

func downloadFile(url string, name string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "pm-assist-update")
	if err != nil {
		return "", err
	}
	path := filepath.Join(tmpDir, name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return path, nil
}

func verifyChecksumsSignature(downloadBase string, checksumsPath string, keyPath string, keyURL string) error {
	if checksumsPath == "" {
		return errors.New("checksums path is required for signature verification")
	}
	if keyPath == "" && keyURL == "" {
		return errors.New("public key path or URL is required for signature verification")
	}
	if _, err := exec.LookPath("cosign"); err != nil {
		return errors.New("cosign not found in PATH for signature verification")
	}
	var err error
	if keyPath == "" && keyURL != "" {
		keyPath, err = downloadFile(keyURL, "checksums.txt.pub")
		if err != nil {
			return err
		}
	}
	sigPath, err := downloadFile(downloadBase+"/checksums.txt.sig", "checksums.txt.sig")
	if err != nil {
		return err
	}
	cmd := exec.Command("cosign", "verify-blob", "--key", keyPath, "--signature", sigPath, checksumsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

func lookupChecksum(path string, asset string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == asset {
			return fields[0], nil
		}
	}
	return "", errors.New("checksum not found")
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func replaceBinary(archivePath string, targetPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Name != "pm-assist" {
			continue
		}
		tmpPath := targetPath + ".tmp"
		out, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tarReader); err != nil {
			out.Close()
			return err
		}
		out.Close()
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			return err
		}
		return os.Rename(tmpPath, targetPath)
	}
	return errors.New("binary not found in archive")
}
