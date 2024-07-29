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

	channel, err := s.GuildChannelCreateComplex(fc.Bot.Guild, discordgo.GuildChannelCreateData{
		Name: "proc-" + event.User.ID,
		Type: discordgo.ChannelTypeGuildText,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:    event.User.ID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel,
			},
			{
				ID:    event.User.ID,
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

	msg := fmt.Sprintf("User %s has entered the airlock. Processing allocated to %s:", event.User.Username, channel.Mention())
	s.ChannelMessageSend(fc.airlockIngress, msg)
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

	fc := flightComputer{Bot: bot}
	fc.AddEventListener(fc.OnJoin)
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
