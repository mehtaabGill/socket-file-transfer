package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
)

//define the global app
var app *fiber.App

func main() {

	//define the applications port and destination address here
	const PORT = "3648"
	const ADDRESS = "192.168.1.70" //CHANGE THIS VALUE ON YOUR MACHINE

	//create template engine
	engine := html.New("./", ".html")

	//create gofiber app
	app = fiber.New(fiber.Config{
		Views: engine,
	})

	setupSocketRoutes()

	app.Get("/", func(c *fiber.Ctx) error {
		//render index.html when the client requests "/"
		return c.Render("index", fiber.Map{
			"adr":  ADDRESS,
			"port": PORT,
		})
	})

	log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%s", PORT)))
}

func setupSocketRoutes() {
	//ws middleware
	wsGroup := app.Group("/socket")

	//middleware to provide access to socket upgrade requests
	wsGroup.Use(func(c *fiber.Ctx) error {

		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", "true")
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	wsGroup.Get("/file-uploader", websocket.New(func(c *websocket.Conn) {
		if c.Locals("allowed") == "true" { //if connection is socket upgrade
			for {

				_, msg, err := c.ReadMessage()
				if err != nil {
					break
				}

				//create the file map.
				//this will store the data received by the socket later on
				var file map[string]string

				err = json.Unmarshal(msg, &file)
				if err != nil {
					fmt.Println(err)
					break
				}

				//rawBytes is the base64 decoded data sent by the client
				//it is the byte array of the contents of the file
				rawBytes, err := base64.StdEncoding.DecodeString(file["data"])
				if err != nil {
					fmt.Println(err)
					break
				}

				//create the new file, write to the file, and clost the file
				newFile, _ := os.Create(file["fileName"])
				newFile.Write(rawBytes)
				newFile.Close()
			}
		}
	}))
}
