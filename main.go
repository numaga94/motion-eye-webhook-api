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
	"github.com/joho/godotenv"
)

func main() {
	app := fiber.New()

	// ~ load env
	godotenv.Load()

	app.Get("/", func(c *fiber.Ctx) error {
		// get current photo
		reqSnapshot, _ := http.NewRequest("GET", os.Getenv("SNAPSHOT_URL"), nil)

		reqSnapshot.Header.Add("cookie", "motion_detected_1=false; monitor_info_1=; capture_fps_1=0.0")

		resSnapshot, _ := http.DefaultClient.Do(reqSnapshot)

		defer resSnapshot.Body.Close()
		bodySnapshot, _ := io.ReadAll(resSnapshot.Body)

		// fmt.Println(resSnapshot)
		// fmt.Println(string(bodySnapshot))

		// send current photo to telegram
		// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
		values := map[string]io.Reader{
			"chat_id": strings.NewReader(os.Getenv("CHAT_ID")),
			"photo":   strings.NewReader(string(bodySnapshot)), // lets assume its this file
			"caption": strings.NewReader(fmt.Sprintf("%v 办公室发现异动。", time.Now().Format(time.RFC3339))),
		}

		// Prepare a form that you will submit to that URL.
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
				fmt.Println(err)
			}

		}
		// Don't forget to close the multipart writer.
		// If you don't close it, your request will be missing the terminating boundary.
		w.Close()

		url := fmt.Sprintf("https://api.telegram.org/bot%v/sendPhoto", os.Getenv("TOKEN"))

		req, _ := http.NewRequest("POST", url, &b)

		req.Header.Add("Content-Type", w.FormDataContentType())

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err.Error())
		}

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		// fmt.Println(string(body))
		return c.JSON(body)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
