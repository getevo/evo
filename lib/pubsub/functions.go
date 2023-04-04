package pubsub

import (
	"github.com/getevo/evo/v2/lib/log"
)

var defaultDriver Interface = nil
var drivers = map[string]Interface{}

func SetDefaultDriver(driver Interface) {
	AddDriver(driver)
	defaultDriver = driver
}

func DriverName() string {
	return defaultDriver.Name()
}

func Drivers() map[string]Interface {
	return drivers
}

func Driver(driver string) (Interface, bool) {
	if v, ok := drivers[driver]; ok {
		return v, ok
	}
	return nil, false
}

func Use(driver string) Interface {
	return drivers[driver]
}

func AddDriver(driver Interface) {
	if _, ok := drivers[driver.Name()]; !ok {
		drivers[driver.Name()] = driver
		var err = drivers[driver.Name()].Register()
		if err != nil {
			log.Fatal("unable to initiate pub/sub driver", "name", driver.Name(), "error", err)
		}
	}
	if defaultDriver == nil {
		defaultDriver = driver
	}
}

func Subscribe(topic string, onMessage func(topic string, message []byte, driver Interface), params ...interface{}) {
	defaultDriver.Subscribe(topic, onMessage, params...)
}
func Publish(topic string, message []byte, params ...interface{}) error {
	return defaultDriver.Publish(topic, message, params...)
}

func Register() error {
	return nil
}
