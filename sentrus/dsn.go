package sentrus

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type parsedDSN struct {
	authHeader string
	endpoint   string
}

func parseDSN(rawDSN string) (parsedDSN, error) {
	if rawDSN == "" {
		return parsedDSN{}, errDSNRequired
	}

	dsn, err := url.Parse(rawDSN)
	if err != nil {
		return parsedDSN{}, fmt.Errorf("parsing DSN: %w", err)
	}
	if err = validateDSNURL(dsn); err != nil {
		return parsedDSN{}, err
	}

	authHeader, err := buildAuthHeader(dsn.User)
	if err != nil {
		return parsedDSN{}, err
	}

	pathPrefix, projectID, err := splitProjectPath(dsn.Path)
	if err != nil {
		return parsedDSN{}, err
	}

	dsn.User = nil
	dsn.Path = fmt.Sprintf("%s/api/%s/envelope/", pathPrefix, projectID)
	dsn.RawPath = ""
	return parsedDSN{authHeader: authHeader, endpoint: dsn.String()}, nil
}

func buildAuthHeader(user *url.Userinfo) (string, error) {
	publicKey := user.Username()
	secret, _ := user.Password()
	if !isValidCredential(publicKey) || (secret != "" && !isValidCredential(secret)) {
		return "", errors.New("DSN contains an invalid credential")
	}

	authHeader := fmt.Sprintf("Sentry sentry_version=7, sentry_key=%s", publicKey)
	if secret != "" {
		authHeader += fmt.Sprintf(", sentry_secret=%s", secret)
	}

	return authHeader, nil
}

func isValidCredential(credential string) bool {
	return credential != "" && !strings.ContainsAny(credential, " ,\r\n")
}

func splitProjectPath(rawPath string) (pathPrefix, projectID string, err error) {
	path := strings.TrimSuffix(rawPath, "/")
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash < 0 || lastSlash == len(path)-1 {
		return "", "", errors.New("DSN must contain a project ID")
	}

	projectID = path[lastSlash+1:]
	if projectID == "" || strings.Contains(projectID, "/") {
		return "", "", errors.New("DSN contains an invalid project ID")
	}

	return path[:lastSlash], projectID, nil
}

func validateDSNURL(dsn *url.URL) error {
	if dsn.Scheme != "http" && dsn.Scheme != "https" {
		return errors.New("DSN scheme must be HTTP or HTTPS")
	}
	if dsn.Host == "" || dsn.User == nil {
		return errors.New("DSN must contain a host and public key")
	}
	if dsn.RawQuery != "" || dsn.Fragment != "" {
		return errors.New("DSN must not contain a query or fragment")
	}

	return nil
}
