package main

import "fmt"

type Stat struct {
	Ad     string // Dayanıklılık
	Seviye int    // 10. seviye
	XP     int    // 75
	// başarımlar
}

type Task struct {
	GorevAdi   string // Antrenman
	Tamamlandi bool   // Tamamlandı
	XPDegeri   int    // 10
	StatAdi    string // Hangi stata XP verecek
}

func main() {
	gokhale := Stat{Ad: "Dayanıklılık", Seviye: 1, XP: 0}
	fmt.Println(gokhale)
	ilkGorev := Task{GorevAdi: "10 Pull up", Tamamlandi: false, XPDegeri: 20, StatAdi: "Dayanıklılık"}
	ilkGorev.Tamamla()
	gokhale.XPEkle(ilkGorev.XPDegeri)
	fmt.Println(gokhale)

	gorevler := []Task{ilkGorev, Task{GorevAdi: "10 Pull up", Tamamlandi: false, XPDegeri: 20}, Task{GorevAdi: "20 Pull up", Tamamlandi: false, XPDegeri: 30}, Task{GorevAdi: "30 Pull up", Tamamlandi: false, XPDegeri: 40}}

	for _, gorev := range gorevler {
		if gorev.Tamamlandi == false {
			fmt.Println(gorev)
		}
	}
}

func (t *Task) Tamamla() {
	t.Tamamlandi = true
}

func (s *Stat) XPEkle(miktar int) {
	s.XP += miktar
}
