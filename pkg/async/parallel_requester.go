package async

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// NodeRequest represents a request to a blockchain node
type NodeRequest struct {
	NodeURL string
	Method  string
	Params  []byte
}

// NodeResponse represents a response from a blockchain node
type NodeResponse struct {
	Data         []byte
	ResponseTime int64
	NodeURL      string
	Error        error
}

// RequestFunc defines the function signature for making requests
type RequestFunc func(ctx context.Context, nodeURL string, method string, params []byte) ([]byte, int64, error)

// ParallelRequester makes parallel requests to multiple nodes with failover
type ParallelRequester struct {
	requestFunc RequestFunc
	timeout     time.Duration
}

// NewParallelRequester creates a new parallel requester
func NewParallelRequester(requestFunc RequestFunc, timeout time.Duration) *ParallelRequester {
	return &ParallelRequester{
		requestFunc: requestFunc,
		timeout:     timeout,
	}
}

// RequestWithFailover makes parallel requests to multiple nodes and returns the first successful response
func (pr *ParallelRequester) RequestWithFailover(ctx context.Context, nodeURLs []string, method string, params []byte) (*NodeResponse, error) {
	if len(nodeURLs) == 0 {
		return nil, errors.New("no node URLs provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, pr.timeout)
	defer cancel()

	// Channel to receive responses
	responseChan := make(chan *NodeResponse, len(nodeURLs))

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each node
	for _, nodeURL := range nodeURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Make request
			data, responseTime, err := pr.requestFunc(ctx, url, method, params)

			// Send response
			select {
			case responseChan <- &NodeResponse{
				Data:         data,
				ResponseTime: responseTime,
				NodeURL:      url,
				Error:        err,
			}:
			case <-ctx.Done():
				return
			}
		}(nodeURL)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect responses
	var successResponse *NodeResponse
	var errors []error

	for response := range responseChan {
		if response.Error == nil && successResponse == nil {
			// First successful response
			successResponse = response
			cancel() // Cancel remaining requests
			return successResponse, nil
		}

		if response.Error != nil {
			errors = append(errors, fmt.Errorf("node %s: %w", response.NodeURL, response.Error))
		}
	}

	// If no successful response, return errors
	if successResponse == nil {
		return nil, fmt.Errorf("all nodes failed: %v", errors)
	}

	return successResponse, nil
}

// RequestAll makes parallel requests to all nodes and returns all responses
func (pr *ParallelRequester) RequestAll(ctx context.Context, nodeURLs []string, method string, params []byte) ([]*NodeResponse, error) {
	if len(nodeURLs) == 0 {
		return nil, errors.New("no node URLs provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, pr.timeout)
	defer cancel()

	// Channel to receive responses
	responseChan := make(chan *NodeResponse, len(nodeURLs))

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each node
	for _, nodeURL := range nodeURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Make request
			data, responseTime, err := pr.requestFunc(ctx, url, method, params)

			// Send response
			select {
			case responseChan <- &NodeResponse{
				Data:         data,
				ResponseTime: responseTime,
				NodeURL:      url,
				Error:        err,
			}:
			case <-ctx.Done():
				return
			}
		}(nodeURL)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect all responses
	var responses []*NodeResponse
	for response := range responseChan {
		responses = append(responses, response)
	}

	return responses, nil
}

// RequestFastest makes parallel requests to all nodes and returns the fastest successful response
func (pr *ParallelRequester) RequestFastest(ctx context.Context, nodeURLs []string, method string, params []byte) (*NodeResponse, error) {
	if len(nodeURLs) == 0 {
		return nil, errors.New("no node URLs provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, pr.timeout)
	defer cancel()

	// Channel to receive first successful response
	successChan := make(chan *NodeResponse, 1)
	errorChan := make(chan error, len(nodeURLs))

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each node
	for _, nodeURL := range nodeURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Make request
			data, responseTime, err := pr.requestFunc(ctx, url, method, params)

			if err != nil {
				select {
				case errorChan <- fmt.Errorf("node %s: %w", url, err):
				case <-ctx.Done():
				}
				return
			}

			// Try to send successful response (only first one will succeed)
			select {
			case successChan <- &NodeResponse{
				Data:         data,
				ResponseTime: responseTime,
				NodeURL:      url,
				Error:        nil,
			}:
				cancel() // Cancel other requests
			case <-ctx.Done():
			}
		}(nodeURL)
	}

	// Wait for result or all failures
	select {
	case response := <-successChan:
		return response, nil
	case <-ctx.Done():
		// Collect errors
		close(errorChan)
		var errors []error
		for err := range errorChan {
			errors = append(errors, err)
		}
		if len(errors) > 0 {
			return nil, fmt.Errorf("all requests failed or timed out: %v", errors)
		}
		return nil, ctx.Err()
	}
}

// RequestWithRetry makes a request with automatic retry on failure
func (pr *ParallelRequester) RequestWithRetry(ctx context.Context, nodeURL string, method string, params []byte, maxRetries int) (*NodeResponse, error) {
	var lastError error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Make request
		data, responseTime, err := pr.requestFunc(ctx, nodeURL, method, params)

		if err == nil {
			return &NodeResponse{
				Data:         data,
				ResponseTime: responseTime,
				NodeURL:      nodeURL,
				Error:        nil,
			}, nil
		}

		lastError = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			break
		}

		// Wait before retry with exponential backoff
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastError)
}

// BatchRequest makes parallel requests for multiple methods
func (pr *ParallelRequester) BatchRequest(ctx context.Context, nodeURL string, requests []NodeRequest) ([]*NodeResponse, error) {
	if len(requests) == 0 {
		return nil, errors.New("no requests provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, pr.timeout)
	defer cancel()

	// Channel to receive responses
	responseChan := make(chan *NodeResponse, len(requests))

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Launch goroutines for each request
	for i, req := range requests {
		wg.Add(1)
		go func(index int, request NodeRequest) {
			defer wg.Done()

			// Make request
			data, responseTime, err := pr.requestFunc(ctx, request.NodeURL, request.Method, request.Params)

			// Send response with index to preserve order
			select {
			case responseChan <- &NodeResponse{
				Data:         data,
				ResponseTime: responseTime,
				NodeURL:      request.NodeURL,
				Error:        err,
			}:
			case <-ctx.Done():
				return
			}
		}(i, req)
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect all responses
	var responses []*NodeResponse
	for response := range responseChan {
		responses = append(responses, response)
	}

	return responses, nil
}
