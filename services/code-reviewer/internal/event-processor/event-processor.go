package event_processor

import (
	"context"
	"encoding/json"
	"go_code_reviewer/pkg/kafka"
	"go_code_reviewer/pkg/log"
	"go_code_reviewer/services/api-gateway/pkg/models"
	"go_code_reviewer/services/code-reviewer/internal/assistant"
	"go_code_reviewer/services/code-reviewer/internal/embedder"
	"go_code_reviewer/services/code-reviewer/internal/errors"
	"go_code_reviewer/services/code-reviewer/internal/metrics"
	"go_code_reviewer/services/code-reviewer/internal/parser"
	"go_code_reviewer/services/code-reviewer/internal/vsc"
	"time"
)

type Module struct {
	projectParser   *parser.ProjectParser
	projectEmbedder *embedder.ProjectEmbedder
	codeAssistant   *assistant.Assistant
	versionControl  vsc.VersionControlSystem
	consumerClint   kafka.Consumer
	workerCount     int32
}

func NewModule(projectParser *parser.ProjectParser, projectEmbedder *embedder.ProjectEmbedder, codeAssistant *assistant.Assistant, versionControl vsc.VersionControlSystem, consumerClint kafka.Consumer, workerCount int32) *Module {
	return &Module{
		projectParser:   projectParser,
		projectEmbedder: projectEmbedder,
		codeAssistant:   codeAssistant,
		versionControl:  versionControl,
		consumerClint:   consumerClint,
		workerCount:     workerCount,
	}
}

func (m *Module) Start() {
	logger := log.GetLogger()
	for i := 0; i < int(m.workerCount); i++ {
		go func() {
			for kafkaMessage := range m.consumerClint.Channel() {
				start := time.Now()
				var err error
				var event models.PullRequestEvent
				err = json.Unmarshal(kafkaMessage.Value, &event)
				if err != nil {
					logger.WithError(err).Error("failed to unmarshal event")
					continue
				}

				if err = m.process(&event); err == nil {
					if err := m.consumerClint.CommitMessage(kafkaMessage); err != nil {
						logger.WithError(err).Error("failed to commit message")
					}
				} else {
					logger.WithError(err).Warn("failed to process message")
				}

				// observe metrics
				go func() {
					status := metrics.Success
					if err != nil {
						status = metrics.Failure
					}
					metrics.Get().ObserveEventProcessing(status)
					metrics.Get().ObserveEventProcessingLatency(status, start)
				}()
			}
		}()
	}
}

func (m *Module) process(event *models.PullRequestEvent) error {
	logger := log.GetLogger()
	logger.Infof("processing pull request number = %v", event.Number)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	repoPath, cleanup, err := m.versionControl.Clone(ctx, event.CloneURL, event.Branch)
	if err != nil {
		logger.WithError(err).Error("failed to clone project")
		return err
	}
	defer cleanup()

	snippets, err := m.projectParser.ParseProject(ctx, repoPath)
	if err != nil {
		logger.WithError(err).Error("failed to parse project")
		return err
	}

	if len(snippets) == 0 {
		logger.Error("no snippets found")
		return errors.ErrNoSnippetFound
	}

	err = m.projectEmbedder.EmbedProject(ctx, snippets)
	if err != nil {
		logger.WithError(err).Error("Failed to embed project")
		return err
	}

	diff, err := m.versionControl.DownloadUrl(ctx, event.DiffURL)
	if err != nil {
		logger.WithError(err).Error("failed to download url")
		return err
	}

	review, err := m.codeAssistant.PerformTask(ctx, assistant.TaskCodeReview, diff)
	if err != nil {
		logger.WithError(err).Error("failed to perform coding task")
		return err
	}

	err = m.versionControl.PostPRComment(ctx, event.Number, review, event.Owner, event.Repo)
	if err != nil {
		logger.WithError(err).Error("failed to post comment")
		return err
	}

	return nil
}
