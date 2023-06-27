// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  testing
package kafkamq

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

var (
	JiagouyunRepo     = "pubrepo.jiagouyun.com/image-repo-for-testing/"
	ZookeeperImage    = JiagouyunRepo + "wurstmeister/zookeeper"
	ZookeeperVersion  = "latest"
	zookeeperResource *dockertest.Resource

	kafkaImageName = JiagouyunRepo + "wurstmeister/kafka"
	kafkaVersions  = []string{"2.12-2.5.1", "2.13-2.7.2"}
	versionMap     = map[string]string{
		"2.12-2.5.1": "2.5.1",
		"2.13-2.7.2": "2.7.0",
		"2.11-1.0.1": "1.0.1",
	}

	dockerPool *dockertest.Pool
	NetworkID  string

	testTopic    = "apm"
	dockerRemote = testutils.GetRemote()
)

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	// 创建镜像启动。
	// 接口实现 topicHandler
	// 数据对比
	config := newSaramaConfig(withVersion("2.5.1"),
		withAssignors(""),
		withOffset(-2),
	)

	err := beforeTest(t)
	assert.NoError(t, err)
	time.Sleep(time.Second * 10)
	defer func() {
		afterTest(t)
	}()
	t.Logf("zookeeper running")
	host := dockerRemote.Host
	if host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}
	tests := []struct {
		resource *dockertest.RunOptions
		process  *mockProcess
	}{
		{
			resource: newKafkaImage(kafkaVersions[0], NetworkID, host),
			process: &mockProcess{
				checkOK: make(chan bool),
				t:       t,
				topics:  []string{testTopic},
				count:   0,
				send: &sender{
					kafkaAddr: host + ":9092",
					topic:     testTopic,
					sendNum:   5,
					body:      sarama.StringEncoder("this msg"),
				},
			},
		},
	}

	for _, test := range tests {
		err = dockerPool.RemoveContainerByName("kafka-example")
		assert.NoError(t, err)
		kafkaResource, err := dockerPool.RunWithOptions(test.resource)
		assert.NoError(t, err)
		assert.NotNil(t, kafkaResource)
		t.Logf("time sleep 20s")
		time.Sleep(time.Second * 20)

		kafka := &kafkaConsumer{
			process: make(map[string]TopicProcess),
			topics:  make([]string, 0),
			addrs:   []string{host + ":9092"},
			groupID: "mock_testing",
			stop:    make(chan struct{}),
			config:  config,
			ready:   make(chan bool),
		}

		kafka.registerP(test.process)
		go kafka.start()

		err = test.process.send.sendToKafka(t)
		assert.NoError(t, err)
		select {
		case <-test.process.checkOK:
		case <-time.After(time.Second * 20):
			t.Log("time out ")
		}
		_ = kafkaResource.Close()
		_ = dockerPool.RemoveContainerByName("kafka-example")
	}
}

type mockProcess struct {
	t       *testing.T
	topics  []string
	checkOK chan bool
	count   int
	send    *sender
}

func (m *mockProcess) Init() error {
	return nil
}

func (m *mockProcess) GetTopics() []string {
	m.t.Logf("topics is %v", m.topics)
	return m.topics
}

func (m *mockProcess) Process(msg *sarama.ConsumerMessage) error {
	m.t.Logf("message topic=%s message=%s", msg.Topic, string(msg.Value))
	m.count++
	if m.send.sendNum == m.count {
		m.checkOK <- true
	}
	return nil
}

