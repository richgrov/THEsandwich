package bots

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session    *discordgo.Session
	Guild      string
	logChannel string
}

func (bot *Bot) Log(sev severity, format string, args ...any) {
	bot.session.ChannelMessageSend(bot.logChannel, SeverityToEmote(sev)+" "+fmt.Sprintf(format, args...))
	log.Printf(format+"\n", args...)
}

func (bot *Bot) FindChannel(name string) error {
	channels, err := bot.session.GuildChannels(bot.Guild)
	if err != nil {
		return err
	}

	log.Printf("%#v", channels)
	return nil
}

func (bot *Bot) AddEventListener(callback any) {
	bot.session.AddHandler(callback)
}

func (bot *Bot) Close() {
	bot.session.Close()
}

func NewBot(token string, logChannel string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	if err := session.Open(); err != nil {
		return nil, err
	}

	if len(session.State.Guilds) != 1 {
		log.Fatalf("Activate guilds is invalid %#v\n", session.State.Guilds)
	}

	return &Bot{
		session:    session,
		Guild:      session.State.Guilds[0].ID,
		logChannel: logChannel,
	}, nil
}
