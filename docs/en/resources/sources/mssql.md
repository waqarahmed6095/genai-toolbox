---
title: "SQL Server"
type: docs
weight: 1
description: >
  SQL Server is a relational database management system (RDBMS).

---

## About

[SQL Server][mssql-docs] is a relational database management system (RDBMS)
developed by Microsoft that allows users to store, retrieve, and manage large
amount of data through a structured format.

[mssql-docs]: https://www.microsoft.com/en-us/sql-server

## Requirements

### Database User

This source only uses standard authentication. You will need to [create a
SQL Server user][mssql-users] to login to the database with.

[mssql-users]: https://learn.microsoft.com/en-us/sql/relational-databases/security/authentication-access/create-a-database-user?view=sql-server-ver16

## Example

```yaml
sources:
    my-mssql-source:
        kind: mssql
        host: 127.0.0.1
        port: 1433
        database: my_db
        user: ${USER_NAME}
        password: ${PASSWORD}
```

{{< notice tip >}}
Use environment variable replacement with the format ${ENV_NAME}
instead of hardcoding your secrets into the configuration file.
{{< /notice >}}

## Reference

| **field** | **type** | **required** | **description**                                                        |
|-----------|:--------:|:------------:|------------------------------------------------------------------------|
| kind      |  string  |     true     | Must be "mssql".                                                       |
| host      |  string  |     true     | IP address to connect to (e.g. "127.0.0.1").                           |
| port      |  string  |     true     | Port to connect to (e.g. "1433").                                      |
| database  |  string  |     true     | Name of the SQL Server database to connect to (e.g. "my_db").          |
| user      |  string  |     true     | Name of the SQL Server user to connect as (e.g. "my-user").            |
| password  |  string  |     true     | Password of the SQL Server user (e.g. "my-password").                  |
