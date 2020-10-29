package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"go.xrstf.de/aws_exporter/pkg/metrics"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type options struct {
	accessKeyID string
	secretKey   string
	listenAddr  string
	debugLog    bool
}

func main() {
	opt := options{
		listenAddr: ":9759",
	}

	flag.StringVar(&opt.accessKeyID, "access-key-id", opt.accessKeyID, "AWS access key ID ($AWS_ACCESS_KEY_ID)")
	flag.StringVar(&opt.secretKey, "secret-key", opt.secretKey, "AWS secret key ($AWS_SECRET_KEY)")
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

	creds := credentials.NewStaticCredentials(opt.accessKeyID, opt.secretKey, "")
	config := aws.NewConfig().WithCredentials(creds).WithRegion("eu-central-1")
	sess := session.Must(session.NewSession())

	log.Info("Preparing metrics collector…")
	prometheus.MustRegister(metrics.NewCollector(log, sess, config, []string{"eu-central-1"}))

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on %s…", opt.listenAddr)
	log.Fatal(http.ListenAndServe(opt.listenAddr, nil))
}
