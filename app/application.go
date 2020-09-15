package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/retailerTool/config"
	"github.com/retailerTool/crawlerPackage"
	"github.com/retailerTool/util"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"time"
)

func createCrawler() crawlerPackage.Crawler {
	return crawlerPackage.Crawler{
		Logger: util.StubLogger{},
	}
}

func createDb() *sqlx.DB {
	db, err := sql.Open("mysql", config.DbConfig.FormatDSN())
	if err != nil {
		fmt.Println("Unable to open mysql connection")
		os.Exit(-1)
	}
	sqlxDb := sqlx.NewDb(db, "mysql")
	return sqlxDb
}

type tgUserStorage struct {
	tgUsers []*tgUser
}

func (storage *tgUserStorage) findUserById(id int) *tgUser {
	for _, v := range storage.tgUsers {
		if v.tgId == id {
			return v
		}
	}
	return nil
}

func (storage *tgUserStorage) getUserById(id int) *tgUser {
	user := storage.findUserById(id)
	if user == nil {
		storage.saveUser(tgUser{
			params: TgUserParams{
				Price: Range{
					Min: 0,
					Max: 99999999999,
				},
				Rooms: Range{
					Min: 0,
					Max: 99999999999,
				},
				Districts: []string{},
				Type:      "rent",
			},
			tgId: id,
		})
	}
	return user
}

func (storage *tgUserStorage) saveUser(user tgUser) {
	storage.tgUsers = append(storage.tgUsers, &user)
}

func (storage *tgUserStorage) deleteUser(id int) {
	for k, v := range storage.tgUsers {
		if v.tgId == id {
			storage.tgUsers = append(storage.tgUsers[:k], storage.tgUsers[k+1:]...)
			return
		}
	}
}

type ApplicationContext struct {
	Db    *sqlx.DB
	users tgUserStorage
}

type crawlerJob struct {
	crawler crawlerPackage.Crawler
	command crawlerPackage.Command
}

type Range struct {
	Min int
	Max int
}

type TgUserParams struct {
	Price     Range
	Rooms     Range
	Districts []string
	Type      string
}

type tgUser struct {
	params TgUserParams
	tgId   int
	sender *tb.User
}

func (context *ApplicationContext) deleteTgUser(id int) {
	_, _ = context.Db.Exec("DELETE FROM tgUsers WHERE tg_id = ?;", id)
}

func (context *ApplicationContext) activateTgUser(id int) {
	_, _ = context.Db.Exec("UPDATE tgUsers SET active = TRUE WHERE tg_id = ?;", id)
}

func (context *ApplicationContext) deactivateTgUser(id int) {
	_, _ = context.Db.Exec("UPDATE tgUsers SET active = FALSE WHERE tg_id = ?;", id)
}

func (context *ApplicationContext) saveTgUser(user tgUser) {
	sqlQuery := "INSERT INTO tgUsers (tg_id, json_value) VALUES (?, ?) ON DUPLICATE KEY UPDATE json_value = VALUES(json_value);"
	jsonParams, _ := json.Marshal(user.params)
	context.Db.Exec(sqlQuery, user.tgId, jsonParams)
}

func (job *crawlerJob) run(ch chan<- []crawlerPackage.Flat) {
	flatStorage := job.crawler.RunCommand(job.command)
	ch <- flatStorage.GetAll()
}

func (context *ApplicationContext) getDaysSinceLastScan(jobType crawlerPackage.JobType) (days int) {
	sqlQuery := "SELECT added_dt FROM logs WHERE type := ? ORDER BY id DESC LIMIT 1;"
	var lastTime time.Time
	context.Db.QueryRow(sqlQuery, jobType).Scan(&lastTime)
	days = int(time.Now().Sub(lastTime).Hours() / 24)
	return
}

func (context *ApplicationContext) createRentJob() *crawlerJob {
	command := crawlerPackage.GetDefaultRigaRuRentJob(context.getDaysSinceLastScan(crawlerPackage.RentJob))
	return &crawlerJob{
		crawler: createCrawler(),
		command: command,
	}
}

func (context *ApplicationContext) createSellJob() *crawlerJob {
	command := crawlerPackage.GetDefaultRigaRuSellJob(context.getDaysSinceLastScan(crawlerPackage.SellJob))
	return &crawlerJob{
		crawler: createCrawler(),
		command: command,
	}
}

