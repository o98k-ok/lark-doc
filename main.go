package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/o98k-ok/aggregation/auth"
	"github.com/o98k-ok/aggregation/doc"
	"github.com/o98k-ok/lazy/v2/alfred"
	qrcode "github.com/skip2/go-qrcode"
)

func format(s string) string {
	v1 := strings.ReplaceAll(s, "<em>", "")
	return strings.ReplaceAll(v1, "</em>", "")
}

func title(entity doc.Entity) string {
	FORMAT := "2006-01-02/03"
	title := format(entity.Title)
	if entity.ViewTS != 0 {
		title = fmt.Sprintf("%s 👀 %s", title, time.Unix(int64(entity.ViewTS), 0).Format(FORMAT))
	} else {
		title = fmt.Sprintf("%s 📝 %s", title, time.Unix(int64(entity.EditTs), 0).Format(FORMAT))
	}
	return title
}

func intro(entity doc.Entity) string {
	intro := format(entity.Intro)
	if len(entity.Owner) != 0 {
		intro = fmt.Sprintf("%s 📢 %s", entity.Owner, intro)
	}
	return intro
}

func entry() {
	cli := alfred.NewApp("lark doc search plugin")

	session := os.Getenv("session")
	count, _ := convertor.ToInt(os.Getenv("count"))
	if count <= 0 {
		count = 9
	}
	cacheFile, err := cachePath()
	if err != nil {
		alfred.ErrItems("cache unavailable", err).Show()
		return
	}

	cli.Bind("trigger", func(s []string) {
		cli := doc.NewLark().CustomSession(session).WithPage(0, 40)
		entities, err := cli.Query("")
		if err != nil {
			alfred.ErrItems("update cache failed", err).Show()
			return
		}

		if err := saveCache(cacheFile, entities); err != nil {
			alfred.Log(err.Error())
			return
		}
		alfred.Log("trigger over")
	})

	cli.Bind("query", func(s []string) {
		if session == "" {
			path, err := auth.SessionPath()
			if err != nil {
				loginItem()
				return
			}
			if _, err := auth.LoadSession(path); err != nil {
				loginItem()
				return
			}
		}
		var entities []doc.Entity
		params := strings.TrimSpace(strings.Join(s, " "))
		if fileContent, err := os.ReadFile(cacheFile); err == nil && len(params) == 0 {
			if err := json.Unmarshal(fileContent, &entities); err != nil {
				return
			}
		} else {
			cli := doc.NewLark().CustomSession(session).WithPage(0, int(count))
			entities, err = cli.Query(params)
			if err != nil {
				if strings.Contains(err.Error(), "run lark-doc login") {
					loginItem()
					return
				}
				alfred.ErrItems("query lark failed", err).Show()
				return
			}
		}

		msg := alfred.NewItems()
		for _, entity := range entities {
			item := alfred.NewItem(title(entity), intro(entity), entity.Url)
			item.Extra = entity.ViewTS
			msg.Append(item)
		}

		msg.Order(func(l, r *alfred.Item) bool {
			return l.Extra.(uint32) > r.Extra.(uint32)
		})

		if len(msg.Items) > int(count) {
			msg.Items = msg.Items[:count]
		}
		msg.Show()
	})

	if err := cli.Run(os.Args); err != nil {
		alfred.ErrItems("alfred run failed", err).Show()
		return
	}
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "logout" {
		if err := logout(); err != nil {
			fmt.Fprintln(os.Stderr, "Logout failed:", err)
			os.Exit(1)
		}
		fmt.Println("Local session and cache removed. This does not revoke the Feishu web session.")
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "login" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		err := login(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Login failed:", err)
			os.Exit(1)
		}
		return
	}
	entry()
}

func login(ctx context.Context) error {
	path, err := auth.SessionPath()
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	session, err := auth.Login(ctx, client, func(payload string) error {
		code, err := qrcode.New(payload, qrcode.Medium)
		if err != nil {
			return err
		}
		fmt.Print(code.ToSmallString(false))
		fmt.Println("Scan with Feishu and confirm on your phone. Ctrl+C to cancel.")
		return nil
	})
	if err != nil {
		return err
	}
	cache, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.Remove(cache); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot clear previous account cache: %w", err)
	}
	if err := auth.SaveSession(path, session); err != nil {
		return err
	}
	fmt.Println("Logged in. Credentials saved in the local credential store.")
	return nil
}
