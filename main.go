package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	app := fiber.New()

	// ~ Load env for config settings
	if err := godotenv.Load(); err != nil {
		fmt.Println("loading env failed")
	}

	port := os.Getenv("PORT")
	snapshotUrl := os.Getenv("SNAPSHOT_URL")
	switchUrl := os.Getenv("SWITCH_URL")
	chatId := os.Getenv("CHAT_ID")
	token := os.Getenv("TOKEN")
	authKey := os.Getenv("AUTH_KEY")
	currentPath, _ := os.Getwd()

	// ~ Default middlewares
	app.Use(logger.New())
	app.Use(favicon.New(favicon.Config{
		File: fmt.Sprintf("%v/favicon/favicon.ico", currentPath),
	}))

	// ~ Set default switch on/off the api to "ON"
	var SWITCH bool = true

	// ~ api GET
	app.Get("/", func(c *fiber.Ctx) error {
		if !SWITCH {
			return c.Status(400).JSON(fiber.Map{"message": "SWITCH is OFF"})
		}

		// * get current photo
		reqSnapshot, _ := http.NewRequest("GET", snapshotUrl, nil)

		reqSnapshot.Header.Add("cookie", "motion_detected_1=false; monitor_info_1=; capture_fps_1=0.0")

		resSnapshot, err := http.DefaultClient.Do(reqSnapshot)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"message": err.Error()})
		}

		defer resSnapshot.Body.Close()
		bodySnapshot, _ := io.ReadAll(resSnapshot.Body)

		// fmt.Println(resSnapshot)
		// fmt.Println(string(bodySnapshot))

		// * Prepare multipart form data from snapshot response
		// * golang multipart form data references: https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
		values := map[string]io.Reader{
			"chat_id": strings.NewReader(chatId),
			"photo":   strings.NewReader(string(bodySnapshot)), // lets assume its this file
			"caption": strings.NewReader(fmt.Sprintf("%v 发现办公室有人或物正在移动。\n《关闭》监控通知: %v?key=%v&status=off", strings.Replace(time.Now().Format(time.RFC3339), "T", " ", 1), switchUrl, authKey)),
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
		// * Don't forget to close the multipart writer.
		// * If you don't close it, your request will be missing the terminating boundary.
		w.Close()

		// * Send current photo to telegram
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

	// ~ api GET
	app.Get("/switch", func(c *fiber.Ctx) error {
		key := strings.TrimSpace(c.Query("key"))

		if key != authKey {
			return c.Status(401).JSON(fiber.Map{"message": "unauthorized"})
		}

		status := strings.ToUpper(strings.TrimSpace(c.Query("status")))

		var msg string
		currentTime := strings.Replace(time.Now().Format(time.RFC3339), "T", " ", 1)

		if status == "ON" {
			SWITCH = true
			msg = url.QueryEscape(fmt.Sprintf("%v\n《已打开》办公室监控。\n当你本人在办公室的时候，可以关闭监控通知。\n《关闭》监控通知: %v?key=%v&status=off", currentTime, switchUrl, authKey))
		} else {
			SWITCH = false
			msg = url.QueryEscape(fmt.Sprintf("%v\n《已关闭》办公室监控。\n当你离开办公室的时候，请不要忘记打开监控通知。\n《打开》监控通知: %v?key=%v&status=on", currentTime, switchUrl, authKey))
		}

		url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage?chat_id=%v&text=%v", token, chatId, msg)

		req, _ := http.NewRequest("GET", url, nil)

		res, err := http.DefaultClient.Do(req)
		if err != nil || res.Status != "200 OK" {
			return c.Status(400).JSON(fiber.Map{"message": err.Error()})
		}
		defer res.Body.Close()

		return c.JSON(fiber.Map{"message": "Okay"})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%v", port)))
}
