package main

import (
	"fmt"
	stdLog "log"
	"net/http"
	"time"

	dwh_common "github.com/corestario/dwh/x/common"
	"github.com/corestario/dwh/x/imgstorage"
	"github.com/gorilla/mux"
)

func main() {
	cfg := dwh_common.ReadCommonConfig(dwh_common.DefaultConfigName, dwh_common.DefaultConfigPath)

	st := imgstorage.NewImgStorage(cfg)

	router := mux.NewRouter()
	router.HandleFunc(dwh_common.StoreImagePath, st.StoreHandler).Methods(http.MethodPost)
	router.HandleFunc(dwh_common.LoadImagePath, st.LoadHandler).Methods(http.MethodGet)
	router.HandleFunc(dwh_common.GetCheckSumPath, st.GetCheckSumHandler).Methods(http.MethodPost)

	srv := http.Server{
		Handler:           router,
		Addr:              fmt.Sprintf("%s:%d", "0.0.0.0", cfg.StoragePort),
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	stdLog.Println("listen and serve start")
	stdLog.Println(srv.ListenAndServe().Error())
}
