package autonomera

import (
	"context"
	"core/internal/domain"
	"core/pkg/integration/autonomera"
	"fmt"
	"log/slog"
)

type Provider struct {
	client    *autonomera.Client
	parser    *Parser
	logger    *slog.Logger
	batchSize int
	startPos  int
}

type Option func(*Provider)

func WithBatchSize(size int) Option {
	return func(p *Provider) {
		p.batchSize = size
	}
}

func WithStartPosition(pos int) Option {
	return func(p *Provider) {
		p.startPos = pos
	}
}

func NewProvider(client *autonomera.Client, parser *Parser, logger *slog.Logger, opts ...Option) *Provider {
	p := &Provider{
		client:    client,
		parser:    parser,
		logger:    logger,
		batchSize: 20,
		startPos:  0,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type OffsetIterator struct {
	client    *autonomera.Client
	parser    *Parser
	logger    *slog.Logger
	batchSize int
	current   int
	hasMore   bool
}

func (p *Provider) FetchNumbers(ctx context.Context) (domain.NumberIterator, error) {
	return &OffsetIterator{
		client:    p.client,
		parser:    p.parser,
		logger:    p.logger,
		batchSize: p.batchSize,
		current:   p.startPos,
		hasMore:   true,
	}, nil
}

func (it *OffsetIterator) Next(ctx context.Context) ([]domain.CarNumber, error) {
	it.logger.Debug("fetching batch",
		"offset", it.current,
		"batch_size", it.batchSize)

	html, err := it.client.FetchNumbersHTML(ctx, it.current)
	if err != nil {
		return nil, fmt.Errorf("fetch html at offset %d: %w", it.current, err)
	}

	numbers, err := it.parser.Parse(html)
	if err != nil {
		return nil, fmt.Errorf("parse html at offset %d: %w", it.current, err)
	}

	it.current += it.batchSize
	it.hasMore = len(numbers) > 0

	return numbers, nil
}

func (it *OffsetIterator) HasNext() bool {
	return it.hasMore
}
