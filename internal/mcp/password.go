package mcp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// PasswordInput represents input for ghost_password
type PasswordInput struct {
	ID       string `json:"id"`
	Password string `json:"password,omitempty"`
}

func (PasswordInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[PasswordInput](nil))
	databaseRefInputProperties(schema)
	schema.Properties["password"].Description = "The new password. If not provided, a secure password will be automatically generated"
	return schema
}

// PasswordOutput represents output for ghost_password
type PasswordOutput struct {
	ID               string   `json:"id"`
	Password         string   `json:"password"`
	ConnectionString string   `json:"connection_string"`
	Warnings         []string `json:"warnings,omitempty"`
}

func (PasswordOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[PasswordOutput](nil))
	databaseIDOutputProperties(schema)
	schema.Properties["password"].Description = "The new password"
	connectionStringOutputProperties(schema)
	warningsOutputProperties(schema)
	return schema
}

func newPasswordTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_password",
		Title:        "Update Database Password",
		Description:  "Reset the password for a database's default user (tsdbadmin). This changes the password on the server itself. This operation is irreversible. Existing connections using the old password will fail to reconnect.",
		InputSchema:  PasswordInput{}.Schema(),
		OutputSchema: PasswordOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(true),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Update Database Password",
		},
	}
}

func (s *Server) handlePassword(ctx context.Context, req *mcp.CallToolRequest, input PasswordInput) (*mcp.CallToolResult, PasswordOutput, error) {
	cfg, client, projectID, err := s.app.GetAll()
	if err != nil {
		return nil, PasswordOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, PasswordOutput{}, err
	}

	// Fetch database details
	getResp, err := client.GetDatabaseWithResponse(ctx, projectID, input.ID)
	if err != nil {
		return nil, PasswordOutput{}, fmt.Errorf("failed to get database details: %w", err)
	}

	if getResp.StatusCode() != http.StatusOK {
		return nil, PasswordOutput{}, common.ExitWithErrorFromStatusCode(getResp.StatusCode(), getResp.JSONDefault)
	}

	if getResp.JSON200 == nil {
		return nil, PasswordOutput{}, errors.New("empty response from API")
	}
	database := *getResp.JSON200

	// Check if the database is ready
	if err := common.CheckReady(database); err != nil {
		return nil, PasswordOutput{}, handleDatabaseError(err)
	}

	// Determine the password to use (generate if not provided)
	password := input.Password
	if password == "" {
		password, err = generateSecurePassword(32)
		if err != nil {
			return nil, PasswordOutput{}, fmt.Errorf("failed to generate password: %w", err)
		}
	}

	// Update the password via API
	updateReq := api.UpdatePasswordRequest{Password: password}
	resp, err := client.UpdatePasswordWithResponse(ctx, projectID, database.Id, updateReq)
	if err != nil {
		return nil, PasswordOutput{}, fmt.Errorf("failed to update password: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return nil, PasswordOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	// Save password to .pgpass
	var warnings []string
	if err := common.SavePassword(database, password, "tsdbadmin"); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to save password to .pgpass: %v", err))
	}

	// Build connection string
	connStr, err := common.BuildConnectionString(common.ConnectionStringArgs{
		Database: database,
		Role:     "tsdbadmin",
		Password: password,
	})
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to build connection string: %v", err))
	}

	return nil, PasswordOutput{
		ID:               database.Id,
		Password:         password,
		ConnectionString: connStr,
		Warnings:         warnings,
	}, nil
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	// Encode as URL-safe base64 to avoid special characters
	encoded := base64.URLEncoding.EncodeToString(bytes)

	// Trim to desired length
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}
