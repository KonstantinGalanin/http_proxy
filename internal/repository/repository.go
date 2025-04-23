package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/KonstantinGalanin/http_proxy/internal/proxy"
)

var (
	InsertRequest     = "INSERT INTO requests (scheme, method, path, host, get_params, headers, cookies, post_params, body) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	SelectRequests    = "SELECT scheme, method, path, host, get_params, headers, cookies, post_params, body FROM requests"
	SelectRequestByID = "SELECT scheme, method, path, host, get_params, headers, cookies, post_params, body FROM requests WHERE id = $1"
	InsertResponse    = "INSERT INTO responses (status_code, status, headers, body, content_length) VALUES ($1, $2, $3, $4, $5)"
)

type PostgresRepo struct {
	DB *sql.DB
}

func New(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

func (p *PostgresRepo) SaveRequest(req *proxy.ParsedRequest) error {
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	cookiesJSON, err := json.Marshal(req.Cookies)
	if err != nil {
		return fmt.Errorf("failed to marshal cookies: %w", err)
	}

	getParamsJSON, err := json.Marshal(req.GetParams)
	if err != nil {
		return fmt.Errorf("failed to marshal get_params: %w", err)
	}

	postParamsJSON, err := json.Marshal(req.PostParams)
	if err != nil {
		return fmt.Errorf("failed to marshal post_params: %w", err)
	}

	fmt.Println(req.Scheme)

	_, err = p.DB.Exec(InsertRequest,
		req.Scheme,
		req.Method,
		req.Path,
		req.Host,
		getParamsJSON,
		headersJSON,
		cookiesJSON,
		postParamsJSON,
		req.Body,
	)
	if err != nil {
		return fmt.Errorf("failed to insert request: %w", err)
	}

	return nil
}

func (p *PostgresRepo) GetRequests() ([]*proxy.ParsedRequest, error) {
	rows, err := p.DB.Query(SelectRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []*proxy.ParsedRequest

	for rows.Next() {
		var (
			scheme, method, path, host, body                        string
			headersJSON, cookiesJSON, getParamsJSON, postParamsJSON []byte
		)

		err := rows.Scan(
			&scheme,
			&method,
			&path,
			&host,
			&getParamsJSON,
			&headersJSON,
			&cookiesJSON,
			&postParamsJSON,
			&body,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request row: %w", err)
		}

		var headers map[string][]string
		if err := json.Unmarshal(headersJSON, &headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}

		var cookies map[string]string
		if err := json.Unmarshal(cookiesJSON, &cookies); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
		}

		var getParams map[string][]string
		if err := json.Unmarshal(getParamsJSON, &getParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal get_params: %w", err)
		}

		var postParams map[string][]string
		if err := json.Unmarshal(postParamsJSON, &postParams); err != nil {
			return nil, fmt.Errorf("failed to unmarshal post_params: %w", err)
		}

		requests = append(requests, &proxy.ParsedRequest{
			Scheme:     scheme,
			Method:     method,
			Path:       path,
			Host:       host,
			Headers:    headers,
			Cookies:    cookies,
			GetParams:  getParams,
			PostParams: postParams,
			Body:       body,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return requests, nil
}

func (p *PostgresRepo) GetRequestByID(id int) (*proxy.ParsedRequest, error) {
	var (
		scheme, method, path, host, body                        string
		headersJSON, cookiesJSON, getParamsJSON, postParamsJSON []byte
	)

	err := p.DB.QueryRow(SelectRequestByID, id).Scan(
		&scheme,
		&method,
		&path,
		&host,
		&getParamsJSON,
		&headersJSON,
		&cookiesJSON,
		&postParamsJSON,
		&body,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get request by ID: %w", err)
	}

	var headers map[string][]string
	if err := json.Unmarshal(headersJSON, &headers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
	}

	var cookies map[string]string
	if err := json.Unmarshal(cookiesJSON, &cookies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	var getParams map[string][]string
	if err := json.Unmarshal(getParamsJSON, &getParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get_params: %w", err)
	}

	var postParams map[string][]string
	if err := json.Unmarshal(postParamsJSON, &postParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post_params: %w", err)
	}

	return &proxy.ParsedRequest{
		Scheme:     scheme,
		Method:     method,
		Path:       path,
		Host:       host,
		Headers:    headers,
		Cookies:    cookies,
		GetParams:  getParams,
		PostParams: postParams,
		Body:       body,
	}, nil
}

func (p *PostgresRepo) SaveResponse(response *proxy.ParsedResponse) error {
	headers, err := json.Marshal(response.Headers)
	if err != nil {
		return err
	}

	_, err = p.DB.Exec(
		InsertResponse,
		response.StatusCode,
		response.Status,
		headers,
		response.Body,
		response.ContentLength,
	)
	return err
}
