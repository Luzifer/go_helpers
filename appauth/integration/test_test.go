// Package integration tests the appauth implementation against a real
// OIDC Server implementation
package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const dexVersion = "v2.45.1"

var (
	dexBinaryPath    string
	dexConfigPath    string
	dexProcess       *exec.Cmd
	dexProcessCancel context.CancelFunc
)

func TestMain(m *testing.M) {
	var err error

	if err = buildDexBinary(); err != nil {
		panic(fmt.Errorf("fetching DEX binary: %w", err))
	}

	if err = startDexProcess(); err != nil {
		panic(fmt.Errorf("starting DEX process: %w", err))
	}

	m.Run()

	if err = terminateDexProcess(); err != nil {
		panic(fmt.Errorf("terminating DEX process: %w", err))
	}
}

func buildDexBinary() error {
	integrationDir, err := getIntegrationDir()
	if err != nil {
		return err
	}

	dexBinaryPath = filepath.Join(integrationDir, fmt.Sprintf("dex_%s", dexVersion))
	dexConfigPath = filepath.Join(integrationDir, "dex-config.yaml")

	if _, err = os.Stat(dexBinaryPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking dex binary: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "appauth-dex-*")
	if err != nil {
		return fmt.Errorf("creating tempdir: %w", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // best effort cleanup for test setup

	tmpBin := filepath.Join(tmpDir, "bin")
	if err = os.MkdirAll(tmpBin, 0o750); err != nil {
		return fmt.Errorf("creating temp bindir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	srcDir := filepath.Join(tmpDir, "src", "dex")
	if err = run(ctx, tmpDir,
		"git",
		"clone",
		"--branch", dexVersion,
		"--depth", "1",
		"https://github.com/dexidp/dex.git",
		srcDir,
	); err != nil {
		return fmt.Errorf("fetching dex %s source: %w", dexVersion, err)
	}

	if err = run(ctx, srcDir,
		"go",
		"build",
		"-mod=readonly",
		"-modcacherw",
		"-ldflags", fmt.Sprintf("-s -w -X main.version=%s", dexVersion),
		"-o", dexBinaryPath,
		"./cmd/dex",
	); err != nil {
		return fmt.Errorf("building dex %s: %w", dexVersion, err)
	}

	return nil
}

func startDexProcess() error {
	ctx, cancel := context.WithCancel(context.Background())
	dexProcessCancel = cancel

	//#nosec:G204 // integration test intentionally starts the pinned Dex binary it built.
	dexProcess = exec.CommandContext(ctx, dexBinaryPath, "serve", dexConfigPath)
	dexProcess.Stdout = os.Stdout
	dexProcess.Stderr = os.Stderr

	if err := dexProcess.Start(); err != nil {
		return fmt.Errorf("starting dex: %w", err)
	}

	if err := waitForDex(); err != nil {
		_ = terminateDexProcess()
		return err
	}

	return nil
}

func run(ctx context.Context, dir string, name string, args ...string) error {
	//#nosec:G204 // integration test intentionally runs fixed tool commands with controlled arguments.
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running %s: %w", name, err)
	}

	return nil
}

func terminateDexProcess() error {
	if dexProcess == nil || dexProcess.Process == nil {
		return nil
	}
	if dexProcessCancel != nil {
		dexProcessCancel()
	}

	err := dexProcess.Wait()
	if err == nil {
		return nil
	}

	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}

	return fmt.Errorf("waiting for dex: %w", err)
}

func getIntegrationDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolving integration dir")
	}

	return filepath.Dir(filename), nil
}

func waitForDex() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	client := &http.Client{Timeout: time.Second}
	url := "http://127.0.0.1:64556/dex/.well-known/openid-configuration"

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("creating dex readiness request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for dex readiness: %w", ctx.Err())

		case <-tick.C:
		}
	}
}
