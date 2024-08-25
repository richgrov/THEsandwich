import discord
from discord import app_commands, ui
import json

with open("config.json", "r") as file:
    config = json.load(file)

intents = discord.Intents.default()
intents.message_content = True

client = discord.Client(intents=intents)

command_tree = app_commands.CommandTree(client)


class SinkModal(ui.Modal, title="Add a sink"):
    def __init__(self, file: discord.Attachment) -> None:
        self.file = file
        super().__init__(title="Add a sink")

    name = ui.TextInput(
        label="Location", placeholder="Where did you find this sink?", required=True
    )

    faucet_clearance = ui.TextInput(
        label="Faucet Clearance", placeholder="3.5 inches", required=True
    )

    async def on_submit(self, interaction: discord.Interaction) -> None:
        await interaction.channel.send(  # pyright: ignore
            "Prepare for greatness", file=await self.file.to_file()
        )
        await interaction.response.send_message("Done :white_check_mark:")


@command_tree.command(
    name="sink",
    description="Add to the sink repository",
)
async def list_repos(interaction: discord.Interaction, file: discord.Attachment):
    await interaction.response.send_modal(SinkModal(file))


@client.event
async def on_ready():
    await command_tree.sync()
    print("Online")


client.run(config["token"])
