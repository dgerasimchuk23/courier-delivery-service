package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Start(addr string, router *mux.Router) {
	log.Printf("Сервер запущен на %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %s", err)
	}
}