func beforeTest(t *testing.T) (err error) {
	t.Helper()
	dockerTCP := dockerRemote.TCPURL()
	t.Logf(dockerTCP)
	dockerPool, err = dockertest.NewPool(dockerTCP)
	if err != nil {
		t.Errorf("err = %v", err)
		return
	}

	if err = dockerPool.Client.Ping(); err != nil {
		t.Errorf("could not connect to docker: %v", err)
		return
	}

	// 初始化 zookeeper
	network, err := dockerPool.Client.CreateNetwork(docker.CreateNetworkOptions{Name: "zookeeper_kafka_network"})
	if err != nil {
		t.Errorf("could not create a network to zookeeper and kafka: %s", err)
		return
	}
	NetworkID = network.ID
	zookeeperResource, err = newZookeeperResource("zookeeper", dockerPool, network.ID)
	if err != nil {
		t.Errorf("new zookeeper resource err=%v", err)
		return err
	}
	host := dockerRemote.Host
	if dockerRemote.Host == "0.0.0.0" || dockerRemote.Host == "" {
		host = "127.0.0.1"
	}
	conn, _, err := zk.Connect([]string{fmt.Sprintf("%s:%s", host, zookeeperResource.GetPort("2181/tcp"))}, 10*time.Second)
	if err != nil {
		t.Errorf("could not connect zookeeper: %s", err)
		return err
	}
	defer conn.Close()
	retryFn := func() error {
		switch conn.State() { //nolint
		case zk.StateHasSession, zk.StateConnected:
			return nil
		default:
			return errors.New("not yet connected")
		}
	}

	if err = dockerPool.Retry(retryFn); err != nil {
		t.Logf("could not connect to zookeeper: %s", err)
		return err
	}
	return err
}

func afterTest(t *testing.T) {
	t.Helper()
	if zookeeperResource == nil {
		return
	}
	if err := dockerPool.Purge(zookeeperResource); err != nil {
		t.Errorf("could not purge zookeeperResource: %s", err)
	}
	if err := dockerPool.Client.RemoveNetwork(NetworkID); err != nil {
		t.Errorf("could not remove network: %s, err:%v", NetworkID, err)
	}
}

func newKafkaImage(version string, id string, dockerIP string) *dockertest.RunOptions {
	opts := &dockertest.RunOptions{
		Hostname:     "kafka",
		Name:         "kafka-example",
		Repository:   kafkaImageName,
		Tag:          version,
		ExposedPorts: []string{"9092/tcp", "9093/tcp"},
		NetworkID:    id,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9093/tcp": {{HostIP: "0.0.0.0", HostPort: "9093/tcp"}},
			"9092/tcp": {{HostIP: "0.0.0.0", HostPort: "9092/tcp"}},
		},
	}
	opts.Env = []string{
		"KAFKA_CREATE_TOPICS=domain.test:1:1:compact",
		fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=INSIDE://%s:9092,OUTSIDE://%s:9093", dockerIP, dockerIP),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
		"KAFKA_LISTENERS=INSIDE://:9092,OUTSIDE://:9093",
		fmt.Sprintf("KAFKA_ZOOKEEPER_CONNECT=%s:2181", "zookeeper"),
		"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
		"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
	}
	return opts
}

func newZookeeperResource(name string, dockerPool *dockertest.Pool, id string) (*dockertest.Resource, error) {
	zookeeperResource, ok := dockerPool.ContainerByName(name)
	if !ok {
		var err error
		zookeeperResource, err = dockerPool.RunWithOptions(&dockertest.RunOptions{
			Name:         name,
			Repository:   ZookeeperImage,
			Tag:          ZookeeperVersion,
			NetworkID:    id,
			Hostname:     "zookeeper",
			ExposedPorts: []string{"2181"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"2181/tcp": {{HostIP: "0.0.0.0", HostPort: "2181/tcp"}},
			},
		})
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	return zookeeperResource, nil
}

type sender struct {
	kafkaAddr string
	topic     string
	sendNum   int
	body      sarama.StringEncoder
}

func (send *sender) sendToKafka(t *testing.T) error {
	t.Helper()
	if send.sendNum == 0 {
		send.sendNum = 1
	}
	t.Logf("send = %+v", send)
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	// 发送消息到 kafka
	producer, err := sarama.NewSyncProducer([]string{send.kafkaAddr}, config)
	if err != nil {
		t.Logf("init sarama failed,%#v", err)
		return err
	}
	t.Logf("new producer,send topic:[%s]", send.topic)
	for i := 0; i < send.sendNum; i++ {
		msg := &sarama.ProducerMessage{
			Topic: send.topic,
			Value: send.body,
		}
		p, off, err := producer.SendMessage(msg)
		if err == nil {
			t.Logf("partition=%d offset=%d", p, off)
		} else {
			t.Errorf("producer err =%v", err)
		}
	}
	// 接收消息
	t.Log("to recover message.....")
	producer.Close()
	return nil
}
