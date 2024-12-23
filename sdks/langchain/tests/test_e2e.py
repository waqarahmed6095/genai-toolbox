import conftest as conftest
import pytest
import pytest_asyncio
import json

from toolbox_langchain_sdk.client import ToolboxClient


@pytest.mark.asyncio
@pytest.mark.usefixtures("toolbox_server")
class TestE2EClient:
    @pytest_asyncio.fixture(scope="function")
    async def toolbox(self):
        toolbox = ToolboxClient("http://localhost:5000")
        yield toolbox
        await toolbox.close()

    # @pytest.mark.asyncio
    # async def test_load_tool(self, toolbox):
    #     tool = await toolbox.load_tool("get-n-rows")
    #     response = await tool.arun({"num_rows": "2"})
    #     result = response["result"]

    #     assert "test text 1" in result
    #     assert "test text 2" in result
    #     assert "test text 3" not in result

    # @pytest.mark.asyncio
    # async def test_load_toolset_all(self, toolbox):
    #     toolset = await toolbox.load_toolset()
    #     assert len(toolset) == 2
    #     tool_names = ["get-n-rows", "get-row-by-id"]
    #     assert toolset[0].name in tool_names
    #     assert toolset[1].name in tool_names

    # @pytest.mark.asyncio
    # async def test_load_toolset_single(self, toolbox):
    #     toolset = await toolbox.load_toolset("my-toolset")
    #     assert len(toolset) == 1
    #     assert toolset[0].name == "get-row-by-id"

    #     toolset = await toolbox.load_toolset("my-toolset-2")
    #     assert len(toolset) == 2
    #     tool_names = ["get-n-rows", "get-row-by-id"]
    #     assert toolset[0].name in tool_names
    #     assert toolset[1].name in tool_names

    @pytest.mark.asyncio
    async def test_load_tool_auth(self, toolbox):
        toolbox.add_auth_header("my-test-auth", lambda: "2")
        tool = await toolbox.load_tool(
            "get-row-by-id-auth",
            # auth_headers={"my-test-auth": lambda: "2"}
        )
        # response = await tool.arun({"table_name": "test_table"})
        response = await tool.arun({})

        result = response["result"]

        assert "row1" in result
        assert "row2" in result
        assert "row3" not in result

