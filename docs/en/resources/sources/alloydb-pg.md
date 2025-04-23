---
title: "AlloyDB for PostgreSQL"
linkTitle: "AlloyDB"
type: docs
weight: 1
description: >
  AlloyDB for PostgreSQL is a fully-managed, PostgreSQL-compatible database for 
  demanding transactional workloads.

---

## About

[AlloyDB for PostgreSQL][alloydb-docs] is a fully-managed, PostgreSQL-compatible
database for demanding transactional workloads. It provides enterprise-grade
performance and availability while maintaining 100% compatibility with
open-source PostgreSQL.

If you are new to AlloyDB for PostgreSQL, you can [create a free trial
cluster][alloydb-free-trial].

[alloydb-docs]: https://cloud.google.com/alloydb/docs
[alloydb-free-trial]: https://cloud.google.com/alloydb/docs/create-free-trial-cluster

## Requirements

### IAM Permissions

By default, AlloyDB for PostgreSQL source uses the [AlloyDB Go
Connector][alloydb-go-conn] to authorize and establish mTLS connections to your
AlloyDB instance. The Go connector uses your [Application Default Credentials
(ADC)][adc] to authorize your connection to AlloyDB.

In addition to [setting the ADC for your server][set-adc], you need to ensure
the IAM identity has been given the following IAM roles (or corresponding
permissions):

- `roles/alloydb.client`
- `roles/serviceusage.serviceUsageConsumer`

To connect to your AlloyDB Source using IAM authentication:

1. Specify your IAM email as the `user` or leave it blank for Toolbox to fetch from ADC.
2. Leave the `password` field blank.

[alloydb-go-conn]: https://github.com/GoogleCloudPlatform/alloydb-go-connector
[adc]: https://cloud.google.com/docs/authentication#adc
[set-adc]: https://cloud.google.com/docs/authentication/provide-credentials-adc

### Networking

AlloyDB supports connecting over both from external networks via the internet
([public IP][public-ip]), and internal networks ([private IP][private-ip]).
For more information on choosing between the two options, see the AlloyDB page
[Connection overview][conn-overview].

You can configure the `ipType` parameter in your source configuration to
`public` or `private` to match your cluster's configuration. Regardless of which
you choose, all connections use IAM-based authorization and are encrypted with
mTLS.

[private-ip]: https://cloud.google.com/alloydb/docs/private-ip
[public-ip]: https://cloud.google.com/alloydb/docs/connect-public-ip
[conn-overview]: https://cloud.google.com/alloydb/docs/connection-overview

### Database User

Currently, this source only uses standard authentication. You will need to [create
a PostgreSQL user][alloydb-users] to login to the database with.

[alloydb-users]: https://cloud.google.com/alloydb/docs/database-users/about

## Example

```yaml
sources:
    my-alloydb-pg-source:
        kind: alloydb-postgres
        project: my-project-id
        region: us-central1
        cluster: my-cluster
        instance: my-instance
        database: my_db
        user: ${USER_NAME}
        password: ${PASSWORD}
        # ipType: "public"
```

{{< notice tip >}}
Use environment variable replacement with the format ${ENV_NAME}
instead of hardcoding your secrets into the configuration file.
{{< /notice >}}

## Reference

| **field** | **type** | **required** | **description**                                                                                                          |
|-----------|:--------:|:------------:|--------------------------------------------------------------------------------------------------------------------------|
| kind      |  string  |     true     | Must be "alloydb-postgres".                                                                                              |
| project   |  string  |     true     | Id of the GCP project that the cluster was created in (e.g. "my-project-id").                                            |
| region    |  string  |     true     | Name of the GCP region that the cluster was created in (e.g. "us-central1").                                             |
| cluster   |  string  |     true     | Name of the AlloyDB cluster (e.g. "my-cluster").                                                                         |
| instance  |  string  |     true     | Name of the AlloyDB instance within the cluster (e.g. "my-instance").                                                    |
| database  |  string  |     true     | Name of the Postgres database to connect to (e.g. "my_db").                                                              |
| user      |  string  |    false     | Name of the Postgres user to connect as (e.g. "my-pg-user"). Defaults to IAM auth using [ADC][adc] email if unspecified. |
| password  |  string  |    false     | Password of the Postgres user (e.g. "my-password"). Defaults to attempting IAM authentication if unspecified.            |
| ipType    |  string  |    false     | IP Type of the AlloyDB instance; must be one of `public` or `private`. Default: `public`.                                |
