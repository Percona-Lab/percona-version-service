# operator-tool

`operator-tool` is designed to generate a source file for a version service. It retrieves a list of product versions from the [Percona Downloads](https://www.percona.com/downloads) API (`https://www.percona.com/products-api.php`) and searches for the corresponding images in the [Docker Hub repository](https://hub.docker.com/u/percona). If an image is not specified in the API, the latest tag of that image will be used.

Build it using `make init`.

## Usage

### Help

```sh
$ ./bin/operator-tool --help
Usage of ./bin/operator-tool:
  -cap int
        Sets a limit on the number of versions allowed for each major version of a product
  -file string
        Specify an older source file. The operator-tool will exclude any versions that are older than those listed in this file
  -include-arch-images
        Include images with "-multi", "-arm64", "-aarch64" suffixes in the output file
  -operator string
        Operator name. Available values: [psmdb pxc ps pg]
  -patch string
        Provide a path to a patch file to add additional images. Must be used together with the --file option
  -verbose
        Show logs
  -version string
        Operator version
```

### Generating source file from zero

```sh
$ ./bin/operator-tool --operator "psmdb" --version "1.17.0" # outputs source file for psmdb-operator
...
$ ./bin/operator-tool --operator "pg" --version "2.5.0"     # outputs source file for pg-operator
...
$ ./bin/operator-tool --operator "ps" --version "0.8.0"     # outputs source file for ps-operator
...
$ ./bin/operator-tool --operator "pxc" --version "1.15.1"   # outputs source file for pxc-operator
...
```

### Generating source file based on older file

```sh
$ ./bin/operator-tool --file ./sources/operator.2.5.0.pg-operator.json --version "1.17.0" # outputs source file for pg-operator, excluding older versions specified in the file
...
```

### Patching existing source file with a patch file

```sh
$ ./bin/operator-tool --file ./sources/operator.2.5.0.pg-operator.json --patch ./tools/operator-tool/patch-file.json.example
...
```

*Patch file example:*
The example patch file can be found [here](./patch-file.json.example).

```json
{
  "operator": {
    "2.4.28": {
      "image_path": "some-path:tag"
    }
  },
  "pmm": {
    "2.50.1": {
      "image_path": "some-path:tag"
    }
  }
}
```