func (context *ApplicationContext) RunApplication() {
	ch := make(chan []crawlerPackage.Flat)

	tickerScan := time.NewTicker(10 * time.Minute)
	go func() {
		for t := range tickerScan.C {
			_ = t
			go context.createRentJob().run(ch)
			go context.createSellJob().run(ch)
		}
	}()

	tickerUsers := time.NewTicker(5 * time.Minute)

	go func() {
		for t := range tickerUsers.C {
			_ = t
			for _, v := range context.users.tgUsers {
				context.saveTgUser(*v) //TODO write normal SQL to save users with single query
			}
		}
	}()

	bot, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 30 * time.Second},
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	prettyFlatPrint := func(flat crawlerPackage.Flat) string {
		return flat.District + " " + flat.Street +
			" Price: " + strconv.Itoa(flat.Price) + "\n" +
			"Rooms: " + strconv.Itoa(flat.Rooms) + " Area: " + strconv.Itoa(flat.ApartmentArea) + "\n" +
			"Text: " + flat.Text +
			"Link: " + flat.Url
	}

	go func() {
		flats := <-ch
		//TODO fix this shit code
		go func() {
			crawlerPackage.NewFlatStorageFromFlats(flats).Save(context.Db)
		}()
		validUsers := []tgUser{}
		validFlats := []crawlerPackage.Flat{}
		// filter bad users
		for _, v := range context.users.tgUsers {
			if v.params.Districts != nil {
				validUsers = append(validUsers, *v)
			}
		}
		// filter flats that not useful for us
		for _, vf := range flats {
			for _, vu := range validUsers {
				if vu.params.Price.Min < vf.Price && vu.params.Price.Max > vf.Price {
					if vu.params.Districts[0] == vf.District {
						if vu.params.Type == vf.Type { // rent or sell
							validFlats = append(validFlats, vf)
							break
						}
					}
				}
			}
		}

		for _, vu := range validUsers {
			userFlats := []crawlerPackage.Flat{}
			for _, vf := range validFlats {
				if vu.params.Price.Min < vf.Price && vu.params.Price.Max > vf.Price {
					if vu.params.Districts[0] == vf.District {
						userFlats = append(userFlats, vf)
					}
				}
			}
			for _, flat := range userFlats {
				bot.Send(vu.sender, prettyFlatPrint(flat))
			}
		}
	}()

	getUserIdFromMsg := func(m *tb.Message) int {
		return m.Sender.ID
	}

	bot.Handle("/delete", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		id := getUserIdFromMsg(m)
		context.deleteTgUser(id)
		context.users.deleteUser(id)
		bot.Send(m.Sender, "All your settings were deleted")
	})

	bot.Handle("/minPrice", func(m *tb.Message) {
		if !m.Private() {
			return
		} // TODO remove copy paste
		id := getUserIdFromMsg(m)
		user := context.users.getUserById(id)
		user.sender = m.Sender
		payload, err := strconv.Atoi(m.Payload)
		if err != nil {
			bot.Send(m.Sender, "Unable to parse your number")
			return
		}
		user.params.Price.Min = payload
		bot.Send(m.Sender, "Your min price: "+strconv.Itoa(payload))
	})

	bot.Handle("/maxPrice", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		id := getUserIdFromMsg(m)
		user := context.users.getUserById(id)
		user.sender = m.Sender
		payload, err := strconv.Atoi(m.Payload)
		if err != nil {
			bot.Send(m.Sender, "Unable to parse your number")
			return
		}
		user.params.Price.Max = payload
		bot.Send(m.Sender, "Your max price: "+strconv.Itoa(payload))
	})

	bot.Handle("/deactivate", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		id := getUserIdFromMsg(m)
		context.deactivateTgUser(id)
		bot.Send(m.Sender, "You will no longer receive notifications")
	})

	bot.Handle("/activate", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		id := getUserIdFromMsg(m)
		context.deactivateTgUser(id)
		bot.Send(m.Sender, "You will receive new notifications")
	})

	bot.Handle("/district", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		id := getUserIdFromMsg(m)
		user := context.users.getUserById(id)
		user.params.Districts = []string{m.Payload} // TODO support more than one district
		user.sender = m.Sender
		bot.Send(m.Sender, "Your scan district is : "+m.Payload)
		bot.Send(m.Sender, "Please make sure that district name is correct according to SS.lv ru cite")
	})

	bot.Start()
}

func (context *ApplicationContext) getDb() *sqlx.DB {
	if context.Db == nil {
		context.Db = createDb()
	}
	return context.Db
}
