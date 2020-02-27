package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sets a channel ID as the autopost daily stats target channel
func setDailyStatsCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	dailyStats := db.GetGuildAutopost(m.GuildID, "dailystats")

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current dailystats channel
	if len(commandStrings) == 1 {
		if dailyStats == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Stats channel is currently not set. Please use `%sdailystats [channel]`", guildSettings.GetPrefix()))
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Stats channel is: `%s - %s` \n\nTo change it please use `%sdailystats [channel]`\nTo disable it please use `%sdailystats disable`", dailyStats.GetName(), dailyStats.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdailystats [channel]`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		dailyStats = nil
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		dailyStats = entities.NewCha(channelName, channelID)
	}

	// Write
	db.SetGuildAutopost(m.GuildID, "dailystats", dailyStats)

	if dailyStats == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Stats has been disabled!")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Stats channel is: `%s - %s`", dailyStats.GetName(), dailyStats.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Sets a channel ID as the autopost anime schedule target channel
func setDailyScheduleCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	dailySchedule := db.GetGuildAutopost(m.GuildID, "dailyschedule")

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current dailyschedule channel
	if len(commandStrings) == 1 {
		if dailySchedule == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost Daily Anime Schedule channel is currently not set. Please use `%sdailyschedule [channel]`", guildSettings.GetPrefix()))
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost Daily Anime Schedule channel is: `%s - %s` \n\nTo change it please use `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", dailySchedule.GetName(), dailySchedule.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sdailyschedule [channel]`\nTo disable it please use `%sdailyschedule disable`", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if dailySchedule == nil {
		dailySchedule = entities.NewCha("", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		dailySchedule = nil
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		dailySchedule = entities.NewCha(channelName, channelID)
	}

	// Write
	db.SetGuildAutopost(m.GuildID, "dailyschedule", dailySchedule)

	if dailySchedule == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost Daily Anime Schedule has been disabled!")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost Daily Anime Schedule channel is: `%s - %s`", dailySchedule.GetName(), dailySchedule.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Sets a channel ID as the autopost new airing anime episodes target channel
func setNewEpisodesCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	newEpisodes := db.GetGuildAutopost(m.GuildID, "newepisodes")

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current new episodes channel
	if len(commandStrings) == 1 {
		if newEpisodes == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: Autopost channel for new airing anime episodes is currently not set. Please use `%snewepisodes [channel]`", guildSettings.GetPrefix()))
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current Autopost channel for new airing anime episodes is: `%s - %s` \n\n To change it please use `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", newEpisodes.GetName(), newEpisodes.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%snewepisodes [channel]`\nTo disable it please use `%snewepisodes disable`", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if newEpisodes == nil {
		newEpisodes = entities.NewCha("", "")
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		newEpisodes = nil
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		newEpisodes = entities.NewCha(channelName, channelID)
	}

	// Write
	db.SetGuildAutopost(m.GuildID, "newepisodes", newEpisodes)

	if newEpisodes == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! Autopost for new airing anime episodes has been disabled!")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	entities.Mutex.Lock()
	entities.SetupGuildSub(m.GuildID)
	entities.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New Autopost channel for new airing anime episodes is: `%s - %s`", newEpisodes.GetName(), newEpisodes.GetID()))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    setDailyStatsCommand,
		Trigger:    "dailystats",
		Aliases:    []string{"dailystat", "daystats", "daystat", "setdailystats", "setdailystat", "setdaystats", "setdaystat"},
		Desc:       "Sets the autopost channel for daily stats",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
	Add(&Command{
		Execute:    setDailyScheduleCommand,
		Trigger:    "dailyschedule",
		Aliases:    []string{"dailyschedul", "dayschedule", "dayschedul", "setdailyschedule", "setdailyschedul"},
		Desc:       "Sets the autopost channel for daily anime schedule",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
	Add(&Command{
		Execute:    setNewEpisodesCommand,
		Trigger:    "newepisodes",
		Aliases:    []string{"newepisode", "newepisod", "episodes", "episode"},
		Desc:       "Sets the autopost channel for new airing anime episodes",
		Permission: functionality.Mod,
		Module:     "autopost",
	})
}
