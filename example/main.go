package main

import (
	"fmt"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/log"
)

func main() {
	evo.Setup()
	log.SetLevel(log.NoticeLevel)
	log.SetWriters(func(message string) {

	})
	//var db = evo.GetDBO()
	/*	var data = map[string]interface{}{}
		db.Raw("SELECT * FROM services").Scan(&data)

		cache.SetDefaultDriver(redis.Driver)
		pubsub.AddDriver(redis.Driver)
		pubsub.SetDefaultDriver(kafka.Driver)

		pubsub.Use("redis").Subscribe("test", func(topic string, message []byte, driver pubsub.Interface) {
			log.Debug("message received", "driver", driver.Name(), "topic", topic, "message", string(message))
		})

		go func() {
			for {
				pubsub.Use("redis").Publish("test", []byte(fmt.Sprint(time.Now().Unix())))
				time.Sleep(1 * time.Second)
			}
		}()

		log.Error("Application has been started", "http", settings.Get("HTTP.WriteTimeout").String(), "bool", true)
		log.SetLevel(log.DebugLevel)*/

	var group = evo.Group("/group").Name("mygroup")
	group.Get("/:id", func(request *evo.Request) interface{} {
		var r = request.Route("mygroup.gregory", 125)
		return r
	}).Name("gregory")

	evo.Get("/struct", func(request *evo.Request) interface{} {
		return struct {
			Text    string `json:"text"`
			Integer int    `json:"integer"`
		}{
			"Hello World", 2023,
		}
	})

	evo.Get("/bytes", func(request *evo.Request) interface{} {
		return []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd'}
	})

	evo.Get("/outcome", func(request *evo.Request) interface{} {
		return []error{fmt.Errorf("my error 1"), fmt.Errorf("my error 2")}
	})
	evo.Run()
}
