package main

import (
	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"

	productsapi "operator-tool/pkg/products-api"
)

func pgVersionMatrix(f *VersionMapFiller, version string) (*vsAPI.VersionMatrix, error) {
	pgVersions, err := productsapi.GetProductVersions("", "postgresql-distribution-16", "postgresql-distribution-15", "postgresql-distribution-14", "postgresql-distribution-13", "postgresql-distribution-12")
	if err != nil {
		return nil, err
	}

	matrix := &vsAPI.VersionMatrix{
		Postgresql: f.Regex("percona/percona-postgresql-operator", `(?:^\d+\.\d+\.\d+-ppg)(\d+\.\d+)(?:-postgres$)`, pgVersions),
		Pgbackrest: f.Regex("percona/percona-postgresql-operator", `(?:^\d+\.\d+\.\d+-ppg)(\d+\.\d+)(?:-pgbackrest)`, pgVersions),
		Pgbouncer:  f.Regex("percona/percona-postgresql-operator", `(?:^\d+\.\d+\.\d+-ppg)(\d+\.\d+)(?:-pgbouncer)`, pgVersions),
		Postgis:    f.Regex("percona/percona-postgresql-operator", `(?:^\d+\.\d+\.\d+-ppg)(\d+\.\d+)(?:-postgres-gis)`, pgVersions),
		Pmm:        f.Latest("percona/pmm-client"),
		Operator:   f.Normal("percona/percona-postgresql-operator", []string{version}, false),
	}
	if err := f.Error(); err != nil {
		return nil, err
	}
	return matrix, nil
}

func psVersionMatrix(f *VersionMapFiller, version string) (*vsAPI.VersionMatrix, error) {
	psVersions, err := productsapi.GetProductVersions("Percona-Server-", "Percona-Server-8.0")
	if err != nil {
		return nil, err
	}

	matrix := &vsAPI.VersionMatrix{
		Mysql:        f.Normal("percona/percona-server", psVersions, true),
		Pmm:          f.Latest("percona/pmm-client"),
		Router:       f.Normal("percona/percona-mysql-router", psVersions, true),
		Backup:       f.Normal("percona/percona-xtrabackup", psVersions, true),
		Operator:     f.Normal("percona/percona-server-mysql-operator", []string{version}, false),
		Haproxy:      f.Latest("percona/haproxy"),
		Orchestrator: f.Latest("percona/percona-orchestrator"),
		Toolkit:      f.Latest("percona/percona-toolkit"),
	}

	if err := f.Error(); err != nil {
		return nil, err
	}
	return matrix, nil
}

func psmdbVersionMatrix(f *VersionMapFiller, version string) (*vsAPI.VersionMatrix, error) {
	mongoVersions, err := productsapi.GetProductVersions("percona-server-mongodb-", "percona-server-mongodb-7.0", "percona-server-mongodb-6.0", "percona-server-mongodb-5.0")
	if err != nil {
		return nil, err
	}

	pbmVersions, err := productsapi.GetProductVersions("percona-backup-mongodb-", "percona-backup-mongodb")
	if err != nil {
		return nil, err
	}

	matrix := &vsAPI.VersionMatrix{
		Mongod:   f.Normal("percona/percona-server-mongodb", mongoVersions, true),
		Pmm:      f.Latest("percona/pmm-client"),
		Backup:   f.Normal("percona/percona-backup-mongodb", pbmVersions, true),
		Operator: f.Normal("percona/percona-server-mongodb-operator", []string{version}, false),
	}

	if err := f.Error(); err != nil {
		return nil, err
	}

	return matrix, nil
}

func pxcVersionMatrix(f *VersionMapFiller, version string) (*vsAPI.VersionMatrix, error) {
	pxcVersions, err := productsapi.GetProductVersions("Percona-XtraDB-Cluster-", "Percona-XtraDB-Cluster-80", "Percona-XtraDB-Cluster-57")
	if err != nil {
		return nil, err
	}

	xtrabackupVersions, err := productsapi.GetProductVersions("Percona-XtraBackup-", "Percona-XtraBackup-8.0", "Percona-XtraBackup-2.4")
	if err != nil {
		return nil, err
	}
	matrix := &vsAPI.VersionMatrix{
		Pxc:          f.Normal("percona/percona-xtradb-cluster", pxcVersions, true),
		Pmm:          f.Latest("percona/pmm-client"),
		Proxysql:     f.Latest("percona/proxysql"),
		Haproxy:      f.Latest("percona/haproxy"),
		Backup:       f.Regex("percona/percona-xtradb-cluster-operator", `(?:^\d+\.\d+\.\d+-pxc\d+\.\d+-backup-pxb)(.*)`, xtrabackupVersions),
		LogCollector: f.Regex("percona/percona-xtradb-cluster-operator", `(^.*)(?:-logcollector)`, []string{version}),
		Operator:     f.Normal("percona/percona-xtradb-cluster-operator", []string{version}, false),
	}

	if err := f.Error(); err != nil {
		return nil, err
	}
	return matrix, nil
}
