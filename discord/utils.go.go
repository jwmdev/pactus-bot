package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	GREEN  = 0x008000
	RED    = 0xFF0000
	PACTUS = 0x052D5A
)

func checkMessage(i *discordgo.InteractionCreate, s *discordgo.Session, guildID, userID string) bool {
	if i.GuildID != guildID || s.State.User.ID == userID {
		return false
	}
	return true
}

func newStatus(name string, value interface{}) discordgo.UpdateStatusData {
	return discordgo.UpdateStatusData{
		Status: fmt.Sprintf("%s: %v", name, value),
	}
}
