package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/garry-sharp/assessment/src/core"
	"github.com/garry-sharp/assessment/src/db"
	"github.com/garry-sharp/assessment/src/logger"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var priceChannel *chan db.Price
var dbName string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Serve(c chan db.Price, dbfn string) {

	dbName = dbfn
	priceChannel = &c
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./src/web/index.html")
	})

	http.HandleFunc("/start", Start)
	http.HandleFunc("/stop", Stop)
	http.HandleFunc("/ws", WSHandler)

	logger.Log("Listening on :3000...", logger.Info, logger.Price)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		logger.Log(err.Error(), logger.Fatal, logger.Price)
	}
}

func Start(w http.ResponseWriter, r *http.Request) {
	quote := r.URL.Query().Get("quote")
	assetsV := r.URL.Query().Get("assets")
	assetsV = strings.ReplaceAll(assetsV, " ", "")
	assets := strings.Split(assetsV, ",")
	durationV := r.URL.Query().Get("duration")
	_duration, err := strconv.ParseInt(durationV, 10, 32)
	if err != nil {
		writeError(w, r, errors.New("Unable to read duration as a number"))
		return
	}

	duration := int32(_duration)
	err = core.VerifyParams(duration, quote, assets)
	if err != nil {
		writeError(w, r, err)
		return
	}

	go func() {
		core.Run(duration, quote, assets, dbName, *priceChannel)
	}()
	write200(w, r)
}

func Stop(w http.ResponseWriter, r *http.Request) {
	logger.Log("Stopping Price Collections", logger.Info, logger.Price)
	core.Stop()
	write200(w, r)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

func write200(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Good"))
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log("Cannot establish websocket connection", logger.Fatal, logger.Price)
	} else {
		for {
			price := <-*priceChannel
			ws.WriteJSON(price)
		}
	}
}
