from __future__ import annotations

import os
import platform
import subprocess
import tempfile
import time
from typing import Generator

import pytest_asyncio
from google.cloud import secretmanager, storage


# Get environment variables
def get_env_var(key: str) -> str:
    value = os.environ.get(key)
    if value is None:
        raise ValueError(f"Must set env var {key}")
    return value


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
    """Provides a temporary file path containing the tools manifest."""
    tools_manifest = access_secret_version(
        project_id=project_id, secret_id="sdk_testing_tools"
    )
    tools_file_path = create_tmpfile(tools_manifest)
    yield tools_file_path
    os.remove(tools_file_path)


def download_blob(
    bucket_name: str, source_blob_name: str, destination_file_name: str
) -> None:
    """Downloads a blob from a gcs bucket."""

    storage_client = storage.Client()

    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(source_blob_name)
    blob.download_to_filename(destination_file_name)

    print(f"Blob {source_blob_name} downloaded to {destination_file_name}.")


def get_toolbox_binary_url(toolbox_version: str) -> str:
    """Constructs the GCS path to the toolbox binary."""
    operating_system = platform.system().lower()
    architecture = platform.machine()
    return f"v{toolbox_version}/{operating_system}/{architecture}/toolbox"


@pytest_asyncio.fixture(scope="session")
def toolbox_server(toolbox_version: str, tools_file_path: str) -> Generator:
    """
    Starts the toolbox server as a subprocess.
    """
    print("Pulling toolbox binary from gcs bucket...")
    source_blob_name = get_toolbox_binary_url(toolbox_version)
    download_blob("genai-toolbox", source_blob_name, "toolbox")

    try:
        print("Opening toolbox server process...")
        # Make toolbox executable
        os.chmod("toolbox", 0o700)
        toolbox_server = subprocess.Popen(
            ["./toolbox", "--tools_file", tools_file_path]
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

    # Clean up toolbox server
    toolbox_server.terminate()
    toolbox_server.wait()
