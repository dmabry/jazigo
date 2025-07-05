package dev

import (
	"fmt"
	"time"

	"github.com/udhos/jazigo/conf"
)

// Spawner launches new goroutines to fetch requests received on channel reqChan.
func Spawner(tab DeviceUpdater, logger hasPrintf, reqChan chan FetchRequest, repository, logPathPrefix string, options *conf.Options, ft *FilterTable) {

	logger.Printf("Spawner: starting")

	for {
		req, ok := <-reqChan
		if !ok {
			logger.Printf("Spawner: request channel closed")
			break
		}

		replyChan := req.ReplyChan // alias

		devID := req.ID
		d, getErr := tab.GetDevice(devID)
		if getErr != nil {
			if replyChan != nil {
				now := time.Now()
				replyChan <- FetchResult{DevID: devID, Msg: fmt.Sprintf("Spawner: could not find device: %v", getErr), Code: fetchErrGetDev, Begin: now, End: now}
			}
			continue
		}

		opt := options.Get()                                                      // get current global data
		go d.Fetch(tab, logger, replyChan, 0, repository, logPathPrefix, opt, ft) // spawn per-request goroutine
	}

	logger.Printf("Spawner: exiting")
}

// Scan scans the list of devices dispatching backup requests to the Spawner thru the request channel reqChan.
func Scan(tab DeviceUpdater, devices []*Device, logger hasPrintf, opt *conf.AppConfig, reqChan chan FetchRequest) (int, int, int) {

	deviceCount := len(devices)
	if deviceCount < 1 {
		logger.Printf("Scan: empty device list, aborting")
		return 0, 0, 0
	}

	begin := time.Now()
	wait := 0       // requests pending
	nextDevice := 0 // device iterator
	req := FetchRequest{ReplyChan: make(chan FetchResult)}
	maxConcurrency := opt.MaxConcurrency // alias
	holdtime := opt.Holdtime             // alias
	elapMax := 0 * time.Second
	elapMin := 24 * time.Hour
	success := 0
	skipped := 0
	deleted := 0

	for nextDevice < deviceCount || wait > 0 {
		// launch requests
		for ; nextDevice < deviceCount; nextDevice++ {
			if maxConcurrency > 0 && wait >= maxConcurrency {
				break // max concurrent limit reached
			}

			d := devices[nextDevice]

			if d.Deleted {
				deleted++
				continue
			}

			if h := d.Holdtime(time.Now(), holdtime); h > 0 {
				// do not handle device yet (holdtime not expired)
				logger.Printf("Scan: %s skipping due to holdtime=%s", d.ID, h)
				skipped++
				continue
			}

			req.ID = d.ID
			reqChan <- req

			wait++ // launched
			logger.Printf("Scan: launched: %s count=%d/%d wait=%d max=%d", req.ID, nextDevice, deviceCount, wait, maxConcurrency)
		}

		if wait < 1 {
			continue
		}

		// wait one response
		r := <-req.ReplyChan
		wait-- // received

		end := time.Now()
		elap := end.Sub(r.Begin)
		logger.Printf("Scan: recv %s %s %s %s msg=[%s] code=%d wait=%d remain=%d skipped=%d elap=%s", r.Model, r.DevID, r.DevHostPort, r.Transport, r.Msg, r.Code, wait, deviceCount-nextDevice, skipped, elap)

		good := r.Code == fetchErrNone

		if good {
			success++
		}
		if elap < elapMin {
			elapMin = elap
		}
		if elap > elapMax {
			elapMax = elap
		}
	}

	elapsed := time.Since(begin)
	average := elapsed / time.Duration(deviceCount)

	logger.Printf("Scan: finished elapsed=%s devices=%d success=%d skipped=%d deleted=%d average=%s min=%s max=%s", elapsed, deviceCount, success, skipped, deleted, average, elapMin, elapMax)

	return success, deviceCount - success, skipped + deleted
}

func updateDeviceStatus(tab DeviceUpdater, devID string, good bool, last time.Time, elapsed time.Duration, logger hasPrintf, holdtime time.Duration) {
	d, getErr := tab.GetDevice(devID)
	if getErr != nil {
		logger.Printf("updateDeviceStatus: '%s' not found: %v", devID, getErr)
		return
	}

	now := time.Now()
	h1 := d.Holdtime(now, holdtime)

	d.lastTry = last
	d.lastElapsed = elapsed
	d.lastStatus = good
	if d.lastStatus {
		d.lastSuccess = d.lastTry
	}

	tab.UpdateDevice(d)

	h2 := d.Holdtime(now, holdtime)
	logger.Printf("updateDeviceStatus: device %s holdtime: old=%v new=%v", devID, h1, h2)
}
