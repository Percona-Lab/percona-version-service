# percona-version-service

Percona version service provides API to share support matrix for PMM products.

For each product we have multiple files in `sources` directory.
Each file contains supported components list for exact version of product.  
For example `pmm.2.29.0.pmm-server.json` contains supported components for PMM Server 2.29.0.


## How to add new product version into supported versions list.

In this example I'll use PMM Server which uses PSMDB operator.
PMM Server uses PSMDB Operator.

### Add new version of PMM
To be able to add new version of PMM to version service we can copy body of file for latest version of PMM with a new name.  
For example if we want to add new version of PMM 2.30.0:
1. we should create new file `pmm.2.30.0.pmm-server.json`. 
2. Copy body of `pmm.2.29.0.pmm-server.json` to a newly created file.
3. Update version in `operator` field

### Add new version of PSMDB operator
To be able to add new version of PSMDB operator to PMM you should update existed `pmm.*.pmm-server.json` file.  
For example if we want to add PSMDB Operator version `1.12.0` to PMM 2.28.0:
1. Add new child to `psmdbOperator` field
2. this child should have the following format
```json lines
{
    "image_path": "[docker image name]",
    "image_hash": "[docker image hash]",
    "status": "recommended", // or "available"
    "critical": false // or true
}
```
3. image path and image hash can be taken from docker hub or by running `docker inspect`
4. status can be `recommended` or `available`

## How to create a new docker image
`make docker-push` will create and push docker image with your changes.  
If you don't want to push docker image to docker hub just run `make docker-build`.  
Both commands support `IMG` environment variable to set docker image name.

## How to publish
To publish your changes to dev environment please create PR to merge your changes to `main` branch.
CI will automatically publish the latest state of repository to dev environment.

To publish changes to prod environment please ask responsible person to deploy your docker image manually.
