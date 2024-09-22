package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var fardbuffer [][]byte
var toiletbuffer [][]byte
var bowelbuffer = make([][]byte, 0)
var devmode bool
var farding bool
var playsoundchannel chan channelinfo
var stoploop chan bool

type soundcollection struct {
	sound []sound
}
type soundType int

const (
	soundbite soundType = iota
	loop
	oneHour
)

type sound struct {
	trigger     string
	buffer      [][]byte
	soundtype   soundType
	description string
}

type channelinfo struct {
	s         *discordgo.Session
	guildID   *discordgo.Guild
	channelID *discordgo.Channel
	author    *discordgo.MessageCreate
	trigger   string
}

var sounds soundcollection

func main() {

	playsoundchannel = make(chan channelinfo)
	stoploop = make(chan bool)

	sounds.LoadSound("fard.dca", "!fard", soundbite)
	sounds.LoadSound("bowel.dca", "!bowel", soundbite)
	sounds.LoadSound("toilet.dca", "!toilet", loop)
	sounds.LoadSound("metal.dca", "!1hourmetalpipe", oneHour)
	sounds.LoadSound("fard.dca", "!1hourfard", oneHour)

	devmode = true

	if token == "" {
		fmt.Println("No token provided. Please run: airhorn -t <bot token>")
		return
	}

	// Load the sound file.

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	go voicefard()

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(MSGlistener)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Airhorn is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.

// loadSound attempts to load an encoded sound file from disk.

// playSound plays the current buffer to the provided channel.
func playSound(buffer [][]byte, s *discordgo.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
func playLoop(buffer [][]byte, s *discordgo.Session, guildID, channelID string) (err error) {
	fmt.Println("loop function reporting in")
	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
outer:
	for {
		for _, buff := range buffer {
			vc.OpusSend <- buff
			select {
			case <-stoploop:
				fmt.Println("loop function received stop signal")
				break outer
			default:
				continue
			}
		}
	}
	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

func oneHourSilence(buffer [][]byte, s *discordgo.Session, guildID, channelID string) (err error) {
	oneHourStart := time.Now()
	fmt.Println("onehour function reporting in")
	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
outer:
	for {
		fmt.Printf("Playing sound after %v seconds \n", time.Since(oneHourStart).Seconds())
		for _, buff := range buffer {
			vc.OpusSend <- buff
			select {
			case <-stoploop:
				fmt.Println("onehour function received stop signal")
				break outer
			default:
				continue
			}
		}
		if time.Since(oneHourStart) > 60*time.Minute {
			break outer
		}
		waitTimer := rand.Intn(60)
		time.Sleep(time.Duration(waitTimer) * time.Second)
	}
	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

func countdown(keepalive chan string, stopchannel chan string) {
	timeLeft := 25 * time.Second
	<-keepalive
	time.Sleep(timeLeft)
	stopchannel <- "stop"
}
func MSGlistener(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!stop") {
		fmt.Println("stop signal received")
		stoploop <- true
		return

	}

	if strings.HasPrefix(m.Content, "+magik") {

		fmt.Println(m.Attachments[0].Filename)
		fmt.Println(m.Attachments[0].URL)
		magikFile := magik(m.Attachments[0].Filename, m.Attachments[0].URL)
		fmt.Println(magikFile)
		reader, err := os.Open(magikFile)
		if err != nil {
			// Could not find channel.
			fmt.Println(err)
			return
		}

		s.ChannelMessageSend(m.ChannelID, "adding some MAGIK:")
		s.ChannelFileSend(m.ChannelID, magikFile, reader)
		os.Remove(magikFile)
		return
	}
	for _, SoundOption := range sounds.sound {

		// check if the message is "!fard"
		if strings.HasPrefix(m.Content, SoundOption.trigger) {

			// Find the channel that the message came from.
			c, err := s.State.Channel(m.ChannelID)
			fmt.Println("joining " + m.ChannelID)
			if err != nil {
				// Could not find channel.
				fmt.Println(err)
				return
			}

			// Find the guild for that channel.
			g, err := s.State.Guild(c.GuildID)

			if err != nil {
				// Could not find guild.
				fmt.Println(err)
				return
			}
			var packageinfo channelinfo
			packageinfo.s = s
			packageinfo.trigger = SoundOption.trigger
			packageinfo.guildID = g
			packageinfo.channelID = c
			packageinfo.author = m
			playsoundchannel <- packageinfo
		}

	}
}

func voicefard() {
	for {
		info := <-playsoundchannel

		for _, SoundOption := range sounds.sound {
			if info.trigger == SoundOption.trigger {

				for _, vs := range info.guildID.VoiceStates {
					if vs.UserID == info.author.Author.ID {
						fmt.Println("playing sound in " + vs.ChannelID)

						switch SoundOption.soundtype {

						case loop:
							err := playLoop(SoundOption.buffer, info.s, info.guildID.ID, vs.ChannelID)
							check(err)
						case soundbite:
							err := playSound(SoundOption.buffer, info.s, info.guildID.ID, vs.ChannelID)
							check(err)
						case oneHour:
							err := oneHourSilence(SoundOption.buffer, info.s, info.guildID.ID, vs.ChannelID)
							check(err)
						}

					}
				}
			}

		}
	}
}

func (sc *soundcollection) LoadSound(path string, trigger string, soundType soundType) {

	var dca sound
	dca.soundtype = soundType
	dca.trigger = trigger
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			sc.sound = append(sc.sound, dca)
			fmt.Println("done with the method, buffer length: ", len(dca.buffer))
			err := file.Close()
			check(err)
			return
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return
		}
		// Append encoded pcm data to the buffer.

		dca.buffer = append(dca.buffer, InBuf)
	}
}

func LoadDCA(path string) (buffer [][]byte) {

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			fmt.Println("done with the method, buffer length: ", len(buffer))
			err := file.Close()
			check(err)
			return
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return
		}
		// Append encoded pcm data to the buffer.

		buffer = append(buffer, InBuf)
	}
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateGameStatus(0, "!fard")
}

func check(e error) {

	if e != nil {
		fmt.Println("Error loading sound: ", e)
		return
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			if *&devmode == false {
				_, _ = s.ChannelMessageSend(channel.ID, "fardbot is ready! Type !fard while in a voice channel to play THE sound.")
			}
			return
		}
	}
}
