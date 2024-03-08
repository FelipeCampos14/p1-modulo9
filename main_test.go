package main

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	godotenv "github.com/joho/godotenv"
	"os"
	"regexp"
	"testing"
	"time"
)

func ReturnRegex(topic string) *regexp.Regexp {
	var re *regexp.Regexp
	switch {
	case topic == "sensor/gases":
		regex := `^\{"packet-id":\d+,"topic-name":"sensor\/gases\/\d+","qos":\d+,"retain-flag":(?:true|false),"payload":\{"current_time":"[^"]+","gases-values":\{"sensor":"[^"]+","unit":"[^"]+","gases-values":\{"carbon_monoxide":\d+\.\d+,"nitrogen_dioxide":\d+\.\d+,"ethanol":\d+\.\d+,"hydrogen":\d+\.\d+,"ammonia":\d+\.\d+,"methane":\d+\.\d+,"propane":\d+\.\d+,"iso_butane":\d+\.\d+\}\}\},"duplicated-flag":(?:true|false)\}$`
		re = regexp.MustCompile(regex)
	case topic == "sensor/radiation":
		regex := `^\{"packet-id":\d+,"topic-name":"sensor\/radiation\/\d+","qos":\d+,"retain-flag":(?:true|false),"payload":\{"current_time":"[^"]+","radiation-values":\{"sensor":"[^"]+","unit":"[^"]+","radiation-values":\{"radiation":\d+\.\d+\}\}\},"duplicated-flag":(?:true|false)\}$`

		re = regexp.MustCompile(regex)
	}
	return re

}

func TestMain(t *testing.T) {
	broker := os.Getenv("BROKER_ADDR")
	username := os.Getenv("HIVE_USER")
	password := os.Getenv("HIVE_PSWD")

	if username == "" || password == "" {
		err := godotenv.Load(".env")
		if err != nil {
			fmt.Printf("\nError loading .env file. error: %s\n", err)
		}
		broker = os.Getenv("BROKER_ADDR")
		username = os.Getenv("HIVE_USER")
		password = os.Getenv("HIVE_PSWD")
	}
	var port = 8883
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", broker, port))
	opts.SetClientID("prova1")
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	t.Run("TestReceiveQos", func(t *testing.T) {
		var messageReceiveQosHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

			if msg.Qos() != 1 {
				t.Error("QoS is not 1")
			}
			fmt.Printf("QoS recebido: %d\n", msg.Qos())
		}
		opts.SetDefaultPublishHandler(MessageHandler)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		if token := client.Subscribe("sensor/#", 1, messageReceiveQosHandler); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			return
		}

		// Subscribe("sensor/#", client, MessageHandler)
		Publish(client, 2)

		time.Sleep(2 * time.Second)
		client.Disconnect(250)

	})
}
