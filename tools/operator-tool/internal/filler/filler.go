package filler

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
	gover "github.com/hashicorp/go-version"

	"operator-tool/internal/util"
	"operator-tool/pkg/registry"
)

// VersionFiller is a helper type for creating a map[string]*vsAPI.Version
// using information retrieved from Docker Hub.
type VersionFiller struct {
	RegistryClient      *registry.RegistryClient
	IncludeArchSuffixes bool

	errs []error
}

func (f *VersionFiller) addErr(err error) {
	f.errs = append(f.errs, err)
}

func (f *VersionFiller) exec(vm map[string]*vsAPI.Version, err error) map[string]*vsAPI.Version {
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
func (f *VersionFiller) addVersionsFromRegistry(image string, versions []string) []string {
	wantedVerisons := make(map[string]struct{}, len(versions))
	coreVersions := make(map[string]struct{})
	for _, v := range versions {
		wantedVerisons[v] = struct{}{}
		coreVersions[util.Goversion(v).Core().String()] = struct{}{}
	}

	tags, err := f.RegistryClient.GetTags(image)
	if err != nil {
		f.addErr(err)
		return nil
	}

	// getVersionMap will search for images with these suffixes. We don't need them in this function
	ignoredSuffixes := append(util.GetArchSuffixes(), "-debug")

	hasIgnoredSuffix := func(tag string) bool {
		for _, s := range ignoredSuffixes {
			if strings.HasSuffix(tag, s) {
				return true
			}
		}
		return false
	}

	for _, tag := range tags {
		if hasIgnoredSuffix(tag) {
			continue
		}
		if _, err := gover.NewVersion(tag); err != nil {
			continue
		}
		if _, ok := coreVersions[util.Goversion(tag).Core().String()]; !ok {
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
// Prerelease versions are preferred for each core version when available. See preferPrereleaseVersionsFilter function.
func (f *VersionFiller) Normal(image string, versions []string, addVersionsFromRegistry bool) map[string]*vsAPI.Version {
	if addVersionsFromRegistry {
		versions = f.addVersionsFromRegistry(image, versions)
	}

	versions = preferPrereleaseVersionsFilter(versions)

	return f.exec(getVersionMap(f.RegistryClient, image, versions, f.IncludeArchSuffixes))
}

// preferPrereleaseVersionsFilter filters a slice of version strings to prioritize prerelease versions
// for each unique core version. For example, if the input is []string{"4.0.4-40", "4.0.4"}, the output
// will be []string{"4.0.4-40"}, as the prerelease version is preferred.
//
// If no prerelease versions are found for a core version, the function returns the non-prerelease versions
// for that core version instead.
func preferPrereleaseVersionsFilter(versions []string) []string {
	verMap := make(map[string][]string)

	// Group versions by core version
	for _, v := range versions {
		coreVer := util.Goversion(v).Core().String()
		verMap[coreVer] = append(verMap[coreVer], v)
	}

	result := []string{}
	for _, versionSlice := range verMap {
		prereleaseVersions := []string{}

		// Get prerelease versions
		for _, version := range versionSlice {
			if util.Goversion(version).Prerelease() != "" {
				prereleaseVersions = append(prereleaseVersions, version)
			}
		}

		if len(prereleaseVersions) == 0 {
			result = append(result, versionSlice...)
			continue
		}
		result = append(result, prereleaseVersions...)
	}

	return result
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
func (f *VersionFiller) Regex(image string, regex string, versions []string) map[string]*vsAPI.Version {
	return f.exec(getVersionMapRegex(f.RegistryClient, image, regex, versions))
}

// Latest returns a map[string]*Version with latest version tag of the specified image.
//
// The map may include image tags with the following suffixes: "", "-amd64", "-arm64", and "-multi".
func (f *VersionFiller) Latest(image string) map[string]*vsAPI.Version {
	return f.exec(getVersionMapLatestVer(f.RegistryClient, image, f.IncludeArchSuffixes))
}

func (f *VersionFiller) Error() error {
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
		for v, versionMap := range vm {
			m[v] = versionMap
		}
	}
	return m, nil
}

func getVersionMap(rc *registry.RegistryClient, image string, versions []string, includeArchSuffixes bool) (map[string]*vsAPI.Version, error) {
	m := make(map[string]*vsAPI.Version)
	for _, v := range versions {
		images, err := rc.GetImages(image, func(tag string) bool {
			allowedSuffixes := []string{""}
			if includeArchSuffixes {
				allowedSuffixes = append(allowedSuffixes, util.GetArchSuffixes()...)
			}
			for _, s := range allowedSuffixes {
				tagWithoutSuffix := tag
				if s != "" {
					var found bool
					tagWithoutSuffix, found = strings.CutSuffix(tag, s)
					if !found {
						continue
					}
				}
				if tagWithoutSuffix == v {
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
		for v, versionMap := range vm {
			m[v] = versionMap
		}
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("image %s with %v tags was not found", image, versions)
	}
	return m, nil
}

func getVersionMapLatestVer(rc *registry.RegistryClient, imageName string, includeArchSuffixes bool) (map[string]*vsAPI.Version, error) {
	image, err := rc.GetLatestImage(imageName, includeArchSuffixes)
	if err != nil {
		return nil, err
	}
	vm, err := versionMapFromImages(image.Tag, []registry.Image{image})
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// versionMapFromImages returns a Version map for a given list of images and a base tag without any suffixes.
//
// Some images on Docker Hub are tagged like <name>, <name>-arm64, <name>-aarch64, <name>-amd64, and <name>-multi.
// This function adds images with amd64 and arm64 builds to the provided map.
//
// Specify includedArchSuffixes to include images with suffixes in resulting map.
// By default includedArchSuffixes includes only <name>-multi and <name>-amd64 images.
//
// Logic:
//   - If an image supports both amd64 and arm64 architectures and has a "-multi" suffix in its tag,
//     the function includes a version of the image tag without the "-multi" suffix in the map.
//   - If no image with both amd64 and arm64 builds is found, separate images for amd64 and arm64
//     are added individually. (if both amd64 and arm64 image suffixes are specified in the includedArchSuffixes)
func versionMapFromImages(baseTag string, images []registry.Image, includedArchSuffixes ...string) (map[string]*vsAPI.Version, error) {
	if len(includedArchSuffixes) == 0 {
		includedArchSuffixes = []string{"-multi", "-amd64"}
	}

	baseTag = trimArchSuffix(baseTag)

	slices.SortFunc(images, func(a, b registry.Image) int {
		return util.Goversion(b.Tag).Compare(util.Goversion(a.Tag))
	})

	var multiImage, amd64Image, arm64Image *registry.Image
	for _, image := range images {
		switch {
		case image.DigestARM64 == "" && image.DigestAMD64 == "":
		case image.DigestARM64 != "" && image.DigestAMD64 != "":
			if image.Tag == baseTag || multiImage == nil {
				multiImage = &image
			}
		case image.DigestARM64 != "":
			if image.Tag == baseTag || arm64Image == nil {
				arm64Image = &image
			}
		case image.DigestAMD64 != "":
			if image.Tag == baseTag || amd64Image == nil {
				amd64Image = &image
			}
		}
	}

	if multiImage == nil && amd64Image == nil && arm64Image == nil {
		return nil, fmt.Errorf("necessary tags for %s image were not found", images[0].Name)
	}

	versions := make(map[string]*vsAPI.Version)
	if multiImage != nil && slices.Contains(includedArchSuffixes, "-multi") {
		versions[baseTag+getArchSuffix(multiImage.Tag)] = &vsAPI.Version{
			ImagePath:      multiImage.FullName(),
			ImageHash:      multiImage.DigestAMD64,
			ImageHashArm64: multiImage.DigestARM64,
		}
		if multiImage.Tag == baseTag {
			return versions, nil
		}
	}
	if amd64Image != nil && slices.Contains(includedArchSuffixes, "-amd64") {
		versions[baseTag+getArchSuffix(amd64Image.Tag)] = &vsAPI.Version{
			ImagePath: amd64Image.FullName(),
			ImageHash: amd64Image.DigestAMD64,
		}
	}
	// Include arm64 if multi image is not specified
	if multiImage == nil && arm64Image != nil && slices.Contains(includedArchSuffixes, "-arm64") && slices.Contains(includedArchSuffixes, "-aarch64") {
		versions[baseTag+getArchSuffix(arm64Image.Tag)] = &vsAPI.Version{
			ImagePath:      arm64Image.FullName(),
			ImageHashArm64: arm64Image.DigestARM64,
		}
	}

	return versions, nil
}

func trimArchSuffix(tag string) string {
	return strings.TrimSuffix(tag, getArchSuffix(tag))
}

func getArchSuffix(tag string) string {
	for _, suffix := range util.GetArchSuffixes() {
		if strings.HasSuffix(tag, suffix) {
			return suffix
		}
	}
	return ""
}
