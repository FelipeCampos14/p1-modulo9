package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	godotenv "github.com/joho/godotenv"
)

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	fmt.Println("Connected")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

var MessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	ToStruct(string(msg.Payload()), msg)
}

var Topics = [2]string{"Freezer", "Geladeira"}

func randFloats(min, max float64) float64 {
	res := min + rand.Float64()*(max-min)
	return res
}

var Values = [2]float64{randFloats(-30.0, -10.0), randFloats(14.0, -2.0)}

func main() {

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
	opts.SetDefaultPublishHandler(MessageHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	Subscribe("sensor/#", client, MessageHandler)
	for i := 0; i < 4; i++ {
		Publish(client, 2)
	}
	client.Disconnect(250)
}

type Payload struct {
	Id          string    `json:"id"`
	Tipo        string    `json:"tipo"`
	Temperatura float64   `json:"temperatura"`
	Timestamp   time.Time `json:"timestamp"`
}

type PublishPacket struct {
	PacketId   int     `json:"packet-id"`
	TopicName  string  `json:"topic-name"`
	Qos        int     `json:"qos"`
	RetainFlag bool    `json:"retain-flag"`
	Payload    Payload `json:"payload"`
	DupFlag    bool    `json:"duplicated-flag"`
}

func (s *PublishPacket) ToJSON() (string, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func CreatePublishPacket(lojaNum int, tipo string, randomFloat float64) string {
	payload := Payload{
		Id:          fmt.Sprintf("lj%bf%b", lojaNum, rand.IntN(3)),
		Tipo:        tipo,
		Temperatura: randomFloat,
		Timestamp:   time.Now(),
	}

	publishpacket := PublishPacket{
		PacketId:   rand.Int(),
		TopicName:  fmt.Sprintf("sensor/%b/%s", lojaNum, tipo),
		Qos:        1,
		RetainFlag: false,
		Payload:    payload,
		DupFlag:    false,
	}

	content, _ := publishpacket.ToJSON()
	return content
}

func Publish(client MQTT.Client, repTime time.Duration) {
	randomInt := rand.IntN(3)

	for i := 0; i < len(Topics); i++ {
		topicStringf := fmt.Sprintf("sensor/%b/%s", randomInt, Topics[i])

		token := client.Publish(topicStringf, 1, false, CreatePublishPacket(randomInt, Topics[i], Values[i]))

		token.Wait()

		if token.Error() != nil {
			fmt.Printf("Failed to publish to topic: %s", Topics[i])
			panic(token.Error())
		}

	}
	time.Sleep(repTime * time.Second)
}

func Subscribe(topic string, client MQTT.Client, handler MQTT.MessageHandler) {
	token := client.Subscribe(topic, 1, handler)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("Failed to subscribe to topic: %v", token.Error())
		panic(token.Error())
	}
	fmt.Printf("\nSubscribed to topic: %s\n", topic)
}

func ToStruct(jsonStr string, msg MQTT.Message) {
	data := PublishPacket{}

	err := json.Unmarshal([]byte(jsonStr), &data)

	if err != nil {
		log.Println(err)
		return
	}
	if msg.Topic() == "sensor/+/freezer" && (data.Payload.Temperatura < -15 || data.Payload.Temperatura > -25) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C")
	}
	if msg.Topic() == "sensor/+/geladeira" && (data.Payload.Temperatura < -10 || data.Payload.Temperatura > -2) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C")
	}
	if msg.Topic() == "sensor/+/freezer" && (data.Payload.Temperatura < -25) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C", " [ALERTA: Temperatura BAIXA]")
	}
	if msg.Topic() == "sensor/+/freezer" && (data.Payload.Temperatura > -15) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C", " [ALERTA: Temperatura ALTA]")
	}
	if msg.Topic() == "sensor/+/geladeira" && (data.Payload.Temperatura < -2) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C", " [ALERTA: Temperatura BAIXA]")
	}
	if msg.Topic() == "sensor/+/geladeira" && (data.Payload.Temperatura > 10) {
		fmt.Println("\nLj: ", data.Payload.Id, ": ", data.Payload.Tipo, " ", data.Payload.Temperatura, "°C", " [ALERTA: Temperatura ALTA]")
	}

}
