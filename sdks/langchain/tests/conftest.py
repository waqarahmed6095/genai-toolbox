from __future__ import annotations

import os
import subprocess
import tempfile
import time
from typing import Generator

import pytest_asyncio
from google.cloud import artifactregistry_v1, secretmanager

# Define the hostname for the Artifact Registry
ARTIFACT_REGISTRY_HOSTNAME = "us-central1-docker.pkg.dev"

# Get environment variables
def get_env_var(key: str) -> str:
    v = os.environ.get(key)
    if v is None:
        raise ValueError(f"Must set env var {key}")
    return v


@pytest_asyncio.fixture(scope="session")
def toolbox_version() -> str:
    return get_env_var("TOOLBOX_VERSION")


@pytest_asyncio.fixture(scope="session")
def project_id() -> str:
    return get_env_var("GOOGLE_CLOUD_PROJECT")

def access_secret_version(
    project_id: str, secret_id: str, version_id: str = "latest"
) -> str:
    """
    Accesses the payload of a given secret version from Secret Manager.

    Args:
        project_id: The ID of the GCP project.
        secret_id: The ID of the secret.
        version_id: The ID of the secret version (defaults to "latest").

    Returns:
        The payload of the secret version as a string.
    """
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{project_id}/secrets/{secret_id}/versions/{version_id}"
    response = client.access_secret_version(request={"name": name})
    return response.payload.data.decode("UTF-8")


def create_tmpfile(content: str) -> str:
    """
    Creates a temporary file with the given content.

    Args:
        content: The content to write to the temporary file.

    Returns:
        The path to the temporary file.
    """
    with tempfile.NamedTemporaryFile(delete=False, mode="w") as tmpfile:
        tmpfile.write(content)
        return tmpfile.name


@pytest_asyncio.fixture(scope="session")
def tools_file_path(project_id: str) -> Generator[str]:
    tools_manifest = access_secret_version(project_id=project_id, secret_id="tools")
    tools_file_path = create_tmpfile(tools_manifest)
    yield tools_file_path
    os.remove(tools_file_path)

def pull_image_from_gar_with_sdk(
    hostname: str,
    project_id: str,
    repository: str,
    image: str,
    toolbox_version: str = "latest",
) -> None:
    """
    Pulls a Docker image from Google Artifact Registry using the Artifact Registry SDK.

    Args:
        hostname: The hostname of the Artifact Registry.
        project_id: The ID of the GCP project.
        repository: The name of the repository.
        image: The name of the image.
        toolbox_version: The version of toolbox to pull (defaults to "latest").
    """
    try:
        image_name = f"{hostname}/{project_id}/{repository}/{image}:{toolbox_version}"
        client = artifactregistry_v1.ArtifactRegistryClient()
        repository_name = f"projects/{project_id}/locations/{hostname.split('-')[0]}/repositories/{repository}"
        image_path = f"{repository_name}/dockerImages/{image}"

        request = artifactregistry_v1.GetDockerImageRequest(name=image_path)
        client.get_docker_image(request=request)

        print(f"Successfully pulled image: {image_name}")
    except Exception as e:
        print(f"Error pulling image: {e}")


@pytest_asyncio.fixture(scope="session")
def toolbox_server(toolbox_version: str, tools_file_path: str) -> Generator:
    """
    Starts the toolbox server as a subprocess.
    """
    print("Pulling toolbox image from Google Artifact Registry...")
    pull_image_from_gar_with_sdk(
        ARTIFACT_REGISTRY_HOSTNAME, "database-toolbox", "toolbox", toolbox_version
    )
    try:
        print("Opening toolbox server process...")
        toolbox_server = subprocess.Popen(
            ["../../toolbox", "--tools_file", tools_file_path]
        )
        # Wait for server to start
        time.sleep(10)
        print("Checking if toolbox is successfully started...")
        assert not toolbox_server.poll(), "Toolbox server failed to start"
    except subprocess.CalledProcessError as e:
        print(e.stderr.decode("utf-8"))
        print(e.stdout.decode("utf-8"))
        raise RuntimeError(f"{e}\n\n{e.stderr.decode('utf-8')}") from e
    yield
    toolbox_server.terminate()
    toolbox_server.wait()
