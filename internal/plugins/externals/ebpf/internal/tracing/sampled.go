package tracing

const (
	SampleDataDogUserReject = "-1"
	SampleDataDogAutoReject = "0"
	SampleDataDogAutoKeep   = "1"
	SampleDataDogUserKeep   = "2"

	SampleW3CKeep   = "1"
	SampleW3CReject = "0"
)

func SampledW3C(v string) bool {
	switch v {
	case SampleW3CKeep:
		return true
	case SampleW3CReject:
		return false
	}
	return true
}

func SampledDataDog(v string) bool {
	switch v {
	case SampleDataDogAutoKeep, SampleDataDogUserKeep:
		return true
	case SampleDataDogAutoReject, SampleDataDogUserReject:
		return false
	}
	return true
}
