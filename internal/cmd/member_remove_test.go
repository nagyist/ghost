package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestMemberRemoveCmd(t *testing.T) {
	experimental := withEnv("GHOST_EXPERIMENTAL", "true")

	setupList := func(m *mock.MockClientWithResponsesInterface) {
		members := []api.Member{
			{UserID: 101, Name: "Alice Smith", Email: "alice@example.com", Role: api.MemberRoleOwner},
			{UserID: 102, Name: "Bob Jones", Email: "bob@example.com", Role: api.MemberRoleDeveloper},
		}
		m.EXPECT().ListMembersWithResponse(validCtx, "test-space").
			Return(&api.ListMembersResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &members,
			}, nil)
	}
	setupRemove := func(m *mock.MockClientWithResponsesInterface) {
		removed := api.Member{UserID: 102, Name: "Bob Jones", Email: "bob@example.com", Role: api.MemberRoleDeveloper}
		m.EXPECT().RemoveMemberWithResponse(validCtx, "test-space", int64(102)).
			Return(&api.RemoveMemberResponse{
				HTTPResponse: httpResponse(http.StatusOK),
				JSON200:      &removed,
			}, nil)
	}

	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"member", "remove", "bob@example.com"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on list",
			args: []string{"member", "remove", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-space").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list members: connection refused",
		},
		{
			name: "API error on list",
			args: []string{"member", "remove", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-space").
					Return(&api.ListMembersResponse{
						HTTPResponse: httpResponse(http.StatusForbidden),
						JSONDefault:  &api.Error{Message: "this endpoint requires user authentication"},
					}, nil)
			},
			wantErr: "this endpoint requires user authentication",
		},
		{
			name: "nil response body on list",
			args: []string{"member", "remove", "bob@example.com"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-space").
					Return(&api.ListMembersResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name:    "member not found",
			args:    []string{"member", "remove", "carol@example.com"},
			opts:    []runOption{experimental},
			setup:   setupList,
			wantErr: "no member with email 'carol@example.com' found; run 'ghost member list' to see the members of this space",
		},
		{
			name:    "non-terminal stdin without confirm flag",
			args:    []string{"member", "remove", "bob@example.com"},
			opts:    []runOption{experimental},
			setup:   setupList,
			wantErr: "cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip",
		},
		{
			name:       "confirmation declined",
			args:       []string{"member", "remove", "bob@example.com"},
			opts:       []runOption{experimental, withStdin("n\n"), withIsTerminal(true)},
			setup:      setupList,
			wantStderr: "Remove Bob Jones (bob@example.com) from the space? [y/N] ",
			wantStdout: "Remove operation cancelled.\n",
		},
		{
			name: "network error on remove",
			args: []string{"member", "remove", "bob@example.com", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().RemoveMemberWithResponse(validCtx, "test-space", int64(102)).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to remove member: connection refused",
		},
		{
			name: "API error on remove",
			args: []string{"member", "remove", "alice@example.com", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().RemoveMemberWithResponse(validCtx, "test-space", int64(101)).
					Return(&api.RemoveMemberResponse{
						HTTPResponse: httpResponse(http.StatusBadRequest),
						JSONDefault:  &api.Error{Message: "the space owner cannot be removed"},
					}, nil)
			},
			wantErr: "the space owner cannot be removed",
		},
		{
			name: "nil response body on remove",
			args: []string{"member", "remove", "bob@example.com", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().RemoveMemberWithResponse(validCtx, "test-space", int64(102)).
					Return(&api.RemoveMemberResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "confirmation accepted",
			args: []string{"member", "remove", "bob@example.com"},
			opts: []runOption{experimental, withStdin("y\n"), withIsTerminal(true)},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupRemove(m)
			},
			wantStderr: "Remove Bob Jones (bob@example.com) from the space? [y/N] ",
			wantStdout: "Removed Bob Jones (bob@example.com) from the space\n",
		},
		{
			name: "confirm flag",
			args: []string{"member", "remove", "bob@example.com", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupRemove(m)
			},
			wantStdout: "Removed Bob Jones (bob@example.com) from the space\n",
		},
		{
			name: "rm alias",
			args: []string{"member", "rm", "bob@example.com", "--confirm"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupRemove(m)
			},
			wantStdout: "Removed Bob Jones (bob@example.com) from the space\n",
		},
	}

	runCmdTests(t, tests)
}
