package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	airlockCategory string
	airlockIngress  string
}

func (fc *flightComputer) OnJoin(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	fc.Log(bots.SevTrace, "Begin processing member join: %#v", event)
	fc.RunJoinRoutine(event.User)
}

func (fc *flightComputer) RunJoinRoutine(user *discordgo.User) {
	channel, err := fc.Session.GuildChannelCreateComplex(fc.Bot.Guild, discordgo.GuildChannelCreateData{
		Name: "proc-" + user.ID,
		Type: discordgo.ChannelTypeGuildText,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:    user.ID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel,
			},
			{
				ID:    user.ID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionSendMessages,
			},
		},
		ParentID: fc.airlockCategory,
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
	content := event.Message.Content
	if len(content) < 1 || content[0] != '.' {
		return
	}

	command := content[1:]

	switch command {
	case "debug join":
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

	fc := flightComputer{
		Bot:             bot,
		airlockCategory: config.AirlockCategory,
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
