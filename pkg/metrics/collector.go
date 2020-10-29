package metrics

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Collector struct {
	log     *logrus.Logger
	sess    *session.Session
	creds   *credentials.Credentials
	regions []string
}

func NewCollector(log *logrus.Logger, sess *session.Session, creds *credentials.Credentials, regions []string) *Collector {
	return &Collector{
		log:     log,
		sess:    sess,
		creds:   creds,
		regions: regions,
	}
}

func (mc *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *Collector) Collect(ch chan<- prometheus.Metric) {
	mc.log.Debug("Collecting metrics…")

	if err := mc.collect(ch); err != nil {
		mc.log.Errorf("Failed to collect metrics: %v", err)
	}

	mc.log.Debug("Done collecting metrics.")
}

func (mc *Collector) collect(ch chan<- prometheus.Metric) error {
	regions := mc.regions

	creds, err := mc.creds.Get()
	if err != nil {
		return err
	}

	if len(regions) == 0 {
		mc.log.Debug("Listing regions…")
		regionStructs, err := mc.listRegions()
		if err != nil {
			return err
		}

		regions = []string{}
		for _, region := range regionStructs {
			regions = append(regions, *region.RegionName)
		}
	}

	wg := sync.WaitGroup{}

	for _, region := range regions {
		wg.Add(1)

		go func(log logrus.FieldLogger, region string) {
			regionLog := log.WithField("region", region)
			regionLog.Debug("Scanning region…")

			if err := mc.collectRegion(regionLog, creds, ch, region); err != nil {
				regionLog.Errorf("Failed to scan: %v", err)
			}

			wg.Done()
		}(mc.log, region)
	}

	wg.Wait()

	return nil
}

func (mc *Collector) collectRegion(log logrus.FieldLogger, creds credentials.Value, ch chan<- prometheus.Metric, region string) error {
	ch <- constMetric(regionInfo, prometheus.GaugeValue, 1.0, creds.AccessKeyID, region)

	log.Debug("Listing VPCs…")
	vpcs, err := mc.listVPCs(region)
	if err != nil {
		return err
	}

	log.Debugf("Found %d VPCs.", len(vpcs))
	for _, vpc := range vpcs {
		ch <- constMetric(vpcInfo, prometheus.GaugeValue, 1.0, creds.AccessKeyID, region, *vpc.VpcId)

		for _, tag := range vpc.Tags {
			ch <- constMetric(vpcTag, prometheus.GaugeValue, 1.0, creds.AccessKeyID, region, *vpc.VpcId, *tag.Key, *tag.Value)
		}
	}

	log.Debug("Listing subnets…")
	subnets, err := mc.listSubnets(region)
	if err != nil {
		return err
	}

	log.Debugf("Found %d subnets.", len(subnets))
	for _, subnet := range subnets {
		ch <- constMetric(subnetInfo, prometheus.GaugeValue, 1.0, creds.AccessKeyID, region, *subnet.VpcId, *subnet.SubnetId)

		for _, tag := range subnet.Tags {
			ch <- constMetric(subnetTag, prometheus.GaugeValue, 1.0, creds.AccessKeyID, region, *subnet.VpcId, *subnet.SubnetId, *tag.Key, *tag.Value)
		}
	}

	return nil
}

// constMetric just helps reducing code noise
func constMetric(desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)
}

func (mc *Collector) listRegions() ([]*ec2.Region, error) {
	// must use a dummy region even to list regions
	ec2Service := ec2.New(mc.sess, mc.awsConfig("us-west-1"))

	output, err := ec2Service.DescribeRegions(nil)
	if err != nil {
		return nil, err
	}

	return output.Regions, nil
}

func (mc *Collector) listVPCs(region string) ([]*ec2.Vpc, error) {
	ec2Service := ec2.New(mc.sess, mc.awsConfig(region))
	perPage := int64(50)
	result := []*ec2.Vpc{}

	var nextToken *string

	for {
		output, err := ec2Service.DescribeVpcs(&ec2.DescribeVpcsInput{
			MaxResults: &perPage,
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, err
		}

		result = append(result, output.Vpcs...)

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return result, nil
}

func (mc *Collector) listSubnets(region string) ([]*ec2.Subnet, error) {
	ec2Service := ec2.New(mc.sess, mc.awsConfig(region))
	perPage := int64(50)
	result := []*ec2.Subnet{}

	var nextToken *string

	for {
		output, err := ec2Service.DescribeSubnets(&ec2.DescribeSubnetsInput{
			MaxResults: &perPage,
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, err
		}

		result = append(result, output.Subnets...)

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return result, nil
}

func (mc *Collector) awsConfig(region string) *aws.Config {
	config := aws.NewConfig().WithCredentials(mc.creds)
	if region != "" {
		config = config.WithRegion(region)
	}

	return config
}
