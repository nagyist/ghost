package cmd

import (
	"testing"
)

func TestMCPGetCmd(t *testing.T) {
	ghostListText := `List Databases [read-only] [open-world]

Tool name: ghost_list

Description:
List all databases.

Output:
  • databases (required): []object, null
    • id (required): string - Database identifier
    • name (required): string - Database name
    • size: string, null - Compute size for dedicated databases
    • status (required): string - Database status
    • storage (required): string - Current storage usage
    • type (required): string - Database type

`

	ghostListJSON := `{
  "annotations": {
    "openWorldHint": true,
    "readOnlyHint": true,
    "title": "List Databases"
  },
  "description": "List all databases.",
  "inputSchema": {
    "additionalProperties": false,
    "type": "object"
  },
  "name": "ghost_list",
  "outputSchema": {
    "additionalProperties": false,
    "properties": {
      "databases": {
        "items": {
          "additionalProperties": false,
          "properties": {
            "id": {
              "description": "Database identifier",
              "type": "string"
            },
            "name": {
              "description": "Database name",
              "type": "string"
            },
            "size": {
              "description": "Compute size for dedicated databases",
              "type": [
                "null",
                "string"
              ]
            },
            "status": {
              "description": "Database status",
              "type": "string"
            },
            "storage": {
              "description": "Current storage usage",
              "examples": [
                "512MiB",
                "1GiB"
              ],
              "type": "string"
            },
            "type": {
              "description": "Database type",
              "type": "string"
            }
          },
          "required": [
            "id",
            "name",
            "type",
            "status",
            "storage"
          ],
          "type": "object"
        },
        "type": [
          "null",
          "array"
        ]
      }
    },
    "required": [
      "databases"
    ],
    "type": "object"
  },
  "title": "List Databases"
}
`

	ghostListYAML := `annotations:
  openWorldHint: true
  readOnlyHint: true
  title: List Databases
description: List all databases.
inputSchema:
  additionalProperties: false
  type: object
name: ghost_list
outputSchema:
  additionalProperties: false
  properties:
    databases:
      items:
        additionalProperties: false
        properties:
          id:
            description: Database identifier
            type: string
          name:
            description: Database name
            type: string
          size:
            description: Compute size for dedicated databases
            type:
              - "null"
              - string
          status:
            description: Database status
            type: string
          storage:
            description: Current storage usage
            examples:
              - 512MiB
              - 1GiB
            type: string
          type:
            description: Database type
            type: string
        required:
          - id
          - name
          - type
          - status
          - storage
        type: object
      type:
        - "null"
        - array
  required:
    - databases
  type: object
title: List Databases
`

	tests := []cmdTest{
		{
			name:    "not found",
			args:    []string{"mcp", "get", "nonexistent"},
			wantErr: `capability "nonexistent" not found`,
		},
		{
			name:       "text output",
			args:       []string{"mcp", "get", "ghost_list"},
			wantStdout: ghostListText,
		},
		{
			name:       "json output",
			args:       []string{"mcp", "get", "ghost_list", "--json"},
			wantStdout: ghostListJSON,
		},
		{
			name:       "yaml output",
			args:       []string{"mcp", "get", "ghost_list", "--yaml"},
			wantStdout: ghostListYAML,
		},
		{
			name:       "describe alias",
			args:       []string{"mcp", "describe", "ghost_list"},
			wantStdout: ghostListText,
		},
	}

	runCmdTests(t, tests)
}
