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

## How to create a new docker image
`make docker-push` will create and push a docker image with your changes.  
If you don't want to push your docker image to DockerHub just run `make
docker-build`.  

By default, the image name is
`perconalab/version-service:$(GIT_BRANCH)-$(GIT_COMMIT)` but it can be
overridden by setting the `IMG` environment variable.

## How to publish your changes
To publish your changes to the developent environment, please create a PR to get your changes merged to the `main` branch.

Once merged, our CI will automatically publish these changes to the dev
environment.

To publish changes to the production environment, please ask responsible person
to deploy your docker image manually.