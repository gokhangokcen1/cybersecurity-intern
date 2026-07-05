package models

type Task struct {
	Baslik     string
	Tamamlandi bool
	Stat       *Stat
	XPDegeri   int
}
