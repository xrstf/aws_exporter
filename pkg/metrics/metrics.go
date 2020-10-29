package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	regionInfo = prometheus.NewDesc(
		"aws_region",
		"Metadata regarding a region",
		[]string{"account", "region"},
		nil,
	)

	vpcInfo = prometheus.NewDesc(
		"aws_vpc",
		"Metadata regarding a VPC",
		[]string{"account", "region", "vpc_id"},
		nil,
	)

	vpcTag = prometheus.NewDesc(
		"aws_vpc_tag",
		"Metadata regarding a VPC tag",
		[]string{"account", "region", "vpc_id", "tag_key", "tag_value"},
		nil,
	)

	subnetInfo = prometheus.NewDesc(
		"aws_subnet",
		"Metadata regarding a subnet",
		[]string{"account", "region", "vpc_id", "subnet_id"},
		nil,
	)

	subnetTag = prometheus.NewDesc(
		"aws_subnet_tag",
		"Metadata regarding a subnet tag",
		[]string{"account", "region", "vpc_id", "subnet_id", "tag_key", "tag_value"},
		nil,
	)
)
