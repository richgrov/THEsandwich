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

type flightComputer struct {
	*bots.Bot
	airlockCategory *discordgo.Channel
	airlockIngress  string
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

func (fc *flightComputer) OnMessage(s *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.ID == fc.Session.State.User.ID {
		return
	}

	command := strings.ToLower(event.Message.Content)

	if strings.Contains(command, "evaluate airlock") {
		fc.RunJoinRoutine(event.Author)
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
