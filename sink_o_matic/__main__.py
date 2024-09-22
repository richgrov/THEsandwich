import discord
from discord import app_commands, ui
import asyncio
import json
import base64
import anthropic
from PIL import Image
from io import BytesIO

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)

command_tree = app_commands.CommandTree(client)

anth = anthropic.AsyncAnthropic(api_key=config["anthropic"])


class SinkModal(ui.Modal, title="Add a sink"):
    def __init__(self, file: discord.Attachment) -> None:
        self.file = file
        super().__init__(title="Add a sink")

    location = ui.TextInput(
        label="Location", placeholder="Where did you find this sink?", required=True
    )

    faucet_clearance = ui.TextInput(
        label="Faucet Clearance", placeholder="3.5 inches", required=True
    )

    async def on_submit(self, interaction: discord.Interaction) -> None:
        sink_channel = interaction.client.get_channel(config["channel"])
        msg = "Sink submitted by " + interaction.user.mention + "\n" + \
                "**Found at:** " + self.location.value + "\n" + \
                "**Faucet clearance:** " + self.faucet_clearance.value + "\n" + \
                "Submit your own with `/sink`"

        await sink_channel.send(msg, file=await self.file.to_file())  #pyright: ignore
        await interaction.response.send_message("Done :white_check_mark:")


@command_tree.command(
    name="sink",
    description="Add to the sink repository",
    guild=discord.Object(config["guild"]),
)
async def list_repos(interaction: discord.Interaction, file: discord.Attachment):
    if not await has_sink(file):
        await interaction.response.send_message(
            ":x: No sink was found in the image you provided"
        )
        return

    await interaction.response.send_modal(SinkModal(file))


def limit_image_size(binary: bytes) -> bytes:
    with Image.open(BytesIO(binary)) as img:
        img.thumbnail((512, 512))
        img = img.convert("RGB")
        output = BytesIO()
        img.save(output, format="JPEG")
        return output.getvalue()


async def has_sink(image: discord.Attachment):
    binary = await image.read()
    binary = limit_image_size(binary)

    b64 = base64.b64encode(binary).decode("utf-8")

    message = await anth.messages.create(
        max_tokens=3,
        messages=[
            {
                "role": "user",
                "content": [
                    {
                        "type": "image",
                        "source": {
                            "type": "base64",
                            "media_type": "image/jpeg",
                            "data": b64,
                        },
                    }
                ],
            }
        ],
        temperature=0,
        system="Say YES if the image shows a sink, say NO otherwise. Some bathtubs may look like sinks- don't count them.",
        model="claude-3-5-sonnet-20240620",
    )

    response = message.content[0].text.lower()  # pyright: ignore
    if response.startswith("yes"):
        return True
    elif response.startswith("no"):
        return False
    else:
        print("Got unexpected response " + response)
        return False


@client.event
async def on_ready():
    await command_tree.sync(guild=discord.Object(config["guild"]))
    print("Online")


async def run_client(token):
    client = discord.Client(intents=discord.Intents.default())
    await client.start(token)

async def main():
    others = [asyncio.create_task(run_client(token)) for token in config["other"]]
    await asyncio.gather(client.start(config["token"]), *others)

# Run the main coroutine
asyncio.run(main())
