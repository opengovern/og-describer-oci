package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/opengovern/og-describer-oci/pkg/describer"
	"github.com/opengovern/og-describer-oci/provider/configs"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/opengovern/og-util/pkg/describe"
	esSinkClient "github.com/opengovern/og-util/pkg/es/ingest/client"
	"github.com/opengovern/og-util/pkg/jq"
	"github.com/opengovern/og-util/pkg/opengovernance-es-sdk"
	"go.uber.org/zap"
)

type Worker struct {
	logger   *zap.Logger
	esClient opengovernance.Client
	jq       *jq.JobQueue

	esSinkClient esSinkClient.EsSinkServiceClient
}

var (
	ManualTriggers = os.Getenv("MANUAL_TRIGGERS")
)

func NewWorker(

	logger *zap.Logger,
	ctx context.Context,
) (*Worker, error) {
	url := os.Getenv("NATS_URL")
	jq, err := jq.New(url, logger)
	if err != nil {
		logger.Error("failed to create job queue", zap.Error(err), zap.String("url", url))
		return nil, err
	}

	topic := configs.JobQueueTopic
	if ManualTriggers == "true" {
		topic = configs.JobQueueTopicManuals
	}
	if err := jq.Stream(ctx, configs.StreamName, " describe job runner queue", []string{topic}, 200000); err != nil {
		logger.Error("failed to create stream", zap.Error(err))
		return nil, err
	}

	w := &Worker{
		logger: logger,
		jq:     jq,
	}

	return w, nil
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("starting to consume")
	topic := configs.JobQueueTopic
	consumer := configs.ConsumerGroup
	if ManualTriggers == "true" {
		topic = configs.JobQueueTopicManuals
		consumer = configs.ConsumerGroupManuals
	}
	consumeCtx, err := w.jq.ConsumeWithConfig(ctx, consumer, configs.StreamName, []string{topic}, jetstream.ConsumerConfig{
		Replicas:          1,
		AckPolicy:         jetstream.AckExplicitPolicy,
		DeliverPolicy:     jetstream.DeliverAllPolicy,
		MaxAckPending:     -1,
		AckWait:           time.Minute * 30,
		InactiveThreshold: time.Hour,
	}, []jetstream.PullConsumeOpt{
		jetstream.PullMaxMessages(1),
	}, func(msg jetstream.Msg) {
		w.logger.Info("received a new job")

		defer msg.Ack()

		ctx, cancel := context.WithTimeoutCause(ctx, time.Minute*25, errors.New("describe worker timed out"))
		defer cancel()

		if err := w.ProcessMessage(ctx, msg); err != nil {
			w.logger.Error("failed to process message", zap.Error(err))
		}
		err := msg.Ack()
		if err != nil {
			w.logger.Error("failed to ack message", zap.Error(err))
		}

		w.logger.Info("processing a job completed")
	})
	if err != nil {
		return err
	}

	w.logger.Info("consuming")

	<-ctx.Done()
	consumeCtx.Drain()
	consumeCtx.Stop()

	return nil
}

func (w *Worker) ProcessMessage(ctx context.Context, msg jetstream.Msg) error {
	startTime := time.Now()
	var input describe.DescribeWorkerInput
	err := json.Unmarshal(msg.Data(), &input)
	if err != nil {
		return err
	}
	runtime.GC()

	w.logger.Info("running job", zap.Uint("id", input.DescribeJob.JobID), zap.String("type", input.DescribeJob.ResourceType), zap.String("providerID", input.DescribeJob.ProviderID))

	err = describer.DescribeHandler(ctx, w.logger, describer.TriggeredByLocal, input)
	endTime := time.Now()

	w.logger.Info("job completed", zap.Uint("id", input.DescribeJob.JobID), zap.String("type", input.DescribeJob.ResourceType), zap.String("providerID", input.DescribeJob.ProviderID), zap.Duration("duration", endTime.Sub(startTime)))
	if err != nil {
		w.logger.Error("failure while running job", zap.Error(err))
		return err
	}

	return nil
}
