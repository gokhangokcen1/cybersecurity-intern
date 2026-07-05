package main

import "fmt"

type SesCikaran interface {
	SesCikar() string
}

type Kopek struct {
	Isim string
}

func (k *Kopek) SesCikar() string {
	return k.Isim + ": Hav hav!"
}

type Kedi struct {
	Isim string
}

func (kd *Kedi) SesCikar() string {
	return kd.Isim + ": Miyav!"
}

func main() {
	hayvanlar := []SesCikaran{
		&Kopek{Isim: "Karabas"},
		&Kedi{Isim: "Pamuk"},
	}
	for _, h := range hayvanlar {
		fmt.Println(h.SesCikar())
	}

}
