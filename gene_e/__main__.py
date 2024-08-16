import discord
import json

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)


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
    print(prompt)


client.run(config["token"])
