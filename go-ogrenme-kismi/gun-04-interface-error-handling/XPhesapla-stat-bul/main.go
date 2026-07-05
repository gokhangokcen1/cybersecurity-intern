package main

import (
	"errors"
	"fmt"
)

type Stat struct {
	Ad     string
	Seviye int
	XP     int
}

func StatBul(stats map[string]Stat, isim string) (Stat, error) {
	stat, ok := stats[isim]
	if !ok {
		return Stat{}, errors.New("Stat bulunamadı: " + isim)
	}
	return stat, nil
}

func SonucYazdir(stat Stat, err error) {
	if err != nil {
		fmt.Println("Hata:", err)
	} else {
		fmt.Println("Bulundu:", stat)
	}
}

type XPVerici interface {
	XPHesapla() int
}

type OkumaGorevi struct {
	SayfaSayisi int
}

type SporGorevi struct {
	Dakika int
}

func (o *OkumaGorevi) XPHesapla() int {
	if o.SayfaSayisi < 20 {
		return 5
	} else {
		return 10
	}
}

func (s *SporGorevi) XPHesapla() int {
	return s.Dakika / 6
}

func main() {
	gorevler := []XPVerici{
		&OkumaGorevi{SayfaSayisi: 50},
		&SporGorevi{Dakika: 90},
	}

	for _, g := range gorevler {
		fmt.Println(g.XPHesapla())
	}

	stats := map[string]Stat{
		"dayaniklilik": {Ad: "Dayanıklılık", Seviye: 1, XP: 10},
		"zeka":         {Ad: "Zeka", Seviye: 7, XP: 86},
	}

	sonuc, err := StatBul(stats, "atletizm")
	SonucYazdir(sonuc, err)

	sonuc2, err2 := StatBul(stats, "dayaniklilik")
	SonucYazdir(sonuc2, err2)
}
