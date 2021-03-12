package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/jonas747/dca"
)

var commands = []string{
	"!audio",
	"!audioid",
	"!stop",
	"!clear",
	"!frase",
	"!frasetts",
	"!jogo",
	"!lista",
	"!a",
	"!aid",
}

type guildVoiceInstance struct {
	err        chan error
	connection *discordgo.VoiceConnection
	lastActive time.Time
	isPlaying  bool
	mutex      sync.Mutex
}

var audioArr []Audio
var audioName map[string][]byte
var audioID map[string][]byte

var guildInstances = make(map[string]*guildVoiceInstance)

func getAudioList() string {
	ret := "Áudios disponíveis (em ordem alfabética): \n"
	ret += "```\n"

	for _, a := range audioArr {
		ret += strconv.Itoa(a.id)
		ret += ": "
		ret += a.name
		ret += "\n"
	}

	ret += "\n```"

	return ret
}

func playSound(playing *guildVoiceInstance, audioBuf []byte) error {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 128

	reader := bytes.NewBuffer(audioBuf)

	encodeSession, err := dca.EncodeMem(reader, opts)
	if err != nil {
		fmt.Println("failed to create encode session")
		return err
	}

	playing.err = make(chan error)
	streamSession := dca.NewStream(encodeSession, playing.connection, playing.err)

	err = <-playing.err

	if err != nil && err != io.EOF {
		fmt.Println(err)
		return err
	}

	streamSession.Lock()

	encodeSession.Cleanup()

	streamSession.Unlock()
	return nil
}

func joinVoice(s *discordgo.Session, m *discordgo.MessageCreate, audioBuf []byte) error {
	channelID, err := findVoiceChannel(s, m)
	if err != nil {
		return err
	}

	if playing, ok := guildInstances[m.GuildID]; ok {
		playing.mutex.Lock()

		if !playing.isPlaying {
			if playing.connection == nil {
				playing.connection, err = s.ChannelVoiceJoin(m.GuildID, channelID, false, true)
				if err != nil {
					playing.mutex.Unlock()
					return err
				}
			} else if playing.connection.ChannelID != channelID {
				playing.connection.Close()
				playing.connection, err = s.ChannelVoiceJoin(m.GuildID, channelID, false, true)
				if err != nil {
					playing.mutex.Unlock()
					return err
				}
			}

			playing.isPlaying = true
			playing.connection.Speaking(true)

			playing.mutex.Unlock()

			err = playSound(playing, audioBuf)

			playing.mutex.Lock()

			playing.isPlaying = false
			playing.connection.Speaking(false)
			playing.lastActive = time.Now().UTC()

			playing.mutex.Unlock()

			if err != nil {
				fmt.Println("failed to play sound", err)
				return err
			}

			return nil
		}

		playing.mutex.Unlock()

		err = sendMessage(s, m, geraErroAudioJaTocando())

		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("guild not found")
}

func findVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println("failed to find channel")
		return "", err
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		fmt.Println("failed to find guild")
		return "", err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			return vs.ChannelID, nil
		}
	}

	return "", errors.New("failed to find voice channel")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	split := strings.SplitN(m.Content, " ", 2)
	splitLen := len(split)

	if splitLen == 0 {
		return
	}

	var err error = nil
	cmd := ""
	arg := ""

	if splitLen == 1 {
		cmd = split[0]
	} else if splitLen == 2 {
		cmd = split[0]
		arg = split[1]
	}

	switch {
	case cmd == "!a":
		err = cmdAudio(s, m, arg)

	case cmd == "!audio":
		err = cmdAudio(s, m, arg)

	case cmd == "!aid":
		err = cmdAudioID(s, m, arg)

	case cmd == "!audioid":
		err = cmdAudioID(s, m, arg)

	case cmd == "!stop":
		cmdStop(s, m)

	case cmd == "!clear":
		err = clearMessages(s, m)

	case cmd == "!frase":
		err = sendMessage(s, m, geraFrase())

	case cmd == "!frasetts":
		err = sendMessageTTS(s, m, geraFrase())

	case cmd == "!jogo":
		err = changeGame(s, m, geraJogo())

	case cmd == "!lista":
		err = sendMessage(s, m, getAudioList())
	}

	if err != nil {
		fmt.Println(err)
	}
}

