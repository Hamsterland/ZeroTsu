package misc

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

var darlingTrigger int

// Periodic events such as Unbanning and RSS timer every 30 sec
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {

	err := s.UpdateStatus(0, "with her darling")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			fmt.Println(err.Error() + "\n" + ErrorLocation(err))
		}
	}

	for range time.NewTicker(30 * time.Second).C {

		// Checks whether it has to post rss thread
		RSSParser(s)

		// Goes through bannedUsers.json if it's not empty and unbans if needed
		MapMutex.Lock()
		if len(BannedUsersSlice) != 0 {
			t := time.Now()
			for index, user := range BannedUsersSlice {
				difference := t.Sub(user.UnbanDate)
				if difference > 0 {

					// Checks if user is in MemberInfo and assigns to user variable if true
					user, ok := MemberInfoMap[user.ID]
					if !ok {
						continue
					}
					// Sets unban date to now
					MemberInfoMap[user.ID].UnbanDate = t.Format("2006-01-02 15:04:05")

					// Unbans user
					err := s.GuildBanDelete(config.ServerID, user.ID)
					if err != nil {
						continue
					}

					// Removes the user ban from bannedUsers.json
					BannedUsersSlice = append(BannedUsersSlice[:index], BannedUsersSlice[index+1:]...)

					// Writes to memberInfo.json and bannedUsers.json
					MemberInfoWrite(MemberInfoMap)
					BannedUsersWrite(BannedUsersSlice)

					// Sends an embed message to bot-log
					err = UnbanEmbed(s, user, "")
				}
			}
		}
		MapMutex.Unlock()
	}
}

func UnbanEmbed(s *discordgo.Session, user *UserInfo, mod string) error {

	var (
		embedMess          discordgo.MessageEmbed
		embed    		   []*discordgo.MessageEmbedField
	)
	// Sets timestamp of unban
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Set embed color
	embedMess.Color = 0x00ff00

	if mod == "" {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned.", user.Username, user.Discrim)
	} else {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned by %v.",user.Username, user.Discrim, mod)
	}

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in bot-log
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	if err != nil {
		return err
	}
	return err
}

// Periodic 20min events
func TwentyMinTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(20 * time.Minute).C {

		// Writes emoji stats to disk
		_, err := EmojiStatsWrite(EmojiStats)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}

		// Writes user gain stats to disk
		_, err = UserChangeStatsWrite(UserStats)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}

		// Writes memberInfo to disk
		MemberInfoWrite(MemberInfoMap)

		// Fetches all guild users
		guild, err := s.Guild(config.ServerID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		// Fetches all server roles
		roles, err := s.GuildRoles(config.ServerID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		// Updates optin role stat
		t := time.Now()
		MapMutex.Lock()
		for chas := range ChannelStats {
			if ChannelStats[chas].RoleCount == nil {
				ChannelStats[chas].RoleCount = make(map[string]int)
			}
			if ChannelStats[chas].Optin {
				ChannelStats[chas].RoleCount[t.Format(DateFormat)] = GetRoleUserAmount(guild, roles, ChannelStats[chas].Name)
			}
		}
		MapMutex.Unlock()

		// Writes channel stats to disk
		_, err = ChannelStatsWrite(ChannelStats)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
}

// Pulls the rss thread and prints it
func RSSParser(s *discordgo.Session) {

	var exists bool

	if len(ReadRssThreads) == 0 {
		return
	}

	// Pulls the feed from /r/anime and puts it in feed variable
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Minute * 1}
	feed, err := fp.ParseURL("http://www.reddit.com/r/anime/new/.rss")
	if err != nil {
		return
	}
	fp.Client = &http.Client{}

	t := time.Now()
	hours := time.Hour * 16

	// Removes a thread if more than 16 hours have passed
	MapMutex.Lock()
	for p := 0; p < len(ReadRssThreadsCheck); p++ {
		// Calculates if it's time to remove
		dateRemoval := ReadRssThreadsCheck[p].Date.Add(hours)
		difference := t.Sub(dateRemoval)

		if difference > 0 {
			// Removes the fact that the thread had been posted already
			RssThreadsTimerRemove(ReadRssThreadsCheck[p].Thread, ReadRssThreadsCheck[p].Date, ReadRssThreadsCheck[p].ChannelID)
		}
	}

	// Iterates through each feed item to see if it finds something from storage
	for _, item := range feed.Items {
		itemTitleLowercase := strings.ToLower(item.Title)
		itemAuthorLowercase := strings.ToLower(item.Author.Name)
		for j := 0; j < len(ReadRssThreads); j++ {
			exists = false
			storageAuthorLowercase := strings.ToLower(ReadRssThreads[j].Author)

			if strings.Contains(itemTitleLowercase, ReadRssThreads[j].Thread) &&
				strings.Contains(itemAuthorLowercase, storageAuthorLowercase) {

				for k := 0; k < len(ReadRssThreadsCheck); k++ {
					if ReadRssThreadsCheck[k].Thread == ReadRssThreads[j].Thread &&
						ReadRssThreadsCheck[k].ChannelID == ReadRssThreads[j].Channel {
						exists = true
						break
					}
				}

				if !exists {
					// Posts latest sub episode thread and pins/unpins
					valid := RssThreadsTimerWrite(ReadRssThreads[j].Thread, t, ReadRssThreads[j].Channel)
					if valid {
						message, err := s.ChannelMessageSend(ReadRssThreads[j].Channel, item.Link)
						if err != nil {
							_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
							continue
						}
						pins, err := s.ChannelMessagesPinned(message.ChannelID)
						if err != nil {
							_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
							continue
						}
						if len(pins) != 0 {
							if strings.Contains(strings.ToLower(pins[0].Content), "episode") ||
								strings.Contains(strings.ToLower(pins[0].Content), "[spoilers]") ||
								strings.Contains(strings.ToLower(pins[0].Content), "[rewatch]") {
								err = s.ChannelMessageUnpin(pins[0].ChannelID, pins[0].ID)
								if err != nil {
									_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
									continue
								}
							}
						}
						err = s.ChannelMessagePin(message.ChannelID, message.ID)
						if err != nil {
							_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
						}
					}
				}
			}
		}
	}
	MapMutex.Unlock()
}

