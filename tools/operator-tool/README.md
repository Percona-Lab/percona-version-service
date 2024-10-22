# operator-tool

`operator-tool` is designed to generate a source file for a version service. It retrieves a list of product versions from the [Percona Downloads](https://www.percona.com/downloads) API (`https://www.percona.com/products-api.php`) and searches for the corresponding images in the [Docker Hub repository](https://hub.docker.com/u/percona). If an image is not specified in the API, the latest tag of that image will be used.

Build it using `make init`.

## Usage

### Help

```
$ ./bin/operator-tool --help
Usage of ./bin/operator-tool:
  -file string
        Specify an older source file. The operator-tool will exclude any versions that are older than those listed in this file.
  -operator string
        Operator name. Available values: [psmdb-operator pxc-operator ps-operator pg-operator]
  -verbose
        Show logs
  -version string
        Operator version

```

### Generating source file from zero

```
$ ./bin/operator-tool --operator "psmdb-operator" --version "1.17.0" # outputs source file for psmdb-operator
...
$ ./bin/operator-tool --operator "pg-operator" --version "2.5.0"     # outputs source file for pg-operator
...
$ ./bin/operator-tool --operator "ps-operator" --version "0.8.0"     # outputs source file for ps-operator
...
$ ./bin/operator-tool --operator "pxc-operator" --version "1.15.1"   # outputs source file for pxc-operator
...
```

### Generating source file based on older file

```
$ ./bin/operator-tool --file ./sources/operator.2.5.0.pg-operator.json --version "1.17.0" # outputs source file for pg-operator, excluding older versions specified in the file
...
```
