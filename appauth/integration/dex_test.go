package integration

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func completeDexLogin(client *http.Client, authURL string) (string, error) {
	nextURL := authURL

	for range 10 {
		if strings.HasPrefix(nextURL, appauthRedirectURL) {
			return nextURL, nil
		}

		body, headers, status, err := testReq(client, http.MethodGet, nextURL, nil)
		if err != nil {
			return "", err
		}

		switch status {
		case http.StatusFound, http.StatusSeeOther:
			nextURL, err = resolveURL(nextURL, headers.Get("Location"))
			if err != nil {
				return "", err
			}

			if strings.HasPrefix(nextURL, appauthRedirectURL) {
				return nextURL, nil
			}

			continue

		case http.StatusOK:

		default:
			return "", fmt.Errorf("unexpected dex response status=%d body=%s", status, body)
		}

		nextURL, err = submitDexPage(client, nextURL, body)
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("dex login did not finish")
}

func extractForm(pageURL, body string) (string, url.Values, error) {
	formMatch := regexp.MustCompile(`(?is)<form[^>]*>`).FindString(body)
	if formMatch == "" {
		return "", nil, fmt.Errorf("missing form in dex page")
	}

	formURL := pageURL
	if actionMatch := regexp.MustCompile(`(?is)\saction=["']([^"']+)["']`).FindStringSubmatch(formMatch); len(actionMatch) == 2 {
		var err error
		formURL, err = resolveURL(pageURL, html.UnescapeString(actionMatch[1]))
		if err != nil {
			return "", nil, err
		}
	}

	values := make(url.Values)
	for _, match := range regexp.MustCompile(`(?is)<input[^>]*>`).FindAllString(body, -1) {
		name := attr(match, "name")
		if name == "" {
			continue
		}

		values.Set(name, attr(match, "value"))
	}

	return formURL, values, nil
}

func resolveURL(baseRaw, location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("missing location")
	}

	baseURL, err := url.Parse(baseRaw)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	locationURL, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parsing location: %w", err)
	}

	return baseURL.ResolveReference(locationURL).String(), nil
}

func submitDexPage(client *http.Client, pageURL, body string) (string, error) {
	formURL, values, err := extractForm(pageURL, body)
	if err != nil {
		return "", err
	}

	if _, ok := values["login"]; ok {
		values.Set("login", dexUserEmail)
		values.Set("password", dexPassword)
	} else {
		values.Set("approval", "approve")
	}

	body, headers, status, err := testReqForm(client, formURL, values)
	if err != nil {
		return "", err
	}

	switch status {
	case http.StatusFound, http.StatusSeeOther:
		return resolveURL(formURL, headers.Get("Location"))

	case http.StatusOK:
		if strings.Contains(body, "<form") {
			return submitDexPage(client, formURL, body)
		}

		return "", fmt.Errorf("unexpected dex form response without form status=%d body=%s", status, body)

	default:
		return "", fmt.Errorf("unexpected dex form response status=%d body=%s", status, body)
	}
}

func attr(tag, name string) string {
	match := regexp.MustCompile(fmt.Sprintf(`(?is)\s%s=["']([^"']*)["']`, regexp.QuoteMeta(name))).FindStringSubmatch(tag)
	if len(match) != 2 {
		return ""
	}

	return html.UnescapeString(match[1])
}
