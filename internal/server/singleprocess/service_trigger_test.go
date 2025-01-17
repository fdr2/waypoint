package singleprocess

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/waypoint/internal/server"
	pb "github.com/hashicorp/waypoint/internal/server/gen"
	serverptypes "github.com/hashicorp/waypoint/internal/server/ptypes"
)

func TestServiceTrigger(t *testing.T) {
	ctx := context.Background()

	// Create our server
	impl, err := New(WithDB(testDB(t)))
	require.NoError(t, err)
	client := server.TestServer(t, impl)

	type Req = pb.UpsertTriggerRequest

	t.Run("create and update", func(t *testing.T) {
		require := require.New(t)

		// Create, should get an ID back
		resp, err := client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: serverptypes.TestValidTrigger(t, nil),
		})
		require.NoError(err)
		require.NotNil(resp)
		result := resp.Trigger
		require.NotEmpty(result.Id)

		// Let's write some data
		testName := "TestyTest"
		result.Name = testName
		resp, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: result,
		})
		require.NoError(err)
		require.NotNil(resp)
		result = resp.Trigger
		require.Equal(result.Name, testName)
	})

	t.Run("create uses default workspace if unset", func(t *testing.T) {
		require := require.New(t)

		// Create, should get an ID back
		resp, err := client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: &pb.Trigger{Project: &pb.Ref_Project{Project: "test_proj"}},
		})
		require.NoError(err)
		require.NotNil(resp)
		result := resp.Trigger
		require.NotEmpty(result.Id)
		require.NotEmpty(result.Workspace)
		require.Equal(result.Workspace.Workspace, "default")
	})

	t.Run("errors if no project defined", func(t *testing.T) {
		require := require.New(t)

		// create should error with no project defined
		resp, err := client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: &pb.Trigger{},
		})
		require.Error(err)
		require.Nil(resp)
	})

	t.Run("update non-existent creates a new trigger", func(t *testing.T) {
		require := require.New(t)

		// Create, should get an ID back
		resp, err := client.UpsertTrigger(ctx, &Req{
			Trigger: serverptypes.TestValidTrigger(t, &pb.Trigger{
				Id: "nope",
			}),
		})
		require.NoError(err)
		require.NotNil(resp)
		require.Equal(resp.Trigger.Id, "nope")
	})
}

func TestServiceTrigger_GetTrigger(t *testing.T) {
	ctx := context.Background()

	// Create our server
	db := testDB(t)
	impl, err := New(WithDB(db))
	require.NoError(t, err)
	client := server.TestServer(t, impl)

	resp, err := client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	triggerId := resp.Trigger.Id

	type Req = pb.UpsertTriggerRequest

	t.Run("get existing", func(t *testing.T) {
		require := require.New(t)

		// Get, should return a trigger
		resp, err := client.GetTrigger(ctx, &pb.GetTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: resp.Trigger.Id,
			},
		})
		require.NoError(err)
		require.NotNil(resp.Trigger)
		require.NotEmpty(resp.Trigger.Id)
		require.Equal(triggerId, resp.Trigger.Id)
	})

	t.Run("get non-existing", func(t *testing.T) {
		require := require.New(t)

		// get, should fail
		resp, err := client.GetTrigger(ctx, &pb.GetTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: "nope",
			},
		})
		require.Error(err)
		require.Nil(resp)
		st, ok := status.FromError(err)
		require.True(ok)
		require.Equal(codes.NotFound, st.Code())
	})
}

func TestServiceTrigger_ListTriggersSimple(t *testing.T) {
	ctx := context.Background()

	// Create our server
	db := testDB(t)
	impl, err := New(WithDB(db))
	require.NoError(t, err)
	client := server.TestServer(t, impl)

	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})

	t.Run("list", func(t *testing.T) {
		require := require.New(t)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{})
		require.NoError(err)
		require.Equal(len(respList.Triggers), 3)
	})
}

