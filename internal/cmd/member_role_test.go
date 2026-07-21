package cmd

import (
	"errors"
	"net/http"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestMemberRoleCmd(t *testing.T) {
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
	setupUpdate := func(role api.MemberRole) func(m *mock.MockClientWithResponsesInterface) {
		return func(m *mock.MockClientWithResponsesInterface) {
			updated := api.Member{UserID: 102, Name: "Bob Jones", Email: "bob@example.com", Role: role}
			m.EXPECT().UpdateMemberRoleWithResponse(validCtx, "test-space", int64(102), api.UpdateMemberRoleRequest{Role: role}).
				Return(&api.UpdateMemberRoleResponse{
					HTTPResponse: httpResponse(http.StatusOK),
					JSON200:      &updated,
				}, nil)
		}
	}

	tests := []cmdTest{
		{
			name:    "invalid role",
			args:    []string{"member", "role", "bob@example.com", "bogus"},
			opts:    []runOption{experimental},
			wantErr: "invalid role 'bogus'; must be one of admin, developer, or viewer",
		},
		{
			name:    "owner role",
			args:    []string{"member", "role", "bob@example.com", "owner"},
			opts:    []runOption{experimental},
			wantErr: "the owner role cannot be granted; every space has exactly one owner",
		},
		{
			name:    "not logged in",
			args:    []string{"member", "role", "bob@example.com", "admin"},
			opts:    []runOption{experimental, withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name: "network error on list",
			args: []string{"member", "role", "bob@example.com", "admin"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListMembersWithResponse(validCtx, "test-space").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to list members: connection refused",
		},
		{
			name: "API error on list",
			args: []string{"member", "role", "bob@example.com", "admin"},
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
			args: []string{"member", "role", "bob@example.com", "admin"},
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
			args:    []string{"member", "role", "carol@example.com", "admin"},
			opts:    []runOption{experimental},
			setup:   setupList,
			wantErr: "no member with email 'carol@example.com' found; run 'ghost member list' to see the members of this space",
		},
		{
			name: "network error on update",
			args: []string{"member", "role", "bob@example.com", "admin"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().UpdateMemberRoleWithResponse(validCtx, "test-space", int64(102), api.UpdateMemberRoleRequest{Role: api.MemberRoleAdmin}).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to update member role: connection refused",
		},
		{
			name: "API error on update",
			args: []string{"member", "role", "bob@example.com", "developer"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().UpdateMemberRoleWithResponse(validCtx, "test-space", int64(102), api.UpdateMemberRoleRequest{Role: api.MemberRoleDeveloper}).
					Return(&api.UpdateMemberRoleResponse{
						HTTPResponse: httpResponse(http.StatusConflict),
						JSONDefault:  &api.Error{Message: "user already has the developer role"},
					}, nil)
			},
			wantErr: "user already has the developer role",
		},
		{
			name: "nil response body on update",
			args: []string{"member", "role", "bob@example.com", "admin"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				m.EXPECT().UpdateMemberRoleWithResponse(validCtx, "test-space", int64(102), api.UpdateMemberRoleRequest{Role: api.MemberRoleAdmin}).
					Return(&api.UpdateMemberRoleResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      nil,
					}, nil)
			},
			wantErr: "empty response from API",
		},
		{
			name: "success",
			args: []string{"member", "role", "bob@example.com", "admin"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupUpdate(api.MemberRoleAdmin)(m)
			},
			wantStdout: "Changed role of Bob Jones (bob@example.com) to admin\n",
		},
		{
			name: "role argument is case-insensitive",
			args: []string{"member", "role", "bob@example.com", "VIEWER"},
			opts: []runOption{experimental},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				setupList(m)
				setupUpdate(api.MemberRoleViewer)(m)
			},
			wantStdout: "Changed role of Bob Jones (bob@example.com) to viewer\n",
		},
	}

	runCmdTests(t, tests)
}
