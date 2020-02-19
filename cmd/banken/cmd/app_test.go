package cmd

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ropes/banken/pkg/traffic"

	log "github.com/sirupsen/logrus"
)

func TestRequests(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "番犬！！ %q", html.EscapeString(r.URL.Path))
	})
	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "yup")
	})
	http.HandleFunc("/neh/yup", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "yup")
	})
	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
	time.Sleep(1 * time.Second)

	// Start Banken instance
	ctx, can := context.WithCancel(context.Background())
	defer can()
	l := log.New()
	l.SetOutput(os.Stderr)
	b := NewBanken(ctx, 10, 10, "", l)

	ifaces, reqs, err := b.Init(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Launch interface listener
	go func() {
		b.Run(ifaces, reqs)
	}()
	time.Sleep(100 * time.Millisecond)

	// Make 5 http requests
	for i := 0; i < 5; i++ {
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatal("status != 200")
		}
	}
	l.Info("Asserting first AlertDetector state is nominal")
	status := b.getAlertState()
	if reflect.TypeOf(status) != reflect.TypeOf(traffic.NominalStatus{}) {
		t.Errorf("status is not nominal: %v", status)
	}

	l.Info("sending 500 requests")
	for i := 0; i < 500; i++ {
		go func() {
			resp, err := http.Get("http://localhost:8081/ski/hihi")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatal("status != 200")
			}
		}()
	}
	time.Sleep(5 * time.Second)
	l.Info("test AlertDetector state post 500 requests")
	status = b.getAlertState()
	if reflect.TypeOf(status) != reflect.TypeOf(traffic.Alert{}) {
		t.Errorf("status is not alerted: %v", status)
	}

	// Assert that the request count is accurate
	now := time.Now()
	start := now.Add(-5 * time.Minute)
	s := b.ad.GetSpanCount(start, now)
	if s < 505 {
		t.Errorf("timeseries span of [%v-%v] was below expected value: %d", start, now, s)
	}
}
