package assets

import (
	"embed"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	//go:embed manifests/*
	manifests embed.FS

	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := batchv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func GetCronJobFromFile(name string) *batchv1.CronJob {
	cronjobBytes, err := manifests.ReadFile(name)
	if err != nil {
		panic(err)
	}
	cronjobObject, err := runtime.Decode(appsCodecs.UniversalDecoder(batchv1.SchemeGroupVersion), cronjobBytes)
	if err != nil {
		panic(err)
	}
	return cronjobObject.(*batchv1.CronJob)
}
