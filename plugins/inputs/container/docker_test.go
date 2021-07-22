package container

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
)

func TestGatherDockerMetric(t *testing.T) {
	mock, err := newDockerClient(dockerEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	cList, err := mock.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("container sum: %d\n", len(cList))

	// 串行
	startTime := time.Now()
	for index, container := range cList {
		s1 := time.Now()
		_, err := mock.gather(container)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%d] ID: %s Cost: %s\n", index, container.ID, time.Since(s1))
	}
	t.Logf("\n串行采集总耗时 cost: %s", time.Since(startTime))
}

func TestGatherDockerMetric2(t *testing.T) {
	mock, err := newDockerClient(dockerEndpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	cList, err := mock.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("container sum: %d\n", len(cList))

	// 并发
	startTime1 := time.Now()
	var wg sync.WaitGroup

	for index, container := range cList {
		wg.Add(1)

		go func(idx int, c types.Container) {

			s1 := time.Now()
			_, err := mock.gather(c)
			if err != nil {
				t.Error(err)
			}
			t.Logf("[%d] ID: %s Cost: %s\n", idx, c.ID, time.Since(s1))

			wg.Done()

		}(index, container)

	}

	wg.Wait()
	t.Logf("\n并发采集总耗时 cost: %s", time.Since(startTime1))
}

// func TestGatherDockerStats(t *testing.T) {
// 	mock, err := newDockerClient(dockerEndpoint, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	cList, err := mock.client.ContainerList(context.Background(), types.ContainerListOptions{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Logf("container sum: %d\n", len(cList))

// 	// 串行
// 	startTime := time.Now()
// 	for index, container := range cList {
// 		s1 := time.Now()
// 		_, err := mock.gatherSingleContainerStats(context.Background(), container)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		t.Logf("[%d] ID: %s Cost: %s\n", index, container.ID, time.Since(s1))
// 	}

// 	t.Logf("\n串行http get采集总耗时 cost: %s", time.Since(startTime))
// }
