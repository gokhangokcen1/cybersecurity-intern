package main

import "fmt"

type Kullanici struct {
	Ad    string
	Yas   int
	Aktif bool
}

type Stat struct {
	Ad     string // "Dayanıklılık", "Zeka", "Sanatçı"
	Seviye int
	XP     int
}

type Task struct {
	Baslik     string
	Tamamlandi bool
	StatAdi    string // Hangi stat'a XP verecel
	XPDegeri   int
}

func main() {

	k := Kullanici{Ad: "Gökhan", Yas: 25, Aktif: true}
	fmt.Println(k.Ad)
	k.Yas = 26

	// k:= Kullanici{"Gökhan", 25, true} // kısa yazım

	gorev := Task{Baslik: "Kitap Oku", Tamamlandi: false, StatAdi: "Zeka", XPDegeri: 5}

	fmt.Println(gorev.XPDegeri)
	fmt.Println(gorev.Tamamlandi)
	xpEkle(&gorev)
	fmt.Println(gorev.XPDegeri)
	gorev.Tamamla()
	fmt.Println(gorev.Tamamlandi)

	// --------------------- MAP ---------------------
	statXP := map[string]int{
		"Dayanıklılık": 50,
		"Zeka":         30,
	}

	statXP["Sanatçı"] = 10
	fmt.Println(statXP["Zeka"])

	// var mı yok mu kontrolü
	deger, varMi := statXP["Güç"]
	if !varMi {
		fmt.Println("Bu stat henüz yok.")
	} else {
		fmt.Println(deger)
	}
}

func xpEkle(t *Task) {
	t.XPDegeri += 10
}

func (t *Task) Tamamla() {
	t.Tamamlandi = true
}
