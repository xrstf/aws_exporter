package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"go.xrstf.de/aws_exporter/pkg/metrics"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type options struct {
	accessKeyID string
	secretKey   string
	regions     string
	listenAddr  string
	debugLog    bool
}

func main() {
	opt := options{
		listenAddr: ":9759",
	}

	flag.StringVar(&opt.accessKeyID, "access-key-id", opt.accessKeyID, "AWS access key ID ($AWS_ACCESS_KEY_ID)")
	flag.StringVar(&opt.secretKey, "secret-key", opt.secretKey, "AWS secret key ($AWS_SECRET_KEY)")
	flag.StringVar(&opt.regions, "regions", opt.regions, "comma-separated list of regions to scan (if empty, all regions are scanned)")
	flag.StringVar(&opt.listenAddr, "listen", opt.listenAddr, "address and port to listen on")
	flag.BoolVar(&opt.debugLog, "debug", opt.debugLog, "enable more verbose logging")
	flag.Parse()

	// setup logging
	var log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	})

	if opt.debugLog {
		log.SetLevel(logrus.DebugLevel)
	}

	if opt.accessKeyID == "" {
		opt.accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}

	if opt.secretKey == "" {
		opt.secretKey = os.Getenv("AWS_SECRET_KEY")
	}

	if opt.accessKeyID == "" {
		log.Fatal("No access key ID given.")
	}

	if opt.secretKey == "" {
		log.Fatal("No secret key given.")
	}

	regions := []string{}
	if opt.regions != "" {
		regions = strings.Split(opt.regions, ",")
	}

	creds := credentials.NewStaticCredentials(opt.accessKeyID, opt.secretKey, "")
	sess := session.Must(session.NewSession())

	log.Info("Preparing metrics collector…")
	prometheus.MustRegister(metrics.NewCollector(log, sess, creds, regions))

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on %s…", opt.listenAddr)
	log.Fatal(http.ListenAndServe(opt.listenAddr, nil))
}
