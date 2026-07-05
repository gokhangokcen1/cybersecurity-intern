package main

import (
	"fmt"

	"github.com/gokhangokcen1/rpg-tracker/models"
)

func main() {
	zekaStat := models.Stat{Ad: "Zeka", Seviye: 1, XP: 0}
	t := models.Task{Baslik: "Kitap Okuma", Tamamlandi: false, Stat: &zekaStat, XPDegeri: 10}
	fmt.Println(zekaStat)
	fmt.Println(t)

	zekaStat.XP = 96
	fmt.Println(t.Stat)
}
