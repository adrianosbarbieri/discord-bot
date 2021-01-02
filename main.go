package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/jonas747/dca"
)

var commands = []string{
	"!audio",
	"!stop",
	"!clear",
	"!frase",
	"!frasetts",
	"!jogo",
	"!lista",
}

var audios map[string][]byte

var guildPlaying = make(map[string]chan error)

func getAudioList() string {
	var lista []string
	for key := range audios {
		lista = append(lista, key)
	}

	sort.Strings(lista)

	ret := "Áudios disponíveis: \n"
	ret += "```\n"

	for _, s := range lista {
		ret += s
		ret += "\n"
	}

	ret += "\n```"

	return ret
}

func playSound(s *discordgo.Session, guildID string, channelID string, audioBuf []byte) (err error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		fmt.Println("Could not join voice channel: ", err)
		return err
	}

	vc.Speaking(true)

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	reader := bytes.NewBuffer(audioBuf)

	encodeSession, err := dca.EncodeMem(reader, opts)
	if err != nil {
		fmt.Println("Could not create encode session: ", err)
		return err
	}

	guildPlaying[guildID] = make(chan error)
	_ = dca.NewStream(encodeSession, vc, guildPlaying[guildID])

	for err := range guildPlaying[guildID] {
		if err != nil && err != io.EOF {
			fmt.Println("An error occured: ", err)
			vc.Speaking(false)
			vc.Disconnect()
			return err
		}

		vc.Speaking(false)
		vc.Disconnect()
		encodeSession.Cleanup()
		return nil
	}

	return nil
}

func joinVoice(s *discordgo.Session, m *discordgo.MessageCreate, audioBuf []byte) error {
	channelID, err := findVoiceChannel(s, m)
	if err != nil {
		return err
	}

	err = playSound(s, m.GuildID, channelID, audioBuf)
	if err != nil {
		fmt.Println("Could not play sound: ", err)
		return err
	}
	return nil
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
		if _, ok := s.VoiceConnections[m.GuildID]; ok {
			err = sendMessage(s, m, GeraErroAudioJaTocando())
		} else if buf, ok := audios[split[1]]; ok {
			err = joinVoice(s, m, buf)
		} else {
			err = sendMessage(s, m, fmt.Sprintf("Não encontrei o áudio %s", split[1]))
		}

	case split[0] == "!stop":
		if _, ok := guildPlaying[m.GuildID]; ok {
			guildPlaying[m.GuildID] <- io.EOF
			delete(guildPlaying, m.GuildID)
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
	err := s.UpdateStatus(0, game)
	if err != nil {
		return err
	}
	return nil
}

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	_, err := s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		return err
	}
	return nil
}

func sendMessageTTS(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	_, err := s.ChannelMessageSendTTS(m.ChannelID, msg)
	if err != nil {
		return err
	}
	return nil
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

	err = s.ChannelMessagesBulkDelete(m.ChannelID, ids)

	if err != nil {
		return err
	}
	return nil
}

func main() {
	token := os.Getenv("LUQUITO_BOT")
	if len(token) == 0 {
		fmt.Println("No token found")
		return
	}

	pathMap, err := readAudioConfig("config.txt")
	if err != nil {
		fmt.Println("Could not read config ", err)
		return
	}

	audios = loadAllFiles(pathMap)

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

	fmt.Println("Ready!")

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Closing...")
	discord.Close()
}
