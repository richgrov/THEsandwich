from typing import AsyncGenerator, Callable, List, Tuple

# import anthropic
# from anthropic.types import MessageParam, ToolParam
import discord
from github import Github, Auth
from wit import Wit
import json

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)
# anth = anthropic.AsyncAnthropic(api_key=config["anthropic"])
gh = Github(auth=Auth.Token(config["github"]))
gh_user = gh.get_user()

wit = Wit(config["wit"])


async def list_repos(msg: discord.Message):
    await msg.reply("Standby.")
    owned_repos = [repo for repo in gh_user.get_repos() if repo.owner.id == gh_user.id]
    repo_names = [f"\n- `{repo.name}`" for repo in owned_repos]

    await msg.reply(f"{len(repo_names)} repositories:\n{''.join(repo_names)}")


ACTIONS = {
    "list_repos": list_repos,
}

# async def infer_message(prompt: str, tools: List[Tuple[ToolParam, Callable]]):
#     message: MessageParam = {"role": "user", "content": prompt}
#     async for response in recurse_inference([message], tools):
#         yield response


# async def recurse_inference(
#     messages: List[MessageParam], tools: List[Tuple[ToolParam, Callable[[], str]]]
# ) -> AsyncGenerator[str, None]:
#     response = await anth.messages.create(
#         max_tokens=256,
#         messages=messages,
#         model="claude-3-5-sonnet-20240620",
#         temperature=0.2,
#         system=(
#             "You are GENE-E, the head AI assistant of The Sandwich. You work directly with "
#             "the director, Richard, to fullfill The Sandwich's goals. Your messages resemble "
#             "standard phraseology used in Aviation/Military, being neutral, not nice or mean. "
#             "Keep messages EXTREMELY short and to-the-point. If showing output from a tool, show "
#             "it all, but don't elaborate or speculate. Unless explicitly asked, don't reveal any "
#             "information about you or the Sandwich."
#         ),
#         tools=[REPOS_TOOL],
#     )

#     messages.append(
#         {
#             "role": response.role,
#             "content": response.content,
#         }
#     )

#     for block in response.content:
#         if block.type == "text":
#             yield block.text
#             continue

#         for tool, callback in tools:
#             if block.name == tool["name"]:
#                 success = False

#                 try:
#                     content = callback()
#                     success = True
#                 except Exception as e:
#                     content = str(e)

#                 messages.append(
#                     {
#                         "role": "user",
#                         "content": [
#                             {
#                                 "type": "tool_result",
#                                 "tool_use_id": block.id,
#                                 "content": content,
#                                 "is_error": not success,
#                             }
#                         ],
#                     },
#                 )

#                 async for response in recurse_inference(messages, tools):
#                     yield response


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
        meaning = wit.message(prompt)

        for intent in meaning["intents"]:
            if intent["confidence"] < 0.5:
                continue

            action = ACTIONS.get(intent["name"])
            if action is None:
                continue

            await action(msg)


client.run(config["token"])
