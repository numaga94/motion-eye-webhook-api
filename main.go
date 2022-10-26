package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	app := fiber.New()

	// ~ load env
	godotenv.Load()

	app.Get("/", func(c *fiber.Ctx) error {
		url := fmt.Sprintf("https://api.telegram.org/bot%v/sendPhoto", os.Getenv("TOKEN"))

		payload := fmt.Sprintf("{\n\t\"chat_id\": \"%v\",\n\t\"caption\": \"a motion detected and captured in attached photo\",\n\t\"photo\": \"%v\"\n}", os.Getenv("CHAT_ID"), os.Getenv("PHOTO"))

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("Content-Type", "application/json")

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)

		fmt.Println(res.Status)
		fmt.Println(string(body))
		return c.JSON(body)
	})

	log.Fatal(app.Listen(":3000"))
}
