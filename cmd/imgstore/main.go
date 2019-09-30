package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgamingfoundation/dwh/imgservice"

	"github.com/gorilla/mux"
)

const PORT = "11535"

func main() {

	st := imgservice.NewImgStore()

	router := mux.NewRouter()
	router.HandleFunc("/imgstore/store", st.StoreHandler).Methods(http.MethodPost)
	router.HandleFunc("/imgstore/load", st.StoreHandler).Methods(http.MethodGet)
	router.HandleFunc("/imgstore/get_check_sum", st.StoreHandler).Methods(http.MethodGet)

	srv := http.Server{
		Handler:           router,
		Addr:              fmt.Sprintf("%s:%d", "0.0.0.0", imgservice.DefaultStorePort),
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	log.Println("listen and serve start")
	log.Println(srv.ListenAndServe().Error())
}
