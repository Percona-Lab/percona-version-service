package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"

	"operator-tool/internal/filler"
	"operator-tool/internal/matrix"
	"operator-tool/internal/util"
	"operator-tool/pkg/registry"
)

const (
	operatorNameSuffix = "-operator"

	operatorNamePSMDB = "psmdb"
	operatorNamePXC   = "pxc"
	operatorNamePS    = "ps"
	operatorNamePG    = "pg"
)

var validOperatorNames = []string{
	operatorNamePSMDB,
	operatorNamePXC,
	operatorNamePS,
	operatorNamePG,
}

var (
	operatorName       = flag.String("operator", "", fmt.Sprintf("Operator name. Available values: %v", validOperatorNames))
	version            = flag.String("version", "", "Operator version")
	filePath           = flag.String("file", "", "Specify an older source file. The operator-tool will exclude any versions that are older than those listed in this file")
	patch              = flag.String("patch", "", "Provide a path to a patch file to add additional images. Must be used together with the --file option")
	verbose            = flag.Bool("verbose", false, "Show logs")
	includeMultiImages = flag.Bool("include-arch-images", false, `Include images with "-multi", "-arm64", "-aarch64" suffixes in the output file`)
	versionCap         = flag.Int("cap", 0, `Sets a limit on the number of versions allowed for each major version of a product`)
)

func main() {
	flag.Parse()

	if *version == "" && *patch == "" {
		log.Fatalln("ERROR: --version should be provided")
	}

	if *filePath != "" {
		product, err := util.ReadBaseFile(*filePath)
		if err != nil {
			log.Fatalln("ERROR: failed to read base file:", err.Error())
		}
		*operatorName = strings.TrimSuffix(product.Versions[0].Product, operatorNameSuffix)
	} else {
		if *operatorName == "" {
			log.Fatalln("ERROR: --operator or --file should be provided")
		}
		if *patch != "" {
			log.Fatalln("ERROR: --file should be provided when --patch is used")
		}
	}

	switch {
	case slices.Contains(validOperatorNames, *operatorName):
		if !*verbose {
			log.SetOutput(io.Discard)
		}

		if err := printSourceFile(*operatorName, *version, *filePath, *patch, *includeMultiImages, *versionCap); err != nil {
			log.SetOutput(os.Stderr)
			log.Fatalln("ERROR: failed to generate source file:", err.Error())
		}
	default:
		log.Fatalf("ERROR: Unknown operator name: %s. Available values: %v\n", *operatorName, validOperatorNames)
	}
}

func printSourceFile(operatorName, version, file, patchFile string, includeArchSuffixes bool, capacity int) error {
	var productResponse *vsAPI.ProductResponse
	var err error

	registryClient := registry.NewClient()

	if file == "" || patchFile == "" {
		productResponse, err = getProductResponse(registryClient, operatorName, version, includeArchSuffixes)
		if err != nil {
			return fmt.Errorf("failed to get product response: %w", err)
		}
		if file != "" {
			if err := deleteOldVersions(file, productResponse.Versions[0].Matrix); err != nil {
				return fmt.Errorf("failed to delete old verisons from version matrix: %w", err)
			}
		}
	} else {
		productResponse, err = patchProductResponse(registryClient, file, patchFile)
		if err != nil {
			return fmt.Errorf("failed to patch product response: %w", err)
		}
	}

	if err := updateMatrixStatuses(productResponse.Versions[0].Matrix); err != nil {
		return fmt.Errorf("failed to update matrix statuses: %w", err)
	}
	if err := limitMajorVersions(productResponse.Versions[0].Matrix, capacity); err != nil {
		return fmt.Errorf("failed to delete versions exceeding capacity: %w", err)
	}

	content, err := util.Marshal(productResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal product response: %w", err)
	}

	fmt.Println(string(content))
	return nil
}

func getProductResponse(rc *registry.RegistryClient, operatorName, version string, includeArchSuffixes bool) (*vsAPI.ProductResponse, error) {
	var versionMatrix *vsAPI.VersionMatrix
	var err error

	f := &filler.VersionFiller{
		RegistryClient:      rc,
		IncludeArchSuffixes: includeArchSuffixes,
	}
	switch operatorName {
	case operatorNamePG:
		versionMatrix, err = matrix.PG(f, version)
	case operatorNamePS:
		versionMatrix, err = matrix.PS(f, version)
	case operatorNamePSMDB:
		versionMatrix, err = matrix.PSMDB(f, version)
	case operatorNamePXC:
		versionMatrix, err = matrix.PXC(f, version)
	default:
		panic("problems with validation. unknown operator name " + operatorName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version matrix: %w", err)
	}
	return &vsAPI.ProductResponse{
		Versions: []*vsAPI.OperatorVersion{
			{
				Product:  operatorName + operatorNameSuffix,
				Operator: version,
				Matrix:   versionMatrix,
			},
		},
	}, nil
}
