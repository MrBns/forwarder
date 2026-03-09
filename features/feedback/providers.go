package feedback

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is the Wire provider set for the feedback feature.
// Wire.Bind maps each port interface to its concrete implementation so that
// dependents receive the interface, keeping the hexagonal boundaries clean.
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(Repository), new(*PostgresRepository)),
	NewService,
	wire.Bind(new(Service), new(*FeedbackService)),
	NewHandler,
)

// Ensure *pgxpool.Pool satisfies the type Wire needs to resolve NewRepository.
// This blank import keeps the dependency explicit in the provider set file.
var _ *pgxpool.Pool
