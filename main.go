package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	cli.Bind("query", func(s []string) {
		env, err := alfred.GetFlowEnv()
		if err != nil {
			alfred.ErrItems("alfred get envs failed", err).Show()
			return
		}

		cli := doc.NewLark().CustomSession(env.GetAsString("session", "")).WithPage(0, env.GetAsInt("count", 9))
		params := strings.Join(s, " ")
		entities, err := cli.Query(params)
		if err != nil {
			alfred.ErrItems("query lark failed", err)
			return
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
