package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type tagResp struct {
	Count        int             `json:"count"`
	NextPage     string          `json:"next"`
	PreviousPage string          `json:"previous"`
	Results      []tagRespResult `json:"results"`
}

type tagRespResult struct {
	ContentType string  `json:"content_type,omitempty"`
	Creator     float64 `json:"creator,omitempty"`
	Digest      string  `json:"digest,omitempty"`
	FullSize    float64 `json:"full_size,omitempty"`
	ID          float64 `json:"id,omitempty"`
	Images      []struct {
		Architecture string  `json:"architecture,omitempty"`
		Digest       string  `json:"digest,omitempty"`
		Features     string  `json:"features,omitempty"`
		LastPulled   string  `json:"last_pulled,omitempty"`
		LastPushed   string  `json:"last_pushed,omitempty"`
		Os           string  `json:"os,omitempty"`
		OsFeatures   string  `json:"os_features,omitempty"`
		OsVersion    string  `json:"os_version,omitempty"`
		Size         float64 `json:"size,omitempty"`
		Status       string  `json:"status,omitempty"`
		Variant      string  `json:"variant,omitempty"`
	} `json:"images,omitempty"`
	LastUpdated         string  `json:"last_updated,omitempty"`
	LastUpdater         float64 `json:"last_updater,omitempty"`
	LastUpdaterUsername string  `json:"last_updater_username,omitempty"`
	MediaType           string  `json:"media_type,omitempty"`
	Name                string  `json:"name,omitempty"`
	Repository          float64 `json:"repository,omitempty"`
	TagLastPulled       string  `json:"tag_last_pulled,omitempty"`
	TagLastPushed       string  `json:"tag_last_pushed,omitempty"`
	TagStatus           string  `json:"tag_status,omitempty"`
	V2                  bool    `json:"v2,omitempty"`
}

func (t tagRespResult) Image(imageName string) Image {
	digestAMD64 := ""
	digestARM64 := ""
	for _, v := range t.Images {
		if v.Os != "linux" {
			continue
		}
		if v.Architecture == "amd64" {
			digestAMD64 = v.Digest
		}
		if v.Architecture == "arm64" {
			digestARM64 = v.Digest
		}
	}
	return Image{
		Name:        imageName,
		Tag:         t.Name,
		DigestAMD64: strings.TrimPrefix(digestAMD64, "sha256:"),
		DigestARM64: strings.TrimPrefix(digestARM64, "sha256:"),
	}
}

type Image struct {
	Name        string
	Tag         string
	DigestAMD64 string
	DigestARM64 string
}

func (i Image) FullName() string {
	return i.Name + ":" + i.Tag
}

type RegistryClient struct {
	c     *http.Client
	cache map[string]any
}

const defaultPageSize = 100

func (r *RegistryClient) readTag(imageName string, tag string) (tagRespResult, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "registry.hub.docker.com",
		Path:   path.Join("v2", "repositories", imageName, "tags", tag),
	}

	var result tagRespResult
	cachedResult, ok := r.cache[u.String()]
	if ok {
		result, ok = cachedResult.(tagRespResult)
		if ok {
			return result, nil
		}
		panic("caching error")
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return result, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("invalid status from docker hub registry: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read body: %w", err)
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal: %w", err)
	}
	r.cache[u.String()] = result
	return result, nil
}

func (r *RegistryClient) listTags(imageName string, page int) (tagResp, error) {
	u := url.URL{
		Scheme:   "https",
		Host:     "registry.hub.docker.com",
		Path:     path.Join("v2", "repositories", imageName, "tags"),
		RawQuery: "page_size=" + strconv.Itoa(defaultPageSize) + "&page=" + strconv.Itoa(page),
	}

	var result tagResp
	cachedResult, ok := r.cache[u.String()]
	if ok {
		result, ok = cachedResult.(tagResp)
		if ok {
			return result, nil
		}
		panic("caching error")
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return result, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("invalid status from docker hub registry: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read body: %w", err)
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal: %w", err)
	}
	r.cache[u.String()] = result
	return result, nil
}

func NewClient() *RegistryClient {
	return &RegistryClient{
		c:     new(http.Client),
		cache: make(map[string]any),
	}
}

func (r *RegistryClient) GetLatestImage(imageName string) (Image, error) {
	resp, err := r.listTags(imageName, 1)
	if err != nil {
		return Image{}, fmt.Errorf("failed to get latest image: %w", err)
	}
	for _, result := range resp.Results {
		if result.Name == "latest" {
			continue
		}
		if strings.Count(result.Name, ".") == 2 {
			return result.Image(imageName), nil
		}
	}
	return Image{}, errors.New("image not found")
}

func (r *RegistryClient) GetTag(image, tag string) (Image, error) {
	resp, err := r.readTag(image, tag)
	if err != nil {
		return Image{}, fmt.Errorf("failed to read tag: %w", err)
	}
	return resp.Image(image), nil
}

func (r *RegistryClient) GetTags(imageName string) ([]string, error) {
	tags := []string{}
	for page := 1; ; page++ {
		resp, err := r.listTags(imageName, page)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", page, err)
		}
		for _, result := range resp.Results {
			tags = append(tags, result.Name)
		}
		if resp.NextPage == "" || len(resp.Results) < defaultPageSize {
			return tags, nil
		}
	}
}

func (r *RegistryClient) GetImages(imageName string, filterFunc func(tag string) bool) ([]Image, error) {
	images := []Image{}
	for page := 1; ; page++ {
		resp, err := r.listTags(imageName, page)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", page, err)
		}
		for _, result := range resp.Results {
			if !filterFunc(result.Name) {
				continue
			}

			images = append(images, result.Image(imageName))
		}
		if resp.NextPage == "" || len(resp.Results) < defaultPageSize {
			return images, nil
		}
	}
}
