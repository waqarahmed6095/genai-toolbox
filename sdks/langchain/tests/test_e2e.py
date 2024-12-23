import asyncio

import pytest
from aiohttp import ClientSession
import pytest_asyncio
from langchain_core.tools import StructuredTool

import os
from unittest import mock
from toolbox_langchain_sdk.client import ToolboxClient
from toolbox_langchain_sdk.utils import ManifestSchema, ToolSchema, ParameterSchema, _schema_to_model


def get_env_var(key: str) -> str:
    v = os.environ.get(key)
    if v is None:
        raise ValueError(f"Must set env var {key}")
    return v

@pytest.mark.asyncio
class TestE2EClient:
    @pytest_asyncio.fixture(scope="function")
    async def toolbox(self):
        toolbox = ToolboxClient("http://localhost:5000")
        yield toolbox
        await toolbox.close()

    @pytest.mark.asyncio
    async def test_load_tool(self, toolbox):
        tool = await toolbox.load_tool("get-n-rows")
        response = await tool.arun({"num_rows": "2"})
        result = response['result']

        assert "test text 1" in result
        assert "test text 2" in result
        assert "test text 3" not in result

    @pytest.mark.asyncio
    async def test_load_toolset_all(self, toolbox):
        toolset = await toolbox.load_toolset()
        assert len(toolset) == 2
        tool_names = ["get-n-rows", "get-row-by-id"]
        assert toolset[0].name in tool_names
        assert toolset[1].name in tool_names

    @pytest.mark.asyncio
    async def test_load_toolset_single(self, toolbox):
        toolset = await toolbox.load_toolset("my-toolset")
        assert len(toolset) == 1
        assert toolset[0].name == "get-row-by-id"

        toolset = await toolbox.load_toolset("my-toolset-2")
        assert len(toolset) == 2
        tool_names = ["get-n-rows", "get-row-by-id"]
        assert toolset[0].name in tool_names
        assert toolset[1].name in tool_names