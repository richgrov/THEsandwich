package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/richgrov/starport/v2/bots"
)

type config struct {
	Token           string `json:"token"`
	AirlockCategory string `json:"airlockCategory"`
	AirlockIngress  string `json:"airlockIngress"`
	LogChannel      string `json:"logs"`
}

type command struct {
	phrase   string
	callback func(*discordgo.MessageCreate)
}

type flightComputer struct {
	*bots.Bot
	airlockCategory *discordgo.Channel
	airlockIngress  string
	commands        []command
}

func (fc *flightComputer) OnJoin(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	fc.Log(bots.SevTrace, "Begin processing member join: %#v", event)
	fc.RunJoinRoutine(event.User)
}

func (fc *flightComputer) RunJoinRoutine(user *discordgo.User) {
	categoryPerms := fc.airlockCategory.PermissionOverwrites

	permOverrides := append([]*discordgo.PermissionOverwrite{}, categoryPerms...)
	permOverrides = append(permOverrides, &discordgo.PermissionOverwrite{
		ID:    user.ID,
		Type:  discordgo.PermissionOverwriteTypeMember,
		Allow: discordgo.PermissionViewChannel,
	}, &discordgo.PermissionOverwrite{
		ID:    user.ID,
		Type:  discordgo.PermissionOverwriteTypeMember,
		Allow: discordgo.PermissionSendMessages,
	})

	channel, err := fc.Session.GuildChannelCreateComplex(fc.Bot.Guild, discordgo.GuildChannelCreateData{
		Name:                 "proc-" + user.ID,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: permOverrides,
		ParentID:             fc.airlockCategory.ID,
	})

	if err != nil {
		fc.Log(bots.SevErr, "error creating channel: %v", err)
		return
	}

	msg := fmt.Sprintf("User %s has entered the airlock. Processing allocated to %s:", user.Username, channel.Mention())
	if _, err := fc.Session.ChannelMessageSend(fc.airlockIngress, msg); err != nil {
		fc.Log(bots.SevErr, "error sending ingress log: %v", err)
	}
}

func (fc *flightComputer) EvaluateAirlockCommand(event *discordgo.MessageCreate) {
	if !fc.HasPermission(event.Message, discordgo.PermissionManageChannels) {
		return
	}

	if _, err := fc.Session.ChannelMessageSendReply(event.ChannelID, ":white_check_mark: Evaluate airlock", event.Reference()); err != nil {
		fc.Log(bots.SevErr, "failed to send command acknowledgement: %v", err)
	}

	if err := fc.Session.MessageReactionAdd(event.ChannelID, event.ID, "\U0001f7e1"); err != nil {
		fc.Log(bots.SevErr, "failed to add command reaction: %v", err)
	}

	fc.RunJoinRoutine(event.Author)

	if err := fc.Session.MessageReactionAdd(event.ChannelID, event.ID, "\U0001F7E2"); err != nil {
		fc.Log(bots.SevErr, "failed to add completion reaction: %v", err)
	}
}

func (fc *flightComputer) DetatchTenCommand(event *discordgo.MessageCreate) {
	if !fc.HasPermission(event.Message, discordgo.PermissionManageMessages) {
		return
	}

	history, err := fc.Session.ChannelMessages(event.ChannelID, 11, "", "", "")
	if err != nil {
		fc.Log(bots.SevErr, "failed to get channel history: %v", err)
	}

	messageIds := make([]string, len(history))
	for i, msg := range history {
		messageIds[i] = msg.ID
	}

	if err := fc.Session.ChannelMessagesBulkDelete(event.ChannelID, messageIds); err != nil {
		fc.Log(bots.SevErr, "failed to detatch messages: %v", err)
	}
}

func (fc *flightComputer) OnMessage(s *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.ID == fc.Session.State.User.ID {
		return
	}

	message := strings.ToLower(event.Message.Content)

	for _, command := range fc.commands {
		if strings.Contains(message, command.phrase) {
			command.callback(event)
			break
		}
	}
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	bot, err := bots.NewBot(config.Token, config.LogChannel)
	if err != nil {
		log.Fatalf("Error initializing bot: %v\n", err)
	}

	airlockCategory, err := bot.Session.Channel(config.AirlockCategory)
	if err != nil {
		log.Fatalf("error fetching airlock category: %v", err)
	}

	fc := flightComputer{
		Bot:             bot,
		airlockCategory: airlockCategory,
		airlockIngress:  config.AirlockIngress,
	}

	fc.commands = []command{
		{"evaluate airlock", fc.EvaluateAirlockCommand},
		{"detatch ten", fc.DetatchTenCommand},
	}

	fc.AddEventListener(fc.OnJoin)
	fc.AddEventListener(fc.OnMessage)
	defer fc.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	log.Println("Stopping")
}

func loadConfig(filename string) (*config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}
