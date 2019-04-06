### ZeroTsu is a Discord all-purpose BOT. Its functionalities are the following:

<br/>

* Channel lock for non-mods via permission change

* Show avatar for a target user. Works for people not in the server

* Extensive member system that tracks past member usernames, nicknames, mod-issued warnings, kicks, bans, reddit verification, verification date, whether in the server, timestamps for punishments, server join date

* Punishment system where you can issue warnings, kick or ban people for a set period of time with the bot and log that using the member system, or remove those punishments and unban

* Website verification system for reddit account linking and confirmation before being able to use the server

* Automated channel creation with various parameters you can give it to make them hidden (opt-in), temporary (auto deletes after a set period of time) and other

* Automated channel creation via non-mod started vote for a temp (auto deleted) channel with a minimum requirement of votes and hard cap of 3 at a time

* Tracks server emoji stats as well as message stats for general and opt-in channels. Also tracks user gain/loss per day
Regex-facilitated filter for phrases. Channel stats graph visualization by day also available in the website

* Full spoiler/opt-in/hidden channel support with reaction based role-giving or just join/leave commands. Tracks hidden channels between two dummy roles (disabled by default in favor of Kaguya BOT)

* Sort all BOT created optin/spoiler roles between the two dummy roles alphabetically

* Sort a category's channels alphabetically

* Subreddit RSS system that will post a thread if it sees it containing specific phrases and is of a set author. Set for /r/anime but can be changed in the rss.go file

* BOT say/edit commands that any mod can use to send or edit important messages with the BOT, or pretend they're a ROBOT

* Automatically give a channel to a user when they join a voice channel, and remove it when they leave using a role named "voice"

* RemindMe feature where it either messages you or pings you with a message you've set after a period of time you've set.

<br/>

How to install:
1. Download in a folder.
2. Edit config.json with your own values. Use only one for each, except for CommandRoles. Everything is required unless stated otherwise:

       BotPrefix is the character that needs to be used before every command

       BotID is the ID of the BOT you are using

       ServerID is the ID of the server the BOT is going to be managing

       BotLogID is the ID of the channel in which the bot will dump errors, timed events, punishments and other things

       CommandRoles are the admin/mod/bot role IDs

       OptInUnder is the name of the top dummy role for spoiler/opt-in/hidden channels

       OptInAbove is the name of the bottom dummy role for spoiler/opt-in/hidden channels

       VoiceChaID is the ID of the voice channel you want the bot to track and give the "voice" role to. Leave empty if not using it

       Website is the address/ip+:port of the website or server. Leave empty if not using it

       ModCategoryID is the ID of a mod category if it exists. Leave empty if it doesn't exist

       VoteChannelCategoryID is the category ID of the category in which the channel created from a user channel creation vote is put
	   
	   Waifus is whether to enable the waifu commands. Use "true" to turn it on. Otherwise leave empty or "false"

3. Make a file called configsecrets.json in the foldder config.json is in and set this up the following way. It's for verification and optional IF Website value above is empty:
```javascript
{
  "RedditName": "redditAppName",
  "RedditSecret": "RedditAppSecret",
  "DiscordSecret": "DiscordBOTSecret"
}
```
4. Set your "ZeroTsuToken" environment variable to the BOT token (either hidden on the system env or in config.go ReadConfig func with os.Setenv("ZeroTsuToken", "TOKEN"))
5. Compile in your favorite IDE or compiler with "go build" (or type "set GOOS=[Preferred OS]" to change OS first (like linux) and then "go build".)
6. Invite BOT to server and give it an admin role
7. Start the BOT and use

<br/>

If you have discovered any bugs or have questions, please message Apiks or raise an issue.

If you use the BOT successfuly, please also let Apiks know