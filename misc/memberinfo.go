package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var (
	MemberInfoMap    = make(map[string]*UserInfo)
	BannedUsersSlice []BannedUsers
	MapMutex         sync.Mutex
	Key              = []byte("VfBhgLzmD4QH3W94pjgdbH8Tyv2HPRzq")
)

// UserInfo is the in memory storage of each user's information
type UserInfo struct {
	ID             string   			`json:"id"`
	Discrim        string   			`json:"discrim"`
	Username       string   			`json:"username"`
	Nickname       string   			`json:"nickname,omitempty"`
	PastUsernames  []string 			`json:"pastUsernames,omitempty"`
	PastNicknames  []string 			`json:"pastNicknames,omitempty"`
	Warnings       []string 			`json:"warnings,omitempty"`
	Kicks          []string 			`json:"kicks,omitempty"`
	Bans           []string 			`json:"bans,omitempty"`
	JoinDate       string   			`json:"joinDate"`
	RedditUsername string   			`json:"redditUser,omitempty"`
	VerifiedDate   string   			`json:"verifiedDate,omitempty"`
	UnbanDate      string   			`json:"unbanDate,omitempty"`
	Timestamps 	   []Punishment			`json:"timestamps,omitempty"`
	Waifu		   Waifu				`json:"waifu,omitempty"`
}

// Creates a struct type in which we'll hold every banned user
type BannedUsers struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	UnbanDate time.Time `json:"unbanDate"`
}

// Struct where we'll hold punishment timestamps
type Punishment struct {
	Punishment string 					`json:"punishment"`
	Type	   string					`json:"type"`
	Timestamp  time.Time				`json:"timestamp"`
}

// Reads member info from memberInfo.json
func MemberInfoRead() {

	// Reads all the member users from the memberInfo.json file and puts them in memberInfoByte as bytes
	memberInfoByte, err := ioutil.ReadFile("database/memberInfo.json")
	if err != nil {
		return
	}

	// Takes all the users from memberInfo.json from byte and puts them into the UserInfo map
	MapMutex.Lock()
	err = json.Unmarshal(memberInfoByte, &MemberInfoMap)
	if err != nil {
		MapMutex.Unlock()
		return
	}

	// Fixes empty IDs. Unneeded unless they show up again. If so uncomment the below
	//for ID, user := range MemberInfoMap {
	//	if user.ID == "" {
	//		user.ID = ID
	//	}
	//}
	MapMutex.Unlock()
}

