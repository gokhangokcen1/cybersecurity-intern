package main

import (
	"errors"
	"fmt"
)

type Esya struct {
	Isim   string
	Miktar int
}

func EsyaBul(envanter map[string]Esya, isim string) (Esya, error) {
	esya, ok := envanter[isim]
	if !ok {
		return Esya{}, errors.New("envanterde bulunamadı: " + isim)
	}
	return esya, nil
}

func SonucYazdir(esya Esya, err error) {
	if err != nil {
		fmt.Println("Hata:", err)
	} else {
		fmt.Println("Bulundu:", esya)
	}
}

func main() {
	envanter := map[string]Esya{
		"kilic": {Isim: "kilic", Miktar: 1},
		"iksir": {Isim: "iksir", Miktar: 3},
	}

	sonuc, err := EsyaBul(envanter, "kalkan")
	SonucYazdir(sonuc, err)

	sonuc2, err2 := EsyaBul(envanter, "kilic")
	SonucYazdir(sonuc2, err2)

}
