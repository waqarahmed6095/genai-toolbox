---
title: "BigQuery"
type: docs
weight: 1
description: >
  BigQuery is Google Cloud's fully managed, petabyte-scale, and cost-effective
  analytics data warehouse that lets you run analytics over vast amounts of 
  data in near real time. With BigQuery, there's no infrastructure to set 
  up or manage, letting you focus on finding meaningful insights using 
  GoogleSQL and taking advantage of flexible pricing models across on-demand 
  and flat-rate options.
---

# BigQuery Source

[BigQuery][bigquery-docs] is Google Cloud's fully managed, petabyte-scale,
and cost-effective analytics data warehouse that lets you run analytics
over vast amounts of data in near real time. With BigQuery, there's no
infrastructure to set up or manage, letting you focus on finding meaningful
insights using GoogleSQL and taking advantage of flexible pricing models
across on-demand and flat-rate options.

If you are new to BigQuery, you can try to
[load and query data with the bq tool][bigquery-quickstart-cli].

BigQuery uses [GoogleSQL][bigquery-googlesql] for querying data. GoogleSQL
is an ANSI-compliant structured query language (SQL) that is also implemented
for other Google Cloud services. SQL queries are handled by cluster nodes
in the same way as NoSQL data requests. Therefore, the same best practices
apply when creating SQL queries to run against your BigQuery data, such as
avoiding full table scans or complex filters.

[bigquery-docs]: https://cloud.google.com/bigquery/docs
[bigquery-quickstart-cli]: https://cloud.google.com/bigquery/docs/quickstarts/quickstart-command-line
[bigquery-googlesql]: https://cloud.google.com/bigquery/docs/reference/standard-sql/

## Requirements

### IAM Permissions

BigQuery uses [Identity and Access Management (IAM)][iam-overview] to control
user and group access to BigQuery resources like projects, datasets, and tables.
Toolbox will use your [Application Default Credentials (ADC)][adc] to authorize
and authenticate when interacting with [BigQuery][bigquery-docs].

In addition to [setting the ADC for your server][set-adc], you need to ensure
the IAM identity has been given the correct IAM permissions for the queries
you intend to run. Common roles include `roles/bigquery.user` (which includes
permissions to run jobs and read data) or `roles/bigquery.dataViewer`. See
[Introduction to BigQuery IAM][grant-permissions] for more information on
applying IAM permissions and roles to an identity.

[iam-overview]: https://cloud.google.com/bigquery/docs/access-control
[adc]: https://cloud.google.com/docs/authentication#adc
[set-adc]: https://cloud.google.com/docs/authentication/provide-credentials-adc
[grant-permissions]: https://cloud.google.com/bigquery/docs/access-control

## Example

```yaml
sources:
  my-bigquery-source:
    kind: "bigquery"
    project: "my-project-id"
```

## Reference

| **field** | **type** | **required** | **description**                                                               |
|-----------|:--------:|:------------:|-------------------------------------------------------------------------------|
| kind      |  string  |     true     | Must be "bigquery".                                                           |
| project   |  string  |     true     | Id of the GCP project that the cluster was created in (e.g. "my-project-id"). |
| location  |  string  |    false     | Specifies the location (e.g., 'us', 'asia-northeast1') in which to run the query job. This location must match the location of any tables referenced in the query. The default behavior is for it to be executed in the US multi-region |
