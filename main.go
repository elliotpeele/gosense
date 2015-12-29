package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elliotpeele/gosense/consumer"
	"github.com/elliotpeele/gosense/data"
	"github.com/elliotpeele/gosense/record"
	"github.com/streadway/amqp"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	port         = flag.Int("port", 8080, "port to start server on")
)

func init() {
	flag.Parse()
}

func updateData(data *data.Data) consumer.HandleFunc {
	return func(d amqp.Delivery) error {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)

		rec, err := record.ParseRecord(string(d.Body))
		if err != nil {
			log.Println(err.Error())
			return err
		}

		data.Set(rec.Key(), rec)

		return nil
	}
}

func main() {
	data := data.NewData()

	c, err := consumer.NewConsumer(*uri, *exchange, *exchangeType, *queue, *bindingKey, *consumerTag, updateData(data))
	if err != nil {
		log.Fatalf("%s", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if !strings.HasPrefix(r.URL.Path, "/data") {
			http.NotFound(w, r)
		} else if r.URL.Path == "/data" {
			for k, _ := range data.Snapshot() {
				w.Write([]byte(fmt.Sprintf("%s\n", k)))
			}
			w.WriteHeader(http.StatusOK)
		} else {
			key := strings.Split(r.URL.Path, "/")[2]
			value, ok := data.Get(key)
			if ok {
				rec := value.(*record.Record)
				if err := rec.JSON(w); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("Error: %s\n", err)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}

		log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Printf("shutting down amqp connection")
		if err := c.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
		log.Fatal(err)
	}
}
