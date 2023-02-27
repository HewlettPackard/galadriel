package datastore

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
)

//const (
//	defaultTTL      = 60 // seconds
//	harvesterPrefix = "harvester"
//	bundleKey       = "latest_bundle"
//)

type Datastore interface {
	EnsureID(id string) (harvesterID string, err error)
	GetCurrentDigest() ([]byte, error)
	IsLeader(harvesterID string) (bool, error)
	UpdateBundle(bundle *spiffebundle.Bundle) error
	RevertBundleUpdate() error
	Close() error
}

type datastoreConfig struct {
	Endpoints []string
	RootPath  string
	CertPath  string
	KeyPath   string
	Username  string
	Password  string
	Insecure  bool
}

type etcd struct {
	c *clientv3.Client
	s *concurrency.Session
	// logger logrus.Logger
}

func MustNewDatastore(dsConfig datastoreConfig) Datastore {
	var tlsConfig *tls.Config

	// if any of the cert/key paths are present, assume that TLS is intended.
	if dsConfig.CertPath != "" {
		clientCert, err := tls.LoadX509KeyPair(dsConfig.CertPath, dsConfig.KeyPath)
		if err != nil {
			logrus.Fatalf("Error loading client certificate: %v", err)
		}

		caBytes, err := os.ReadFile(dsConfig.RootPath)
		if err != nil {
			logrus.Fatalf("Error loading CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caBytes)

		tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: dsConfig.Insecure,
		}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   dsConfig.Endpoints,
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		TLS:         tlsConfig,
		Username:    dsConfig.Username,
		Password:    dsConfig.Password,
	})
	if err != nil {
		panic(err)
	}

	// start a new leased session
	s, err := concurrency.NewSession(cli)
	if err != nil {
		logrus.Fatalf("Error creating etcd session: %v", err)
	}

	return &etcd{
		c: cli,
		s: s,
	}
}

func (d *etcd) RegisterHarvester() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) EnsureID(id string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) GetCurrentDigest() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) IsLeader(harvesterID string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) UpdateBundle(bundle *spiffebundle.Bundle) error {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) RevertBundleUpdate() error {
	//TODO implement me
	panic("implement me")
}

func (d *etcd) Close() error {
	//TODO implement me
	panic("implement me")
}
