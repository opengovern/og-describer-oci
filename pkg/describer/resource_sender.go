package describer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"github.com/opengovern/og-util/pkg/es"
	"github.com/opengovern/og-util/proto/src/golang"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	MinBufferSize   int           = 10
	MaxBufferSize   int           = 100
	ChannelSize     int           = 1000
	BufferEmptyRate time.Duration = 5 * time.Second
)

type ResourceSender struct {
	authToken                 string
	logger                    *zap.Logger
	resourceChannel           chan *es.Resource
	resourceIDs               []string
	doneChannel               chan interface{}
	conn                      *grpc.ClientConn
	grpcEndpoint              string
	ingestionPipelineEndpoint string
	jobID                     uint

	client     golang.EsSinkServiceClient
	httpClient *http.Client

	sendBuffer    []*es.Resource
	useOpenSearch bool
}

func NewResourceSender(grpcEndpoint, ingestionPipelineEndpoint string, describeToken string, jobID uint, useOpenSearch bool, logger *zap.Logger) (*ResourceSender, error) {
	rs := ResourceSender{
		authToken:                 describeToken,
		logger:                    logger,
		resourceChannel:           make(chan *es.Resource, ChannelSize),
		resourceIDs:               nil,
		doneChannel:               make(chan interface{}),
		conn:                      nil,
		grpcEndpoint:              grpcEndpoint,
		ingestionPipelineEndpoint: ingestionPipelineEndpoint,
		jobID:                     jobID,
		useOpenSearch:             useOpenSearch,

		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	if err := rs.Connect(); err != nil {
		return nil, err
	}

	go rs.ResourceHandler()
	return &rs, nil
}

func (s *ResourceSender) Connect() error {
	var opts []grpc.DialOption
	if s.authToken != "" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
		opts = append(opts, grpc.WithPerRPCCredentials(oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: s.authToken,
			}),
		}))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(
		s.grpcEndpoint,
		opts...,
	)
	if err != nil {
		return err
	}
	s.conn = conn

	client := golang.NewEsSinkServiceClient(conn)
	s.client = client
	return nil
}

func (s *ResourceSender) ResourceHandler() {
	t := time.NewTicker(BufferEmptyRate)
	defer t.Stop()

	for {
		select {
		case resource := <-s.resourceChannel:
			if resource == nil {
				s.flushBuffer(true)
				s.doneChannel <- struct{}{}
				return
			}

			s.resourceIDs = append(s.resourceIDs, resource.ResourceID)
			s.sendBuffer = append(s.sendBuffer, resource)

			if len(s.sendBuffer) > MaxBufferSize {
				s.flushBuffer(true)
			}
		case <-t.C:
			s.flushBuffer(false)
		}
	}
}

func (s *ResourceSender) sendToBackend(resourcesToSend []es.Doc) {
	grpcCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"resource-job-id": fmt.Sprintf("%d", s.jobID),
	}))

	docs := make([]*anypb.Any, 0, len(resourcesToSend))
	for _, resource := range resourcesToSend {
		docBytes, err := json.Marshal(resource)
		if err != nil {
			s.logger.Error("failed to marshal resource", zap.Error(err))
			continue
		}
		docs = append(docs, &anypb.Any{Value: docBytes})
	}

	_, err := s.client.Ingest(grpcCtx, &golang.IngestRequest{Docs: docs})
	if err != nil {
		s.logger.Error("failed to send resource", zap.Error(err))
		if errors.Is(err, io.EOF) {
			err = s.Connect()
			if err != nil {
				s.logger.Error("failed to reconnect", zap.Error(err))
			}
		}
		return
	}
}

func (s *ResourceSender) flushBuffer(force bool) {
	if len(s.sendBuffer) == 0 {
		return
	}

	if !force && len(s.sendBuffer) < MinBufferSize {
		return
	}

	resourcesToSend := make([]es.Doc, 0, 2*len(s.sendBuffer))

	for _, resource := range s.sendBuffer {
		kafkaResource := resource
		keys, idx := kafkaResource.KeysAndIndex()
		kafkaResource.EsID = es.HashOf(keys...)
		kafkaResource.EsIndex = idx

		lookupResource := es.LookupResource{
			PlatformID:      resource.PlatformID,
			ResourceID:      resource.ResourceID,
			ResourceName:    resource.ResourceName,
			IntegrationType: configs.IntegrationName,
			ResourceType:    strings.ToLower(resource.ResourceType),
			IntegrationID:   resource.IntegrationID,
			DescribedBy:     resource.DescribedBy,
			DescribedAt:     resource.DescribedAt,
			Tags:            resource.CanonicalTags,
		}
		lookupKeys, lookupIdx := lookupResource.KeysAndIndex()
		lookupResource.EsID = es.HashOf(lookupKeys...)
		lookupResource.EsIndex = lookupIdx

		resourcesToSend = append(resourcesToSend, kafkaResource)
		resourcesToSend = append(resourcesToSend, lookupResource)
	}

	s.sendToBackend(resourcesToSend)
	s.sendBuffer = nil
}

func (s *ResourceSender) Finish() {
	s.resourceChannel <- nil
	_ = <-s.doneChannel
	s.conn.Close()
}

func (s *ResourceSender) GetResourceIDs() []string {
	return s.resourceIDs
}

func (s *ResourceSender) Send(resource *es.Resource) {
	s.resourceChannel <- resource
}
