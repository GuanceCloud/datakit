package trace

import (
	"fmt"
	"strings"
)

func (z *ZipkinTracer) Decode(octets []byte) error {
	switch z.ContentType {
	case "application/x-protobuf":
		return z.parseZipkinProtobuf(octets)
	case "application/json":
		return z.parseZipkinJson(octets)
	case "application/x-thrift":
		return z.parseZipkinThrift(octets)
	default:
		return fmt.Errorf("Zipkin unsupported content-type: %s", z.ContentType)
	}
}

func (z *ZipkinTracer) parseZipkinProtobuf(octets []byte) error {
	version := strings.ToUpper(z.Version)
	switch version {
	case "V2","":
		return z.parseZipkinProtobufV2(octets)
	default:
		return fmt.Errorf("Zipkin content-type: application/x-protobuf unsuportted version: %s", z.Version)
	}
	return nil
}

func (z *ZipkinTracer) parseZipkinJson(octets []byte) error {
	version := strings.ToUpper(z.Version)
	switch version {
	case "V2":
		return z.parseZipkinJsonV2(octets)
	case "V1":
		return z.parseZipkinJsonV1(octets)
	default:
		return fmt.Errorf("Zipkin content-type: application/json unsuportted version %s", z.Version)
	}
}

func (z *ZipkinTracer) parseZipkinThrift(octets []byte) error {
	version := strings.ToUpper(z.Version)
	switch version {
	case "V1", "":
		return z.parseZipkinThriftV1(octets)
	default:
		return fmt.Errorf("Zipkin content-type: application/x-thrift unsuportted version %s", z.Version)
	}
}