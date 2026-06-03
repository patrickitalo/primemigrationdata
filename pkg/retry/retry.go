package retry

import (
	"context"
	"fmt"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     float64 // Multiplicador para delay exponencial
}

type RetryableFunc func() error

func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       1 * time.Second,
		Backoff:     2.0,
	}
}

// Do executa uma função com retry automático
func Do(fn RetryableFunc, config RetryConfig) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < config.MaxAttempts {
			delay := time.Duration(float64(config.Delay) * float64(attempt) * config.Backoff)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("falha após %d tentativas. Último erro: %w", config.MaxAttempts, lastErr)
}

// DoWithContext executa uma função com retry e contexto para cancelamento
func DoWithContext(ctx context.Context, fn RetryableFunc, config RetryConfig) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operação cancelada: %w", ctx.Err())
		default:
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < config.MaxAttempts {
			delay := time.Duration(float64(config.Delay) * float64(attempt) * config.Backoff)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("operação cancelada durante retry: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("falha após %d tentativas. Último erro: %w", config.MaxAttempts, lastErr)
}
