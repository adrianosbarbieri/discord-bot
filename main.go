package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
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
}

type GuildVoiceInstance struct {
	err        chan error
	connection *discordgo.VoiceConnection
	lastActive time.Time // time of the last audio played
	isPlaying  bool
	mutex      sync.Mutex
}

var audioArr []Audio
var audioName map[string][]byte
var audioID map[string][]byte

var guildVoiceConnection = make(map[string]*GuildVoiceInstance)

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

func playSound(playing *GuildVoiceInstance, audioBuf []byte) error {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 128

	reader := bytes.NewBuffer(audioBuf)

	encodeSession, err := dca.EncodeMem(reader, opts)
	if err != nil {
		fmt.Println("Could not create encode session: ", err)
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

	if playing, ok := guildVoiceConnection[m.GuildID]; ok {
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
				fmt.Println("Could not play sound: ", err)
				return err
			}

			return nil
		}

		playing.mutex.Unlock()

		err = sendMessage(s, m, GeraErroAudioJaTocando())

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
		fmt.Println("Could not find channel: ", err)
		return "", err
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		fmt.Println("Could not find guild: ", err)
		return "", err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			return vs.ChannelID, nil
		}
	}

	return "", fmt.Errorf("Could not find voice channel")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	split := strings.SplitN(m.Content, " ", 2)
	if len(split) == 0 {
		return
	}

	var err error = nil

	switch {
	case split[0] == "!audio":
		if buf, ok := audioName[split[1]]; ok {
			err = joinVoice(s, m, buf)
		} else {
			err = sendMessage(s, m, fmt.Sprintf("Não encontrei o áudio %s", split[1]))
		}

	case split[0] == "!audioid":
		if buf, ok := audioID[split[1]]; ok {
			err = joinVoice(s, m, buf)
		} else {
			err = sendMessage(s, m, fmt.Sprintf("Não encontrei o áudio %s", split[1]))
		}

	case split[0] == "!stop":
		if playing, ok := guildVoiceConnection[m.GuildID]; ok {
			playing.mutex.Lock()
			b := playing.isPlaying
			playing.mutex.Unlock()

			if b {
				playing.err <- io.EOF
			}
		}

	case split[0] == "!clear":
		clearMessages(s, m)

	case split[0] == "!frase":
		msg := GeraFrase()
		err = sendMessage(s, m, msg)

	case split[0] == "!frasetts":
		msg := GeraFrase()
		err = sendMessageTTS(s, m, msg)

	case split[0] == "!jogo":
		jogo := GeraJogo()
		err = changeGame(s, m, jogo)

	case split[0] == "!lista":
		lista := getAudioList()
		err = sendMessage(s, m, lista)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func changeGame(s *discordgo.Session, m *discordgo.MessageCreate, game string) error {
	return s.UpdateStatus(0, game)
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
	for _, g := range guildVoiceConnection {
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
	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case _ = <-ticker.C:
			disconnectWhenIdle()
		}
	}
}

func main() {
	token := os.Getenv("LUQUITO_BOT")
	if len(token) == 0 {
		fmt.Println("No token found")
		return
	}

	var err error = nil

	audioArr, err = ReadAudioConfig("config.txt")
	if err != nil {
		fmt.Println("Could not read config ", err)
		return
	}

	LoadAllFiles(audioArr)

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
		panic(err)
	}

	discord.AddHandler(messageHandler)

	fmt.Println("Connecting...")

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	guilds := discord.State.Guilds
	for _, g := range guilds {
		guildVoiceConnection[g.ID] = &GuildVoiceInstance{
			isPlaying:  false,
			err:        nil,
			connection: nil,
			lastActive: time.Now().UTC(),
		}
	}

	go disconnectWhenIdleTick()

	fmt.Println("Ready!")

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Closing...")
	discord.Close()
}
