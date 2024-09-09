package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/o98k-ok/aggregation/doc"
	"github.com/o98k-ok/lazy/v2/alfred"
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
	cacheFile := os.Getenv("cache_path")
	if cacheFile == "" {
		cacheFile = "lark.json"
	}

	cli.Bind("trigger", func(s []string) {
		cli := doc.NewLark().CustomSession(session).WithPage(0, 40)
		entities, err := cli.Query("")
		if err != nil {
			return
		}

		encoder, err := os.OpenFile(cacheFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer encoder.Close()
		json.NewEncoder(encoder).Encode(entities)
	})

	cli.Bind("query", func(s []string) {
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
				alfred.ErrItems("query lark failed", err)
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
	entry()
}