func TestServiceTrigger_ListTriggersWithFilters(t *testing.T) {
	ctx := context.Background()

	// Create our server
	db := testDB(t)
	impl, err := New(WithDB(db))
	require.NoError(t, err)
	client := server.TestServer(t, impl)

	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})

	t.Run("list default workspace triggers", func(t *testing.T) {
		require := require.New(t)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace: &pb.Ref_Workspace{Workspace: "default"},
		})
		require.NoError(err)
		require.Equal(len(respList.Triggers), 3)
	})

	t.Run("list non-existent workspace triggers", func(t *testing.T) {
		require := require.New(t)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace: &pb.Ref_Workspace{Workspace: "fake"},
		})
		require.NoError(err)
		require.Equal(len(respList.Triggers), 0)
	})

	t.Run("list project triggers", func(t *testing.T) {
		require := require.New(t)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace: &pb.Ref_Workspace{Workspace: "default"},
			Project:   &pb.Ref_Project{Project: "p_test"},
		})
		require.NoError(err)
		require.Equal(len(respList.Triggers), 3)
	})

	t.Run("list app triggers", func(t *testing.T) {
		require := require.New(t)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace:   &pb.Ref_Workspace{Workspace: "default"},
			Project:     &pb.Ref_Project{Project: "p_test"},
			Application: &pb.Ref_Application{Project: "p_test", Application: "a_test"},
		})
		require.NoError(err)
		require.Equal(len(respList.Triggers), 3)
	})

	t.Run("filter to one app", func(t *testing.T) {
		require := require.New(t)

		_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: &pb.Trigger{
				Project: &pb.Ref_Project{
					Project: "secret_project",
				},
				Application: &pb.Ref_Application{
					Application: "another_one",
					Project:     "secret_project",
				},
				Workspace: &pb.Ref_Workspace{
					Workspace: "staging",
				},
			},
		})
		require.NoError(err)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace:   &pb.Ref_Workspace{Workspace: "staging"},
			Project:     &pb.Ref_Project{Project: "secret_project"},
			Application: &pb.Ref_Application{Project: "secret_project", Application: "another_one"},
		})
		require.NoError(err)
		require.Equal(1, len(respList.Triggers))
	})

	t.Run("filter on tags", func(t *testing.T) {
		require := require.New(t)

		_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: &pb.Trigger{
				Project: &pb.Ref_Project{
					Project: "secret_project",
				},
				Application: &pb.Ref_Application{
					Application: "another_one",
					Project:     "secret_project",
				},
				Workspace: &pb.Ref_Workspace{
					Workspace: "staging",
				},
				Tags: []string{"prod", "test"},
			},
		})
		require.NoError(err)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace: &pb.Ref_Workspace{Workspace: "staging"},
			Tags:      []string{"prod"},
		})
		require.NoError(err)
		require.Equal(1, len(respList.Triggers))
	})

	t.Run("filter on missing tags returns nothing", func(t *testing.T) {
		require := require.New(t)

		_, err = client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
			Trigger: &pb.Trigger{
				Project: &pb.Ref_Project{
					Project: "secret_project",
				},
				Application: &pb.Ref_Application{
					Application: "another_one",
					Project:     "secret_project",
				},
				Workspace: &pb.Ref_Workspace{
					Workspace: "staging",
				},
				Tags: []string{"prod", "test"},
			},
		})
		require.NoError(err)

		respList, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Workspace: &pb.Ref_Workspace{Workspace: "staging"},
			Tags:      []string{"pikachu"},
		})
		require.NoError(err)
		require.Equal(0, len(respList.Triggers))
	})

	t.Run("filter with missing workspace on app returns error", func(t *testing.T) {
		require := require.New(t)

		_, err := client.ListTriggers(ctx, &pb.ListTriggerRequest{
			Application: &pb.Ref_Application{Project: "secret_project", Application: "another_one"},
		})
		require.Error(err)
	})
}

func TestServiceTrigger_DeleteTrigger(t *testing.T) {
	ctx := context.Background()

	// Create our server
	db := testDB(t)
	impl, err := New(WithDB(db))
	require.NoError(t, err)
	client := server.TestServer(t, impl)

	resp, err := client.UpsertTrigger(ctx, &pb.UpsertTriggerRequest{
		Trigger: serverptypes.TestValidTrigger(t, nil),
	})
	triggerId := resp.Trigger.Id

	type Req = pb.UpsertTriggerRequest

	t.Run("get existing then delete", func(t *testing.T) {
		require := require.New(t)

		// Get, should return a trigger
		resp, err := client.GetTrigger(ctx, &pb.GetTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: resp.Trigger.Id,
			},
		})
		require.NoError(err)
		require.NotNil(resp.Trigger)
		require.NotEmpty(resp.Trigger.Id)
		require.Equal(triggerId, resp.Trigger.Id)

		_, err = client.DeleteTrigger(ctx, &pb.DeleteTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: triggerId,
			},
		})
		require.NoError(err)

		// get, should fail
		resp, err = client.GetTrigger(ctx, &pb.GetTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: triggerId,
			},
		})
		require.Error(err)
		require.Nil(resp)
		st, ok := status.FromError(err)
		require.True(ok)
		require.Equal(codes.NotFound, st.Code())
	})

	t.Run("delete non-existing", func(t *testing.T) {
		require := require.New(t)

		resp, err := client.DeleteTrigger(ctx, &pb.DeleteTriggerRequest{
			Ref: &pb.Ref_Trigger{
				Id: "nope",
			},
		})
		require.NoError(err)
		require.NotNil(resp)
	})
}
