import pytest
import pytest_asyncio
from aiohttp import ClientResponseError

from toolbox_langchain_sdk.client import ToolboxClient


@pytest.mark.asyncio
@pytest.mark.usefixtures("toolbox_server")
class TestE2EClient:
    @pytest_asyncio.fixture(scope="function")
    async def toolbox(self):
        toolbox = ToolboxClient("http://localhost:5000")
        yield toolbox
        await toolbox.close()

    @pytest.mark.asyncio
    async def test_load_tool(self, toolbox):
        tool = await toolbox.load_tool("get-n-rows")
        response = await tool.ainvoke({"num_rows": "2"})
        result = response["result"]

        assert "row1" in result
        assert "row2" in result
        assert "row3" not in result

    @pytest.mark.asyncio
    async def test_load_toolset_specific(self, toolbox):
        toolset = await toolbox.load_toolset("my-toolset")
        assert len(toolset) == 1
        assert toolset[0].name == "get-row-by-id"

        toolset = await toolbox.load_toolset("my-toolset-2")
        assert len(toolset) == 2
        tool_names = ["get-n-rows", "get-row-by-id"]
        assert toolset[0].name in tool_names
        assert toolset[1].name in tool_names

    @pytest.mark.asyncio
    async def test_load_toolset_all(self, toolbox):
        toolset = await toolbox.load_toolset()
        assert len(toolset) == 5
        tool_names = [
            "get-n-rows",
            "get-row-by-id",
            "get-row-by-id-auth",
            "get-row-by-email-auth",
            "get-row-by-content-auth",
        ]
        assert {tool.name for tool in toolset} == tool_names

    # If a tool requires no auth but auth tokens are passed, then they are ignored
    @pytest.mark.asyncio
    async def test_run_tool_unauth_with_auth(self, toolbox, auth_token2):
        tool = await toolbox.load_tool(
            "get-row-by-id", auth_tokens={"my-test-auth": lambda: auth_token2}
        )
        response = await tool.arun({"id": "2"})
        assert "row2" in response["result"]

    @pytest.mark.asyncio
    async def test_run_tool_no_auth(self, toolbox):
        tool = await toolbox.load_tool(
            "get-row-by-id-auth",
        )
        with pytest.raises(ClientResponseError, match="401, message='Unauthorized'"):
            await tool.arun({"id": "2"})

    @pytest.mark.asyncio
    @pytest.mark.skip(reason="b/388259742")
    async def test_run_tool_wrong_auth(self, toolbox, auth_token2):
        toolbox.add_auth_token("my-test-auth", lambda: auth_token2)
        tool = await toolbox.load_tool(
            "get-row-by-id-auth",
        )
        with pytest.raises(ClientResponseError, match="401, message='Unauthorized'"):
            await tool.arun({"id": "2"})

    @pytest.mark.asyncio
    async def test_run_tool_auth(self, toolbox, auth_token1):
        toolbox.add_auth_token("my-test-auth", lambda: auth_token1)
        tool = await toolbox.load_tool(
            "get-row-by-id-auth",
        )
        response = await tool.arun({"id": "2"})
        assert "row2" in response["result"]

    @pytest.mark.asyncio
    async def test_run_tool_param_auth_no_auth(self, toolbox, auth_token1):
        tool = await toolbox.load_tool("get-row-by-email-auth")
        with pytest.raises(PermissionError, match="Login required"):
            await tool.arun({})

    @pytest.mark.asyncio
    async def test_run_tool_param_auth(self, toolbox, auth_token1):
        tool = await toolbox.load_tool(
            "get-row-by-email-auth", auth_tokens={"my-test-auth": lambda: auth_token1}
        )
        response = await tool.arun({})
        result = response["result"]
        assert "row4" in result
        assert "row5" in result
        assert "row6" in result

    @pytest.mark.asyncio
    async def test_run_tool_param_auth_no_field(self, toolbox, auth_token1):
        tool = await toolbox.load_tool(
            "get-row-by-content-auth", auth_tokens={"my-test-auth": lambda: auth_token1}
        )
        with pytest.raises(ClientResponseError, match="400, message='Bad Request'"):
            await tool.arun({})
