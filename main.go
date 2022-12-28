package main

import (
	"fmt"

	"github.com/o98k-ok/aggregation/doc"
)

func main() {
	// fmt.Println(cookie.NewKooky("feishu.cn", "session").Filter())
	entities, err := doc.NewLark().Query("crm")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, entity := range entities {
		fmt.Println(entity)
	}
}
