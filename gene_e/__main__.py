from typing import AsyncGenerator
import anthropic
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


def list_repos():
    repo_names = [repo.name for repo in gh.get_user().get_repos()]
    return ", ".join(repo_names)


async def infer_message(prompt: str) -> AsyncGenerator[str, None]:
    response = await anth.messages.create(
        max_tokens=256,
        messages=[
            {
                "role": "user",
                "content": prompt,
            },
        ],
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

    for block in response.content:
        if block.type == "text":
            yield block.text
            continue


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
        async for response in infer_message(prompt):
            await msg.reply(response)


client.run(config["token"])
