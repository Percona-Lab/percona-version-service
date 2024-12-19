# percona-version-service

Percona version service provides an API to share a support matrix for PMM
products.

For each product, there are multiple files in the `sources` directory. Each file
contains a list of supported components for a specific version of product.  
For example `pmm.2.29.0.pmm-server.json` contains the list of supported
components for PMM Server 2.29.0.


## How to add new product version into the list of supported versions

In this example, we'll use PMM Server which uses PSMDB operator.

### Add new version of PMM
To add a new version of PMM to the version service you can copy the
body of a file for latest version of PMM with a new name.  
For example, if we want to add new version of PMM 2.30.0:
1. Create create a new file `pmm.2.30.0.pmm-server.json`. 
2. Copy the body of file  `pmm.2.29.0.pmm-server.json` to the newly created
   file.
3. Update the version in `operator` field

### Add a new version of PSMDB operator
To add a new version of PSMDB operator to PMM, you should update the existing
`pmm.*.pmm-server.json` file.  
For example, if you want to add PSMDB Operator version `1.12.0` to PMM 2.28.0:
1. Add a new child to the `psmdbOperator` field
2. This child should have the following format:
```json lines
{
    "image_path": "[docker image name]",
    "image_hash": "[docker image hash]",
    "status": "recommended", // or "available"
    "critical": false // or true
}
```
3. The values for `image_path` and `image_hash` can be obtained from docker hub
   or by running `docker inspect`
4. The `status` can be set to `recommended` or `available`


## How to add new metadata for a product
Add a file to `sources/metadata/{product_name}/{any-identifier}.yaml`.  
The file supports the following format:
```
version: 0.6.0
recommended:
  cli: 0.4.0
  k8s: 1.27
supported:
  pg: ^2.1.0
  pxc: 0.0.0 || 1.1.1
```

`recommended` and `supported` are `map[string]string`.  
`recommended` field holds a specific version. `supported` hold a semver constraint.

Making a request to `/metadata/v1/{product}` will return all stored metadata for the given product.

## How to create a new docker image
`make docker-push` will create and push a docker image with your changes.  
If you don't want to push your docker image to DockerHub just run `make
docker-build`.  

By default, the image name is
`perconalab/version-service:$(GIT_BRANCH)-$(GIT_COMMIT)` but it can be
overridden by setting the `IMG` environment variable.

## How to publish your changes
We use Continuous Integration and Continuous Deployment paradigms to have the latest changes on server.  
For that we have 2 main branches for 2 different environments.
* Development environment is automatically deployed based on sources in `main` branch.  
* Production environment is automatically deployed based on sources in `production` branch.

They are independent of each other. Which means that to get your changes on both environments you have to create 2 different branches.
Each of them should be based on the branch related to the environment you want to deploy and PRs should be created to the same branch. 

e.g.: To publish your changes to the `development` environment, you should create new branch from `main` branch and then create a PR to get your changes merged to the `main` branch.  
Once your PR is merged, our CI will automatically publish these changes to corresponding environment.

## How to update gRPC API

Add you're new fields to appropriate messages in the `api/version.proto` file.

In order to compile the proto file and generate all the necessary code, first run `make init` to ensure all the tooling in installed. After that run `make gen` in order to generate everything.
