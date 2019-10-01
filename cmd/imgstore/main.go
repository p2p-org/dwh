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
	st := imgservice.NewImgStore(false)

	router := mux.NewRouter()
	router.HandleFunc(imgservice.StoreImagePath, st.StoreHandler).Methods(http.MethodPost)
	router.HandleFunc(imgservice.LoadImagePath, st.LoadHandler).Methods(http.MethodGet)
	router.HandleFunc(imgservice.GetCheckSumPath, st.GetCheckSumHandler).Methods(http.MethodPost)

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
