package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/42LoCo42/go-zeolite"
	"github.com/42LoCo42/z85m"
	"github.com/bwmarrin/discordgo"

	_ "github.com/breml/rootcerts"
)

func main() {
	tokenFile := os.Args[1]
	authorID := os.Args[2]
	channelID := os.Args[3]

	if err := zeolite.Init(); err != nil {
		panic(err)
	}

	ident, err := zeolite.NewIdentity()
	if err != nil {
		panic(err)
	}
	_ = ident

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		panic(err)
	}

	discord, err := discordgo.New(strings.TrimSpace(string(token)))
	if err != nil {
		panic(err)
	}

	ch := make(chan string, 1)
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID != authorID || m.ChannelID != channelID {
			return
		}
		ch <- m.Content
	})

	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages
	if err := discord.Open(); err != nil {
		panic(err)
	}

	log.Print("Press enter to start...")
	in := bufio.NewReader(os.Stdin)
	in.ReadString('\n')

	log.Print("Opening encrypted stream...")
	adapter := mkAdapter(ch, channelID, discord)
	stream, err := ident.NewStream(&adapter, trustAll)
	if err != nil {
		panic(err)
	}

	log.Print("Stream active. Kallisto is ready.")

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				panic(err)
			}
			log.Print("< ", string(msg))
		}
	}()

	for {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		stream.Send([]byte(strings.TrimSpace(line)))
	}

	log.Print("Shutting down...")
	discord.Close()
	log.Print("Goodbye!")
}

func trustAll(k zeolite.SignPK) (bool, error) {
	return true, nil
}

type Adapter struct {
	Channel   chan string
	ChannelID string
	Discord   *discordgo.Session
	ReadCount int
	MsgBuf    []byte
}

func mkAdapter(channel chan string, channelID string, discord *discordgo.Session) Adapter {
	return Adapter{
		channel,
		channelID,
		discord,
		0,
		[]byte{},
	}
}

func (a *Adapter) Read(p []byte) (n int, err error) {
	if len(a.MsgBuf) > 0 {
		n := copy(p, a.MsgBuf)
		a.MsgBuf = []byte{}
		return n, nil
	}

	enc := <-a.Channel
	dec, err := z85m.Decode([]byte(enc))
	if err != nil {
		return n, err
	}

	if a.ReadCount < 5 {
		a.ReadCount++
		return copy(p, dec), nil
	} else {
		a.MsgBuf = dec[4:]
		return copy(p, dec[:4]), nil
	}
}

func (a *Adapter) Write(p []byte) (n int, err error) {
	enc, err := z85m.Encode(p)
	if err != nil {
		return n, err
	}
	_, err = a.Discord.ChannelMessageSend(a.ChannelID, string(enc))
	if err != nil {
		return n, err
	}

	return len(p), nil
}
