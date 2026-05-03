// Package github provides helpers for self-updating binaries from GitHub releases.
package github

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"text/template"
	"time"

	update "github.com/inconshreveable/go-update"
)

const (
	defaultTimeout      = 60 * time.Second
	defaultNamingScheme = `{{.ProductName}}_{{.GOOS}}_{{.GOARCH}}{{.EXT}}`
)

// Updater is the core struct of the update library holding all configurations
type (
	Updater struct {
		repo      string
		myVersion string

		HTTPClient     *http.Client
		RequestTimeout time.Duration
		Context        context.Context //nolint:containedctx // kept for historical purposes
		Filename       string

		releaseCache string
	}
)

var errReleaseNotFound = errors.New("release not found")

// NewUpdater initializes a new Updater and tries to guess the Filename
func NewUpdater(repo, myVersion string) (*Updater, error) {
	var err error
	u := &Updater{
		repo:      repo,
		myVersion: myVersion,

		HTTPClient:     http.DefaultClient,
		RequestTimeout: defaultTimeout,
		Context:        context.Background(),
	}

	u.Filename, err = u.compileFilename()

	return u, err
}

// Apply downloads the new binary from Github, fetches the SHA256 sum
// from the SHA256SUMS file and applies the update to the currently
// running binary
func (u *Updater) Apply() (err error) {
	updateAvailable, err := u.HasUpdate(false)
	if err != nil {
		return err
	}
	if !updateAvailable {
		return nil
	}

	checksum, err := u.getSHA256()
	if err != nil {
		return err
	}

	newRelease, err := u.getFile(u.Filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := newRelease.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing release file: %w", closeErr))
		}
	}()

	if err = update.Apply(newRelease, update.Options{
		Checksum: checksum,
	}); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	return nil
}

// HasUpdate checks which tag was used in the latest version and compares
// it to the current version. If it differs the function will return
// true. No comparison is done to determine whether the found version
// is higher than the current one.
//
//revive:disable-next-line:flag-parameter // does not switch to alternative behavior, just disables cache
func (u *Updater) HasUpdate(forceRefresh bool) (bool, error) {
	if forceRefresh {
		u.releaseCache = ""
	}

	latest, err := u.getLatestRelease()
	switch err {
	case nil:
		return u.myVersion != latest, nil
	case errReleaseNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (u Updater) compileFilename() (string, error) {
	repoName := strings.Split(u.repo, "/")
	if len(repoName) != 2 {
		return "", fmt.Errorf("repository name not in format <owner>/<repository>")
	}

	tpl, err := template.New("filename").Parse(defaultNamingScheme)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, map[string]any{
		"GOOS":        runtime.GOOS,
		"GOARCH":      runtime.GOARCH,
		"EXT":         ext,
		"ProductName": repoName[1],
	}); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

func (u Updater) getFile(filename string) (file io.ReadCloser, err error) {
	release, err := u.getLatestRelease()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(u.Context, u.RequestTimeout)
	defer cancel()

	requestURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", u.repo, release, filename)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	res, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing response body: %w", closeErr))
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file not found: %q", requestURL)
	}

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, res.Body); err != nil {
		return nil, fmt.Errorf("caching data in memory: %w", err)
	}

	return io.NopCloser(buf), nil
}

func (u *Updater) getLatestRelease() (release string, err error) {
	if u.releaseCache != "" {
		return u.releaseCache, nil
	}

	result := struct {
		TagName string `json:"tag_name"`
	}{}

	ctx, cancel := context.WithTimeout(u.Context, u.RequestTimeout)
	defer cancel()

	requestURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	res, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing response body: %w", closeErr))
		}
	}()

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding result: %w", err)
	}

	if res.StatusCode != http.StatusOK || result.TagName == "" {
		return "", errReleaseNotFound
	}

	u.releaseCache = result.TagName

	return result.TagName, nil
}

func (u Updater) getSHA256() (h []byte, err error) {
	shaFile, err := u.getFile("SHA256SUMS")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := shaFile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing SHA256SUMS file: %w", closeErr))
		}
	}()

	scanner := bufio.NewScanner(shaFile)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, u.Filename) {
			continue
		}

		h, err = hex.DecodeString(line[0:64])
		if err != nil {
			return nil, fmt.Errorf("decoding hash: %w", err)
		}

		return h, nil
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning SHA256SUMS: %w", err)
	}

	return nil, fmt.Errorf("no SHA256 found for file %q", u.Filename)
}
