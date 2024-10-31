package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"

	"operator-tool/registry"
)

// VersionMapFiller is a helper type for creating a map[string]*vsAPI.Version
// using information retrieved from Docker Hub.
type VersionMapFiller struct {
	RegistryClient *registry.RegistryClient
	errs           []error
}

func (f *VersionMapFiller) addErr(err error) {
	f.errs = append(f.errs, err)
}

func (f *VersionMapFiller) exec(vm map[string]*vsAPI.Version, err error) map[string]*vsAPI.Version {
	if err != nil {
		f.addErr(err)
		return nil
	}
	return vm
}

// addVersionsFromRegistry searches Docker Hub for all tags associated with the specified image
// and appends any missing tags that match the MAJOR.MINOR.PATCH version format to the returned versions slice.
//
// Tags with a "-debug" suffix are excluded.
func (f *VersionMapFiller) addVersionsFromRegistry(image string, versions []string) []string {
	wantedVerisons := make(map[string]struct{}, len(versions))
	coreVersions := make(map[string]struct{})
	for _, v := range versions {
		wantedVerisons[v] = struct{}{}
		coreVersions[goversion(v).Core().String()] = struct{}{}
	}

	tags, err := f.RegistryClient.GetTags(image)
	if err != nil {
		f.addErr(err)
		return nil
	}

	for _, tag := range tags {
		if strings.HasSuffix(tag, "-debug") {
			continue
		}
		if _, err := gover.NewVersion(tag); err != nil {
			continue
		}
		if _, ok := coreVersions[goversion(tag).Core().String()]; !ok {
			continue
		}
		if _, ok := wantedVerisons[tag]; ok {
			continue
		}
		versions = append(versions, tag)
	}
	return versions
}

// Normal returns a map[string]*Version for the specified image by filtering tags
// with the given list of versions.
//
// The map may include image tags with the following suffixes: "", "-amd64", "-arm64", and "-multi".
func (f *VersionMapFiller) Normal(image string, versions []string, addVersionsFromRegistry bool) map[string]*vsAPI.Version {
	if addVersionsFromRegistry {
		versions = f.addVersionsFromRegistry(image, versions)
	}
	return f.exec(getVersionMap(f.RegistryClient, image, versions))
}

// Regex returns a map[string]*Version for the specified image by filtering tags
// with the given list of versions and a regular expression.
//
// The regex argument must contain at least one matching group, which will be used
// to filter the necessary images. For example, given the regex "(^.*)(?:-logcollector)"
// and versions []string{"1.2.1"}, the tag "1.2.1-logcollector" will be included,
// while "1.3.1-logcollector", "1.2.1-some-string", and "1.2.1" will not be included.
//
// The map may include image tags with the following suffixes: "", "-amd64", "-arm64", and "-multi".
func (f *VersionMapFiller) Regex(image string, regex string, versions []string) map[string]*vsAPI.Version {
	return f.exec(getVersionMapRegex(f.RegistryClient, image, regex, versions))
}

// Latest returns a map[string]*Version with latest version tag of the specified image.
//
// The map may include image tags with the following suffixes: "", "-amd64", "-arm64", and "-multi".
func (f *VersionMapFiller) Latest(image string) map[string]*vsAPI.Version {
	return f.exec(getVersionMapLatestVer(f.RegistryClient, image))
}

func (f *VersionMapFiller) Error() error {
	return errors.Join(f.errs...)
}

func getVersionMapRegex(rc *registry.RegistryClient, image string, regex string, versions []string) (map[string]*vsAPI.Version, error) {
	m := make(map[string]*vsAPI.Version)
	r := regexp.MustCompile(regex)
	for _, v := range versions {
		images, err := rc.GetImages(image, func(tag string) bool {
			matches := r.FindStringSubmatch(tag)
			if len(matches) <= 1 {
				return false
			}
			if matches[1] != v {
				return false
			}
			return true
		})
		if err != nil {
			return nil, err
		}
		if len(images) == 0 {
			log.Printf("DEBUG: tag %s for image %s with regexp %s was not found\n", v, image, regex)
			continue
		}

		vm, err := versionMapFromImages(v, images)
		if err != nil {
			return nil, err
		}
		m[v] = vm
	}
	return m, nil
}

func getVersionMap(rc *registry.RegistryClient, image string, versions []string) (map[string]*vsAPI.Version, error) {
	m := make(map[string]*vsAPI.Version)
	for _, v := range versions {
		images, err := rc.GetImages(image, func(tag string) bool {
			allowedSuffixes := []string{"", "-amd64", "-arm64", "-multi"}
			for _, s := range allowedSuffixes {
				if tag+s == v {
					return true
				}
			}
			return false
		})
		if err != nil {
			return nil, err
		}
		if len(images) == 0 {
			log.Printf("DEBUG: tag %s for image %s was not found\n", v, image)
			continue
		}
		vm, err := versionMapFromImages(v, images)
		if err != nil {
			return nil, err
		}
		m[v] = vm
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("image %s with %v tags was not found", image, versions)
	}
	return m, nil
}

func getVersionMapLatestVer(rc *registry.RegistryClient, imageName string) (map[string]*vsAPI.Version, error) {
	image, err := rc.GetLatestImage(imageName)
	if err != nil {
		return nil, err
	}
	vm, err := versionMapFromImages(image.Tag, []registry.Image{image})
	if err != nil {
		return nil, err
	}
	return map[string]*vsAPI.Version{
		image.Tag: vm,
	}, nil
}

// versionMapFromImages returns a Version for a given list of images and a base tag without any suffixes.
//
// Some images on Docker Hub are tagged like <name>, <name>-arm64, <name>-amd64, and <name>-multi.
// This function attempts to use information from images with both amd64 and arm64 builds. If both are not available, it defaults to amd64.
//
// If multiple provided images share the same suffix, the function returns a Version with information for the latest image.
func versionMapFromImages(baseTag string, images []registry.Image) (*vsAPI.Version, error) {
	slices.SortFunc(images, func(a, b registry.Image) int {
		return goversion(b.Tag).Compare(goversion(a.Tag))
	})
	imageName := images[0].Name
	var multiImage, amd64Image, arm64Image *registry.Image
	for _, image := range images {
		if strings.HasSuffix(image.Tag, "-arm64") {
			arm64Image = &image
			continue
		}
		if multiImage == nil {
			if (image.DigestAMD64 != "" && image.DigestARM64 != "") || strings.HasSuffix(image.Tag, "-multi") {
				multiImage = &image
				continue
			}
		}
		if image.Tag == baseTag || amd64Image == nil {
			amd64Image = &image
			continue
		}
	}
	var imagePath, imageHash, imageHashArm64 string

	switch {
	case multiImage != nil:
		imagePath = multiImage.FullName()
		imageHash = multiImage.DigestAMD64
		imageHashArm64 = multiImage.DigestARM64
	case amd64Image != nil && arm64Image != nil:
		log.Printf("WARNING: Image %s has both %s and %s tags, but doesn't have \"-multi\" tag. Using %s\n", imageName, amd64Image, arm64Image, amd64Image)
		fallthrough
	case amd64Image != nil:
		imagePath = amd64Image.FullName()
		imageHash = amd64Image.DigestAMD64
	case arm64Image != nil:
		imagePath = arm64Image.FullName()
		imageHashArm64 = arm64Image.DigestARM64
	default:
		return nil, fmt.Errorf("necessary tags for %s image were not found", imageName)
	}

	return &vsAPI.Version{
		ImagePath:      imagePath,
		ImageHash:      imageHash,
		ImageHashArm64: imageHashArm64,
	}, nil
}