// Adds the voice role whenever a user joins the config voice chat
func VoiceRoleHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {

	var roleIDString string

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				return
			}
		}
	}()

	if config.VoiceChaID == "" {
		return
	}

	m, err := s.State.Member(v.GuildID, v.UserID)
	if err != nil {
		m, err = s.GuildMember(v.GuildID, v.UserID)
		if err != nil {
			return
		}
		return
	}

	// Fetches role ID
	guildRoles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
	for roleID := range guildRoles {
		if guildRoles[roleID].Name == "voice" {
			roleIDString = guildRoles[roleID].ID
		}
	}

	s.RWMutex.Lock()
	if v.ChannelID == config.VoiceChaID {
		// Adds role
		for _, role := range m.Roles {
			if role == roleIDString {
				s.RWMutex.Unlock()
				return
			}
		}
		err = s.GuildMemberRoleAdd(v.GuildID, v.UserID, roleIDString)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				s.RWMutex.Unlock()
				return
			}
			s.RWMutex.Unlock()
			return
		}
		s.RWMutex.Unlock()
		return
	}

	// Removes role
	for _, role := range m.Roles {
		if role == roleIDString {
			err := s.GuildMemberRoleRemove(v.GuildID, v.UserID, roleIDString)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
				if err != nil {
					s.RWMutex.Unlock()
					return
				}
				s.RWMutex.Unlock()
				return
			}
			break
		}
	}
	s.RWMutex.Unlock()
}

// Print fluff message on bot ping
func OnBotPing(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) && m.Author.ID == "128312718779219968" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Professor!")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) && m.Author.ID == "66207186417627136" {
		randomNum := rand.Intn(5)
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Bug hunter!")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Player!")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Big brain!")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Poster expert!")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 5 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Idiot!")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		return
	}
	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Daaarling~")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		darlingTrigger = 0
		return
	}
	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Baka!")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		darlingTrigger++
	}
}

// If there's a manual ban handle it correctly
func OnGuildBan(s *discordgo.Session, e *discordgo.GuildBanAdd) {
	s.State.RWMutex.RLock()
	for i := 0; i < len(BannedUsersSlice); i++ {
		if BannedUsersSlice[i].ID == e.User.ID {
			s.State.RWMutex.RUnlock()
			return
		}
	}
	_, err := s.ChannelMessageSend(config.BotLogID, fmt.Sprintf("%v#%v was manually permabanned. ID: %v", e.User.Username, e.User.Discriminator, e.User.ID))
	if err != nil {
		s.State.RWMutex.RUnlock()
		return
	}
	s.State.RWMutex.RUnlock()
}