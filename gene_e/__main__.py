import anthropic
import discord
import json

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)
anth = anthropic.AsyncAnthropic(api_key=config["anthropic"])


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
        )

        if response.content[0].type != "text":
            return

        await msg.reply(response.content[0].text)


client.run(config["token"])
