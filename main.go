package main

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	mathRand "math/rand"
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
	"!lista2",
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
var audioID map[int][]byte

var guildInstances = make(map[string]*guildVoiceInstance)

func getAudioList() string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString("Áudios disponíveis (em ordem alfabética): \n```\n")

	for _, a := range audioArr {
		strBuilder.WriteString(strconv.Itoa(a.id) + ": " + a.name + "\n")
	}

	strBuilder.WriteString("\n```")

	return strBuilder.String()
}

func getAudioList2() string {
	l := len(audioArr)
	audioArr2 := make([]*Audio, l)

	for i := range audioArr {
		audioArr2[i] = &audioArr[i]
	}

	sort.Slice(audioArr2, func(i, j int) bool {
		return audioArr2[i].id < audioArr2[j].id
	})

	strBuilder := strings.Builder{}
	strBuilder.WriteString("Áudios disponíveis (em ordem numérica): \n```\n")

	for _, a := range audioArr2 {
		strBuilder.WriteString(strconv.Itoa(a.id) + ": " + a.name + "\n")
	}

	strBuilder.WriteString("\n```")

	return strBuilder.String()
}

func playSound(playing *guildVoiceInstance, audioBuf []byte) error {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 128

	reader := bytes.NewBuffer(audioBuf)

	encodeSession, err := dca.EncodeMem(reader, opts)
	if err != nil {
		return err
	}

	playing.err = make(chan error)
	streamSession := dca.NewStream(encodeSession, playing.connection, playing.err)

	err = <-playing.err

	if err != nil && err != io.EOF {
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
			playing.connection, err = s.ChannelVoiceJoin(m.GuildID, channelID, false, true)
			if err != nil {
				playing.mutex.Unlock()
				return err
			}

			playing.isPlaying = true
			err = playing.connection.Speaking(true)
			if err != nil {
				playing.isPlaying = false
				playing.connection.Disconnect()
				playing.connection.Close()
				playing.mutex.Unlock()
				return err
			}

			playing.mutex.Unlock()

			err = playSound(playing, audioBuf)
			if err != nil {
				playing.mutex.Lock()
				playing.isPlaying = false
				playing.connection.Disconnect()
				playing.connection.Close()
				playing.connection = nil
				playing.mutex.Unlock()
				return err
			}

			playing.mutex.Lock()

			playing.isPlaying = false

			err = playing.connection.Speaking(false)
			if err != nil {
				playing.connection.Disconnect()
				playing.connection.Close()
				playing.connection = nil
				playing.mutex.Unlock()
				return err
			}

			playing.lastActive = time.Now().UTC()

			playing.mutex.Unlock()

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
		return "", err
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
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

	arg = strings.Trim(arg, " ")

	switch {
	case cmd == "!a":
		fallthrough
	case cmd == "!audio":
		err = cmdAudio(s, m, arg)

	case cmd == "!aid":
		fallthrough
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

	case cmd == "!lista2":
		err = sendMessage(s, m, getAudioList2())
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
	i, err := strconv.Atoi(msg)
	if err != nil {
		return nil
	}

	if buf, ok := audioID[i]; ok {
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
		<-ticker.C
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
		panic(err)
	}

	loadAllFiles(exPath, audioArr)

	sort.Slice(audioArr, func(i, j int) bool {
		return audioArr[i].name < audioArr[j].name
	})

	audioName = make(map[string][]byte)
	audioID = make(map[int][]byte)

	for _, a := range audioArr {
		audioName[a.name] = a.buf
		id := a.id
		audioID[id] = a.buf
	}

	var b [8]byte
	_, err = cryptoRand.Read(b[:])
	if err != nil {
		panic(err)
	}

	mathRand.Seed(int64(binary.LittleEndian.Uint64(b[:])))

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	discord.AddHandler(messageHandler)

	fmt.Println("connecting...")

	err = discord.Open()
	if err != nil {
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

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	fmt.Println("closing...")
	discord.Close()
}
