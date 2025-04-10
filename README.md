
![logo](./logo.png)
# MCP Toolbox for Databases

> [!NOTE] 
> MCP Toolbox for Databases is currently in beta, and may see breaking 
> changes until the first stable release (v1.0).

MCP Toolbox for Databases is an open source MCP server for databases It was
designed with enterprise-grade and production-quality in mind. It enables you to
develop tools easier, faster, and more securely by handling the complexities
such as connection pooling, authentication, and more.

This README provides a brief overview. For comprehensive details, see the [full
documentation](https://googleapis.github.io/genai-toolbox/).


> [!NOTE] 
> This product was originally named “Gen AI Toolbox for Databases” as
> its initial development predated MCP, but was renamed to align with recently
> added MCP compatibility. 

<!-- TOC ignore:true -->
## Table of Contents

<!-- TOC -->

- [Why Toolbox?](#why-toolbox)
- [General Architecture](#general-architecture)
- [Getting Started](#getting-started)
    - [Installing the server](#installing-the-server)
    - [Running the server](#running-the-server)
    - [Integrating your application](#integrating-your-application)
- [Configuration](#configuration)
    - [Sources](#sources)
    - [Tools](#tools)
    - [Toolsets](#toolsets)
- [Versioning](#versioning)
- [Contributing](#contributing)

<!-- /TOC -->


##  Why Toolbox?

Toolbox helps you build Gen AI tools that let your agents access data in your
database. Toolbox provides:
- **Simplified development**: Integrate tools to your agent in less than 10
  lines of code, reuse tools between multiple agents or frameworks, and deploy
  new versions of tools more easily.
- **Better performance**: Best practices such as connection pooling,
  authentication, and more.
- **Enhanced security**: Integrated auth for more secure access to your data
- **End-to-end observability**: Out of the box metrics and tracing with built-in
  support for OpenTelemetry.


## General Architecture

Toolbox sits between your application's orchestration framework and your
database, providing a control plane that is used to modify, distribute, or
invoke tools. It simplifies the management of your tools by providing you with a
centralized location to store and update tools, allowing you to share tools
between agents and applications and update those tools without necessarily
redeploying your application.

![architecture](./docs/en/getting-started/introduction/architecture.png)

## Getting Started

### Installing the server
For the latest version, check the [releases page][releases] and use the
following instructions for your OS and CPU architecture.

[releases]: https://github.com/googleapis/genai-toolbox/releases

<details open>
<summary>Binary</summary>

To install Toolbox as a binary:

<!-- {x-release-please-start-version} -->
```sh
# see releases page for other versions
export VERSION=0.4.0
curl -O https://storage.googleapis.com/genai-toolbox/v$VERSION/linux/amd64/toolbox
chmod +x toolbox
```

</details>

<details>
<summary>Container image</summary>
You can also install Toolbox as a container:

```sh
# see releases page for other versions
export VERSION=0.4.0
docker pull us-central1-docker.pkg.dev/database-toolbox/toolbox/toolbox:$VERSION
```

</details>

<details>
<summary>Compile from source</summary>

To install from source, ensure you have the latest version of
[Go installed](https://go.dev/doc/install), and then run the following command:

```sh
go install github.com/googleapis/genai-toolbox@v0.4.0
```
<!-- {x-release-please-end} -->

</details>

### Running the server

[Configure](#configuration) a `tools.yaml` to define your tools, and then
execute `toolbox` to start the server:

```sh
./toolbox --tools_file "tools.yaml"
```

You can use `toolbox help` for a full list of flags! To stop the server, send a
terminate signal (`ctrl+c` on most platforms).

For more detailed documentation on deploying to different environments, check
out the resources in the [How-to
section](https://googleapis.github.io/genai-toolbox/how-to/)

### Integrating your application

Once your server is up and running, you can load the tools into your
application. See below the list of Client SDKs for using various frameworks:

<details open>
<summary>Core</summary>

1. Install [Toolbox Core SDK][toolbox-core]:
    ```bash
    pip install toolbox-core
    ```
1. Load tools:
    ```python
    from toolbox_core import ToolboxClient

    # update the url to point to your server
    client = ToolboxClient("http://127.0.0.1:5000")

    # these tools can be passed to your application! 
    tools = await client.load_toolset("toolset_name")
    ```

For more detailed instructions on using the Toolbox Core SDK, see the
[project's README][toolbox-core-readme].

[toolbox-core]: https://pypi.org/project/toolbox-core/
[toolbox-core-readme]: https://github.com/googleapis/mcp-toolbox-sdk-python/tree/main/packages/toolbox-core/README.md

</details>
<details>
<summary>LangChain / LangGraph</summary>

1. Install [Toolbox LangChain SDK][toolbox-langchain]:
    ```bash
    pip install toolbox-langchain
    ```
1. Load tools:
    ```python
    from toolbox_langchain import ToolboxClient

    # update the url to point to your server
    client = ToolboxClient("http://127.0.0.1:5000")

    # these tools can be passed to your application! 
    tools = client.load_toolset()
    ```

For more detailed instructions on using the Toolbox LangChain SDK, see the
[project's README][toolbox-langchain-readme].

[toolbox-langchain]: https://pypi.org/project/toolbox-langchain/
[toolbox-langchain-readme]: https://github.com/googleapis/mcp-toolbox-sdk-python/blob/main/packages/toolbox-langchain/README.md

</details>

<details>
<summary>LlamaIndex</summary>

1. Install [Toolbox Llamaindex SDK][toolbox-llamaindex]:
    ```bash
    pip install toolbox-llamaindex
    ```
1. Load tools:
    ```python
    from toolbox_llamaindex import ToolboxClient

    # update the url to point to your server
    client = ToolboxClient("http://127.0.0.1:5000")

    # these tools can be passed to your application! 
    tools = client.load_toolset()
    ```

For more detailed instructions on using the Toolbox Llamaindex SDK, see the
[project's README][toolbox-llamaindex-readme].

[toolbox-llamaindex]: https://pypi.org/project/toolbox-llamaindex/
[toolbox-llamaindex-readme]: https://github.com/googleapis/genai-toolbox-llamaindex-python/blob/main/README.md

</details>

## Configuration

The primary way to configure Toolbox is through the `tools.yaml` file. If you
have multiple files, you can tell toolbox which to load with the `--tools_file
tools.yaml` flag.

You can find more detailed reference documentation to all resource types in the
[Resources](https://googleapis.github.io/genai-toolbox/resources/).
### Sources

The `sources` section of your `tools.yaml` defines what data sources your
Toolbox should have access to. Most tools will have at least one source to
execute against.

```yaml
sources:
  my-pg-source:
    kind: postgres
    host: 127.0.0.1
    port: 5432
    database: toolbox_db
    user: toolbox_user
    password: my-password
```

For more details on configuring different types of sources, see the
[Sources](https://googleapis.github.io/genai-toolbox/resources/sources).

### Tools

The `tools` section of a `tools.yaml` define the actions an agent can take: what
kind of tool it is, which source(s) it affects, what parameters it uses, etc.

```yaml
tools:
  search-hotels-by-name:
    kind: postgres-sql
    source: my-pg-source
    description: Search for hotels based on name.
    parameters:
      - name: name
        type: string
        description: The name of the hotel.
    statement: SELECT * FROM hotels WHERE name ILIKE '%' || $1 || '%';
```

For more details on configuring different types of tools, see the
[Tools](https://googleapis.github.io/genai-toolbox/resources/tools).


### Toolsets

The `toolsets` section of your `tools.yaml` allows you to define groups of tools
that you want to be able to load together. This can be useful for defining
different groups based on agent or application.

```yaml
toolsets:
    my_first_toolset:
        - my_first_tool
        - my_second_tool
    my_second_toolset:
        - my_second_tool
        - my_third_tool
```

You can load toolsets by name:

```python
# This will load all tools
all_tools = client.load_toolset()

# This will only load the tools listed in 'my_second_toolset'
my_second_toolset = client.load_toolset("my_second_toolset")
```

## Versioning

This project uses [semantic versioning](https://semver.org/), including a
`MAJOR.MINOR.PATCH` version number that increments with:

- MAJOR version when we make incompatible API changes
- MINOR version when we add functionality in a backward compatible manner
- PATCH version when we make backward compatible bug fixes

The public API that this applies to is the CLI associated with Toolbox, the
interactions with official SDKs, and the definitions in the `tools.yaml` file. 

## Contributing

Contributions are welcome. Please, see the [CONTRIBUTING](CONTRIBUTING.md)
to get started.

Please note that this project is released with a Contributor Code of Conduct.
By participating in this project you agree to abide by its terms. See
[Contributor Code of Conduct](CODE_OF_CONDUCT.md) for more information.
