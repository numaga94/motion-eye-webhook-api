package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	app := fiber.New()

	// ~ load env for config settings
	godotenv.Load()
	var (
		port         string
		snapshot_url string
		token        string
		chat_id      string
	)

	if os.Getenv("PORT") == "" {
		port = os.Args[1]
	} else {
		port = os.Getenv("PORT")
	}

	if os.Getenv("SNAPSHOT_URL") == "" {
		snapshot_url = os.Args[2]
	} else {
		snapshot_url = os.Getenv("SNAPSHOT_URL")
	}

	if os.Getenv("TOKEN") == "" {
		token = os.Args[3]
	} else {
		token = os.Getenv("TOKEN")
	}

	if os.Getenv("CHAT_ID") == "" {
		chat_id = os.Args[4]
	} else {
		chat_id = os.Getenv("CHAT_ID")
	}

	// ~ Default middlewares
	app.Use(logger.New())

	// ~ api GET
	app.Get("/", func(c *fiber.Ctx) error {
		// ~ get current photo
		reqSnapshot, _ := http.NewRequest("GET", snapshot_url, nil)

		reqSnapshot.Header.Add("cookie", "motion_detected_1=false; monitor_info_1=; capture_fps_1=0.0")

		resSnapshot, err := http.DefaultClient.Do(reqSnapshot)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"message": err.Error()})
		}

		defer resSnapshot.Body.Close()
		bodySnapshot, _ := io.ReadAll(resSnapshot.Body)

		// fmt.Println(resSnapshot)
		// fmt.Println(string(bodySnapshot))

		// ~ Prepare multipart form data from snapshot response
		// * golang multipart form data references: https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
		values := map[string]io.Reader{
			"chat_id": strings.NewReader(chat_id),
			"photo":   strings.NewReader(string(bodySnapshot)), // lets assume its this file
			"caption": strings.NewReader(fmt.Sprintf("%v 办公室发现异动。", strings.Replace(time.Now().Format(time.RFC3339), "T", " ", 1))),
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		for key, r := range values {
			var (
				fw  io.Writer
				err error
			)
			if x, ok := r.(io.Closer); ok {
				defer x.Close()
			}
			// Add an image file
			if key == "photo" {
				if fw, err = w.CreateFormFile(key, "photo"); err != nil {
					fmt.Println(err.Error())
				}
			} else {
				// Add other fields
				if fw, err = w.CreateFormField(key); err != nil {
					fmt.Println(err.Error())
				}
			}
			if _, err = io.Copy(fw, r); err != nil {
				fmt.Println(err.Error())
			}

		}
		// Don't forget to close the multipart writer.
		// If you don't close it, your request will be missing the terminating boundary.
		w.Close()

		// ~ send current photo to telegram
		url := fmt.Sprintf("https://api.telegram.org/bot%v/sendPhoto", token)

		req, _ := http.NewRequest("POST", url, &b)

		req.Header.Add("Content-Type", w.FormDataContentType())

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"message": err.Error()})
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		// fmt.Println(string(body))
		return c.JSON(body)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%v", port)))
}