func cmdStop(s *discordgo.Session, m *discordgo.MessageCreate) {
	if playing, ok := guildInstances[m.GuildID]; ok {
		playing.mutex.Lock()
		b := playing.isPlaying
		playing.mutex.Unlock()

		if b {
			playing.err <- io.EOF
		}
	}
}

func cmdAudioID(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	if buf, ok := audioID[msg]; ok {
		return joinVoice(s, m, buf)
	}

	return sendMessage(s, m, fmt.Sprintf("Não encontrei o áudio %s", msg))
}

func cmdAudio(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	if buf, ok := audioName[msg]; ok {
		return joinVoice(s, m, buf)
	}

	return sendMessage(s, m, fmt.Sprintf("Não encontrei o áudio %s", msg))
}

func changeGame(s *discordgo.Session, m *discordgo.MessageCreate, game string) error {
	return s.UpdateGameStatus(0, game)
}

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, msg)
	return err
}

func sendMessageTTS(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	_, err := s.ChannelMessageSendTTS(m.ChannelID, msg)
	return err
}

func clearMessages(s *discordgo.Session, m *discordgo.MessageCreate) error {
	ids := make([]string, 0)

	messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
	if err != nil {
		return err
	}

	for _, message := range messages {
		if message.Author.ID == s.State.User.ID {
			ids = append(ids, message.ID)
		} else {
			for _, cmd := range commands {
				if strings.HasPrefix(message.Content, cmd) {
					ids = append(ids, message.ID)
					break
				}
			}
		}
	}

	return s.ChannelMessagesBulkDelete(m.ChannelID, ids)
}

func disconnectWhenIdle() {
	for _, g := range guildInstances {
		g.mutex.Lock()

		if g.connection != nil && !g.isPlaying {
			diff := time.Now().UTC().Sub(g.lastActive)
			if diff >= time.Second*60 {
				g.connection.Disconnect()
				g.connection.Close()
				g.connection = nil
			}
		}

		g.mutex.Unlock()
	}
}

func disconnectWhenIdleTick() {
	ticker := time.NewTicker(time.Second * 30)

	for {
		_ = <-ticker.C
		disconnectWhenIdle()
	}
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	exPath := filepath.Dir(ex)

	token := os.Getenv("LUQUITO_BOT")
	if len(token) == 0 {
		panic("no token found")
	}

	audioArr, err = readAudioConfig(filepath.Join(exPath, "config.txt"))
	if err != nil {
		log.Fatal("failed to read config\n", err)
	}

	loadAllFiles(exPath, audioArr)

	sort.Slice(audioArr, func(i, j int) bool {
		return audioArr[i].name < audioArr[j].name
	})

	audioName = make(map[string][]byte)
	audioID = make(map[string][]byte)

	for _, a := range audioArr {
		audioName[a.name] = a.buf
		id := strconv.Itoa(a.id)
		audioID[id] = a.buf
	}

	rand.Seed(time.Now().UnixNano())

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("failed to create session")
		panic(err)
	}

	discord.AddHandler(messageHandler)

	fmt.Println("connecting...")

	err = discord.Open()
	if err != nil {
		fmt.Println("failed to open connection")
		panic(err)
	}

	guilds := discord.State.Guilds
	for _, g := range guilds {
		guildInstances[g.ID] = &guildVoiceInstance{
			isPlaying:  false,
			err:        nil,
			connection: nil,
			lastActive: time.Now().UTC(),
		}
	}

	go disconnectWhenIdleTick()

	fmt.Println("ready!")

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("closing...")
	discord.Close()
}