// Writes member info to memberInfo.json
func MemberInfoWrite(info map[string]*UserInfo) {

	// Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	err = ioutil.WriteFile("database/memberInfo.json", MarshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Initializes user in memberInfo if he doesn't exist there
func InitializeUser(u *discordgo.Member) {

	var temp UserInfo

	// Sets ID, username and discriminator
	temp.ID = u.User.ID
	temp.Username = u.User.Username
	temp.Discrim = u.User.Discriminator

	// Stores time of joining
	t := time.Now()
	z, _ := t.Zone()
	join := t.Format("2006-01-02 15:04:05") + " " + z

	// Sets join date
	temp.JoinDate = join

	MemberInfoMap[u.User.ID] = &temp
}

// Checks if user exists in memberInfo on joining server and adds him if he doesn't
// Also updates usernames and/or nicknames
// Also updates discriminator
// Also verifies them if they're already verified in memberinfo
func OnMemberJoinGuild(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	var (
		flag        = false
		initialized = false
		nameFlag	= false
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	// Pulls info on user if possible
	user, err := s.GuildMember(config.ServerID, e.User.ID)
	if err != nil {
		return
	}

	// If memberInfo is empty, it initializes
	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {

		// Initializes the first user of memberInfo
		InitializeUser(user)

		flag = true
		initialized = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" {
			dm, _ := s.UserChannelCreate(user.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
				"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
		}

	} else {
		// Checks if user exists in memberInfo.json. If yes it changes flag to true
		if _, ok := MemberInfoMap[user.User.ID]; ok {
			flag = true
		}
	}

	// If user still doesn't exist after check above, it initializes user
	if !flag {

		// Initializes the new user
		InitializeUser(user)
		initialized = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" {
			dm, _ := s.UserChannelCreate(user.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
				"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
		}
	}

	// Writes User Initialization to disk
	MemberInfoWrite(MemberInfoMap)

	// Fetches user from memberInfo
	existingUser, ok := MemberInfoMap[user.User.ID]
	if !ok {
		MapMutex.Unlock()
		return
	}

	// If user is already in memberInfo but hasn't verified before tell him to verify now
	if MemberInfoMap[user.User.ID].RedditUsername == "" && !initialized {

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" {
			dm, _ := s.UserChannelCreate(user.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
				"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
		}
	}
	MapMutex.Unlock()

	// Checks if the user's current username is the same as the one in the database. Otherwise updates
	if user.User.Username != existingUser.Username {
		flag := true
		nameFlag = true
		lower := strings.ToLower(user.User.Username)

		for _, names := range existingUser.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			existingUser.PastUsernames = append(existingUser.PastUsernames, user.User.Username)
			existingUser.Username = user.User.Username
		}
	}

	// Checks if the user's current nickname is the same as the one in the database. Otherwise updates
	if existingUser.Nickname != user.Nick && user.Nick != "" {
		flag := true
		nameFlag = true
		lower := strings.ToLower(user.Nick)

		for _, names := range existingUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			existingUser.PastNicknames = append(existingUser.PastNicknames, user.Nick)
			existingUser.Nickname = user.Nick
		}
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.User.Discriminator != existingUser.Discrim {
		nameFlag = true
		existingUser.Discrim = user.User.Discriminator
	}

	// Saves the updates to memberInfoMap and writes to disk if need be
	if nameFlag {
		MapMutex.Lock()
		MemberInfoMap[user.User.ID] = existingUser
		MemberInfoWrite(MemberInfoMap)
		MapMutex.Unlock()
	}
}

// OnMemberUpdate listens for member updates to compare nicks
func OnMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {

	var writeFlag bool

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	s.RWMutex.RLock()
	userMember := e
	s.RWMutex.RUnlock()

	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {
		MapMutex.Unlock()
		return
	}

	// Fetches user from memberInfo if possible
	user, ok := MemberInfoMap[userMember.User.ID]
	if !ok {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()

	// Checks nicknames and updates if needed
	if user.Nickname != userMember.Nick && userMember.Nick != "" {
		flag := true
		lower := strings.ToLower(userMember.Nick)

		for _, names := range user.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastNicknames = append(user.PastNicknames, userMember.Nick)
			user.Nickname = userMember.Nick
		}

		writeFlag = true
	}

	// Checks if username or discrim were changed, else do NOT write to disk
	if !writeFlag {
		return
	}

	// Saves the updates to memberInfoMap and writes to disk
	MapMutex.Lock()
	MemberInfoMap[userMember.User.ID] = user
	MemberInfoWrite(MemberInfoMap)
	MapMutex.Unlock()
}

// OnPresenceUpdate listens for user updates to compare usernames and discrim
func OnPresenceUpdate(s *discordgo.Session, e *discordgo.PresenceUpdate) {

	var writeFlag bool

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	s.RWMutex.RLock()
	userMember := e.User
	s.RWMutex.RUnlock()

	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {
		MapMutex.Unlock()
		return
	}

	// Fetches user from memberInfo if possible
	user, ok := MemberInfoMap[userMember.ID]
	if !ok {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()

	// Checks usernames and updates if needed
	if user.Username != userMember.Username && userMember.Username != "" {
		flag := true
		lower := strings.ToLower(userMember.Username)

		for _, names := range user.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastUsernames = append(user.PastUsernames, userMember.Username)
			user.Username = userMember.Username
		}
		writeFlag = true
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.Discrim != userMember.Discriminator {
		user.Discrim = userMember.Discriminator
		writeFlag = true
	}

	// Checks if username or discrim were changed, else do NOT write to disk
	if !writeFlag {
		return
	}

	// Saves the updates to memberInfoMap and writes to disk
	MapMutex.Lock()
	MemberInfoMap[userMember.ID] = user
	MemberInfoWrite(MemberInfoMap)
	MapMutex.Unlock()
}

// Encrypt string to base64 crypto using AES
func Encrypt(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fmt.Println(err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt from base64 to decrypted string
func Decrypt(key []byte, cryptoText string) (string, bool) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		fmt.Println("ciphertext too short")
		return "", false
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), true
}

// Cleans up duplicate nicknames and usernames in memberInfo.json
func DuplicateUsernamesAndNicknamesCleanup() {
	MapMutex.Lock()
	DuplicateRecursion()
	MapMutex.Unlock()

	MemberInfoWrite(MemberInfoMap)

	fmt.Println("FINISHED WITH DUPLICATES")
}

// Helper of above
func DuplicateRecursion() {
	for _, value := range MemberInfoMap {
		// Remove duplicate usernames
		for index, username := range value.PastUsernames {
			for indexDuplicate, usernameDuplicate := range value.PastUsernames {
				if index != indexDuplicate && username == usernameDuplicate {
					value.PastUsernames = append(value.PastUsernames[:indexDuplicate], value.PastUsernames[indexDuplicate+1:]...)
					DuplicateRecursion()
					return
				}
			}
		}
		// Remove duplicate nicknames
		for index, nickname := range value.PastNicknames {
			for indexDuplicate, nicknameDuplicate := range value.PastNicknames {
				if index != indexDuplicate && nickname == nicknameDuplicate {
					value.PastNicknames = append(value.PastNicknames[:indexDuplicate], value.PastNicknames[indexDuplicate+1:]...)
					DuplicateRecursion()
					return
				}
			}
		}

	}
}

// Updates user usernames to the current ones they're using in memberInfo.json
func UsernameCleanup(s *discordgo.Session, e *discordgo.Ready) {
	var progress int
	MapMutex.Lock()
	for _, mapUser := range MemberInfoMap {
		user, err := s.User(mapUser.ID)
		if err != nil {
			progress++
			continue
		}
		if mapUser.Username != user.Username {
			mapUser.Username = user.Username
		}
		if mapUser.Discrim != user.Discriminator {
			mapUser.Discrim = user.Discriminator
		}
		progress++
		fmt.Printf("%v out of %v \n", progress, len(MemberInfoMap))
	}
	MapMutex.Unlock()

	MemberInfoWrite(MemberInfoMap)

	fmt.Println("FINISHED WITH USERNAMES")
}