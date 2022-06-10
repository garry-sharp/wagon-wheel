package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/garry-sharp/assessment/src/db"
	"github.com/garry-sharp/assessment/src/logger"
)

var wg sync.WaitGroup

var run *bool

func Run(duration int32, quote string, assets []string, dbfn string, updates chan db.Price) {

	if run == nil {
		r := true
		run = &r
	} else if *run == true {
		return
	}

	if duration < 3000 {
		logger.Log("Duration cannot be less than 3s (3000ms)", logger.Fatal, logger.Price)
	}

	logger.Log(fmt.Sprintf("Calling Run with the following parameters Assets: %s, Quote: %s, Duration: %d", assets, quote, duration), logger.Info, logger.Price)
	*run = true
	db.Init(dbfn)
	gormDB := db.GetDB()

	for *run == true {
		wg.Add(len(assets))
		for _, asset := range assets {
			go func(asset string) {
				if p, err := GetPrice(asset, quote); err != nil {
					wg.Done()
				} else {
					logger.Log(fmt.Sprintf("%s - %.2f %s", asset, p, quote), logger.Info, logger.Price)
					price := db.Price{Quote: quote, Asset: asset, Price: p}
					if update := gormDB.Where(db.Price{Quote: quote, Asset: asset}).Updates(&price); update.RowsAffected == 0 {
						gormDB.Create(&price)
					}
					if updates != nil {
						updates <- price
					}
					wg.Done()
				}
			}(asset)
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		wg.Wait()
	}
}

func Stop() {
	if run == nil {
		r := false
		run = &r
	} else {
		*run = false
	}
}

func IsRunning() bool {
	return *run
}

func VerifyParams(duration int32, quote string, assets []string) error {
	if duration < 3000 {
		return errors.New("Duration must be greater than 3000")
	}
	if len(assets) == 0 {
		return errors.New("At least one asset must be specified")
	}
	return nil
}
