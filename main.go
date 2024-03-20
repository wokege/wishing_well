package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var blockedWords = []string{
	"chi biet uoc", "chỉ biết ước", "only know wish", "u0c", "ư0c", "uoc", "ước", "wish",
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file")
	}

	log.Printf("Connecting to database...")
	db, err := sql.Open("mysql", os.Getenv("MARIADB_CONNECTION_STRING"))
	gormDb, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	} else {
		log.Printf("Database is up!")
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("error creating Discord session, %s", err)
		return
	}

	discord.AddHandler(func(session *discordgo.Session, msgCreate *discordgo.Connect) {
		log.Printf("Connected to Discord! Logged in as %s#%s\n", discord.State.User.Username, discord.State.User.Discriminator)
	})

	discord.AddHandler(func(session *discordgo.Session, event *discordgo.MessageCreate) {
		if event.Author.ID == session.State.User.ID || event.Author.Bot {
			return
		}

		content := event.Message.Content
		msg := event.Message

		if len(msg.Mentions) > 0 || len(msg.MentionRoles) > 0 || len(msg.MentionChannels) > 0 || msg.MentionEveryone {
			return
		}

		if len(content) > 25 {
			return
		}

		words := strings.Fields(content)

		if len(words) > 5 {
			return
		}

		for _, piece := range words {
			for _, blocked := range blockedWords {
				if strings.Contains(piece, blocked) {
					authorId, _ := strconv.ParseUint(event.Message.Author.ID, 10, 64)
					msgId, _ := strconv.ParseUint(event.Message.ID, 10, 64)

					log.Printf("Detected wish by message %v from %v : %s", msgId, authorId, content)

					// match
					err := gormDb.Transaction(func(tx *gorm.DB) error {
						var user User
						if err := tx.Where(&User{DiscordId: authorId}).FirstOrCreate(&user).Error; err != nil {
							return err
						}

						entry := Log{
							ID:        0,
							UserId:    user.ID,
							MessageId: msgId,
							Count:     -1,
						}

						if err := tx.Create(&entry).Error; err != nil {
							log.Printf("Recorded message %v to database", msgId)
						}

						return nil
					})

					if err != nil {
						log.Printf("Failed to record : %s", err)
					}

					ref := discordgo.MessageReference{MessageID: event.Message.ID}
					msg := discordgo.MessageSend{
						Content:   "Số lần được ước của quý khách vừa giảm đi 1.",
						Reference: &ref,
					}

					session.ChannelMessageSendComplex(event.ChannelID, &msg)

					return
				}
			}
		}
	})

	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection, %s", err)
		return
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}
