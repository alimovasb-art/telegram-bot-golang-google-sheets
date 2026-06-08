package main

import (
	"context"
	"log"

	"github.com/alimovasb-art/telegram-bot-golang-google-sheets/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
