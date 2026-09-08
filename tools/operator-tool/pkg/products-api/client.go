package productsapi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	gover "github.com/hashicorp/go-version"
)

func GetProductVersions(trimPrefix string, products ...string) ([]string, error) {
	versions := []string{}
	for _, product := range products {
		productVersions, err := get(trimPrefix, product)
		if err != nil {
			return nil, fmt.Errorf("failed to get product versions for %s: %w", product, err)
		}
		versions = append(versions, productVersions...)
	}
	if len(versions) == 0 {
		return nil, errors.New("not found")
	}
	return versions, nil
}

func get(trimPrefix, product string) ([]string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "www.percona.com",
		Path:   "products-api.php",
	}
	values := make(url.Values)
	values.Add("version", product)
	resp, err := http.PostForm(u.String(), values)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status from docker registry: %s", resp.Status)
	}
	versions, err := parseXML(resp.Body, trimPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	versionMap := make(map[string]struct{})
	for _, v := range versions {
		versionMap[v] = struct{}{}
		ver := gover.Must(gover.NewVersion(v))
		versionMap[ver.Core().String()] = struct{}{}
	}
	return mapToSlice(versionMap), nil
}

func mapToSlice(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

func parseXML(data io.Reader, trimPrefix string) ([]string, error) {
	type option struct {
		Value string `xml:"value,attr"`
		Text  string `xml:",chardata"`
	}

	var options []option
	decoder := xml.NewDecoder(data)

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if startElem, ok := token.(xml.StartElement); ok && startElem.Name.Local == "option" {
			var option option
			if err := decoder.DecodeElement(&option, &startElem); err != nil {
				return nil, fmt.Errorf("failed to decode element: %w", err)
			}
			if option.Value != "" {
				options = append(options, option)
			}
		}
	}

	var versions []string
	for _, option := range options {
		if trimPrefix != "" && !strings.HasPrefix(option.Value, trimPrefix) {
			return nil, errors.New("prefix " + trimPrefix + " was not found in versions")
		}
		versions = append(versions, strings.TrimPrefix(option.Value, trimPrefix))
	}

	return versions, nil
}
