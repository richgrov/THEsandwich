from typing import AsyncGenerator, Callable, List, Tuple
import anthropic
from anthropic.types import MessageParam, ToolParam
import discord
from github import Github, Auth
import json

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)
anth = anthropic.AsyncAnthropic(api_key=config["anthropic"])
gh = Github(auth=Auth.Token(config["github"]))

REPOS_TOOL: ToolParam = {
    "name": "list_repos",
    "description": "List all of The Sandwich's repositories",
    "input_schema": {
        "type": "object",
    },
}


def list_repos():
    repo_names = [repo.name for repo in gh.get_user().get_repos()]
    return ", ".join(repo_names)


async def infer_message(prompt: str, tools: List[Tuple[ToolParam, Callable]]):
    message: MessageParam = {"role": "user", "content": prompt}
    async for response in recurse_inference([message], tools):
        yield response


async def recurse_inference(
    messages: List[MessageParam], tools: List[Tuple[ToolParam, Callable[[], str]]]
) -> AsyncGenerator[str, None]:
    response = await anth.messages.create(
        max_tokens=256,
        messages=messages,
        model="claude-3-5-sonnet-20240620",
        temperature=0.2,
        system=(
            "You are GENE-E, the head AI assistant of The Sandwich. You work directly with "
            "the director, Richard, to fullfill The Sandwich's goals. Your messages are "
            "short and consise, but unambiguous. Your tone is neutral- never overly nice or "
            "mean. Don't reveal any information about you or the Sandwich unless explicitly "
            "asked. Keep responses short and to the point."
        ),
        tools=[REPOS_TOOL],
    )

    messages.append(
        {
            "role": response.role,
            "content": response.content,
        }
    )

    for block in response.content:
        if block.type == "text":
            yield block.text
            continue

        for tool, callback in tools:
            if block.name == tool["name"]:
                success = False

                try:
                    content = callback()
                    success = True
                except Exception as e:
                    content = str(e)

                messages.append(
                    {
                        "role": "user",
                        "content": [
                            {
                                "type": "tool_result",
                                "tool_use_id": block.id,
                                "content": content,
                                "is_error": not success,
                            }
                        ],
                    },
                )

                async for response in recurse_inference(messages, tools):
                    yield response


@client.event
async def on_ready():
    print("Online")


@client.event
async def on_message(msg: discord.Message):
    assert client.user is not None

    if msg.author == client.user:
        return

    if not client.user.mentioned_in(msg):
        return

    prompt = msg.content.lstrip(f"<@{client.user.id}>").strip()

    async with msg.channel.typing():
        async for response in infer_message(prompt, [(REPOS_TOOL, list_repos)]):
            await msg.reply(response)


client.run(config["token"])
