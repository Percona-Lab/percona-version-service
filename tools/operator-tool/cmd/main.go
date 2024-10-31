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
	operatorName = flag.String("operator", "", fmt.Sprintf("Operator name. Available values: %v", validOperatorNames))
	version      = flag.String("version", "", "Operator version")
	filePath     = flag.String("file", "", "Specify an older source file. The operator-tool will exclude any versions that are older than those listed in this file.")
	patch        = flag.String("patch", "", "Provide a path to a patch file to add additional images. Must be used together with the --file option.")
	verbose      = flag.Bool("verbose", false, "Show logs")
)

func main() {
	flag.Parse()

	if *version == "" && *patch == "" {
		log.Fatalln("ERROR: --version should be provided")
	}

	if *filePath != "" {
		product, err := readBaseFile(*filePath)
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

		if err := printSourceFile(*operatorName, *version, *filePath, *patch); err != nil {
			log.SetOutput(os.Stderr)
			log.Fatalln("ERROR: failed to generate source file:", err.Error())
		}
	default:
		log.Fatalf("ERROR: Unknown operator name: %s. Available values: %v\n", *operatorName, validOperatorNames)
	}
}

func printSourceFile(operatorName, version, file, patchFile string) error {
	var productResponse *vsAPI.ProductResponse
	var err error

	registryClient := registry.NewClient()

	if file == "" || patchFile == "" {
		productResponse, err = getProductResponse(registryClient, operatorName, version)
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

	updateMatrixStatuses(productResponse.Versions[0].Matrix)

	content, err := marshal(productResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal product response: %w", err)
	}

	fmt.Println(string(content))
	return nil
}

func getProductResponse(rc *registry.RegistryClient, operatorName, version string) (*vsAPI.ProductResponse, error) {
	var versionMatrix *vsAPI.VersionMatrix
	var err error

	f := &VersionMapFiller{
		RegistryClient: rc,
	}
	switch operatorName {
	case operatorNamePG:
		versionMatrix, err = pgVersionMatrix(f, version)
	case operatorNamePS:
		versionMatrix, err = psVersionMatrix(f, version)
	case operatorNamePSMDB:
		versionMatrix, err = psmdbVersionMatrix(f, version)
	case operatorNamePXC:
		versionMatrix, err = pxcVersionMatrix(f, version)
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
