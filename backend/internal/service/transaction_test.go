package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/arkhe-systems/senddock/internal/db"
)

// These tests exercise the flows that write more than one row to mean one thing, against a
// real database. They skip unless DATABASE_URL points at a migrated instance.
//
// The failure they look for is the one nobody sees: a first row committed and a second one
// lost. The only way to reproduce it deterministically is to make the second write fail on
// purpose, which a trigger does — a unique key cannot be made to collide on demand without
// taking the decision away from the code under test.

type allowAllGate struct{}

func (allowAllGate) AllowsFeature(context.Context, string) bool { return true }

func transactionTestDB(t *testing.T) (*sql.DB, *db.Queries) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping transaction integration test")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		t.Fatalf("ping database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return conn, db.New(conn)
}

// failInsertsInto makes every insert into table fail, so the second write of a flow can be
// told to fail while the first one is already done.
func failInsertsInto(t *testing.T, conn *sql.DB, table string) {
	t.Helper()

	const trigger = "senddock_test_fail_insert"
	drop := func() {
		_, _ = conn.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS `+trigger+` ON `+table)
		_, _ = conn.ExecContext(context.Background(), `DROP FUNCTION IF EXISTS `+trigger+`()`)
	}
	// A previous run may have died before its own cleanup ran.
	drop()

	if _, err := conn.ExecContext(context.Background(),
		`CREATE FUNCTION `+trigger+`() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'forced failure for test'; END; $$ LANGUAGE plpgsql`,
	); err != nil {
		t.Fatalf("create trigger function: %v", err)
	}
	if _, err := conn.ExecContext(context.Background(),
		`CREATE TRIGGER `+trigger+` BEFORE INSERT ON `+table+` FOR EACH ROW EXECUTE FUNCTION `+trigger+`()`,
	); err != nil {
		drop()
		t.Fatalf("create trigger: %v", err)
	}
	t.Cleanup(drop)
}

// workspaceFixture creates an owner and a workspace they own, which is what the flows that
// manage members require before they will write anything.
func workspaceFixture(t *testing.T, conn *sql.DB) (workspaceID, ownerID uuid.UUID, ownerEmail string) {
	t.Helper()

	ctx := context.Background()
	ownerEmail = "tx-owner-" + uuid.NewString() + "@example.test"

	if err := conn.QueryRowContext(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`,
		ownerEmail, "Transaction Owner",
	).Scan(&ownerID); err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	if err := conn.QueryRowContext(ctx,
		`INSERT INTO workspaces (name, created_by) VALUES ($1, $2) RETURNING id`,
		"tx-test", ownerID,
	).Scan(&workspaceID); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.ExecContext(ctx, `DELETE FROM workspaces WHERE id = $1`, workspaceID); err != nil {
			t.Logf("cleanup workspace: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, ownerID); err != nil {
			t.Logf("cleanup owner: %v", err)
		}
	})
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`,
		workspaceID, ownerID,
	); err != nil {
		t.Fatalf("insert owner membership: %v", err)
	}

	return workspaceID, ownerID, ownerEmail
}

// deleteUserByEmail removes a row the test may have left behind, so a failing run does not
// poison the next one.
func deleteUserByEmail(t *testing.T, conn *sql.DB, email string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = conn.ExecContext(context.Background(), `DELETE FROM users WHERE email = $1`, email)
	})
}

func TestRegisterCreatesAccountWorkspaceAndMembership(t *testing.T) {
	conn, queries := transactionTestDB(t)
	auth := NewAuthService(queries, conn, "test-secret-that-is-long-enough-for-tests")

	email := "register-" + uuid.NewString() + "@example.test"
	deleteUserByEmail(t, conn, email)

	tokens, err := auth.Register(context.Background(), email, "Correct-Horse-Battery-1", "Grace Hopper")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Register returned incomplete tokens: %+v", tokens)
	}

	var userID uuid.UUID
	if err := conn.QueryRowContext(context.Background(),
		`SELECT id FROM users WHERE email = $1`, email).Scan(&userID); err != nil {
		t.Fatalf("registered user is missing: %v", err)
	}

	var memberships int
	if err := conn.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workspace_members m
		   JOIN workspaces w ON w.id = m.workspace_id
		  WHERE m.user_id = $1 AND m.role = 'owner' AND w.created_by = $1`, userID).Scan(&memberships); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if memberships != 1 {
		t.Errorf("registration produced %d owned workspaces, want 1", memberships)
	}
}

// The account is written first and the workspace second, so a failure on the workspace used
// to leave an account behind holding an email that can never be registered again.
func TestRegisterLeavesNoAccountWhenTheWorkspaceWriteFails(t *testing.T) {
	conn, queries := transactionTestDB(t)
	auth := NewAuthService(queries, conn, "test-secret-that-is-long-enough-for-tests")

	email := "register-orphan-" + uuid.NewString() + "@example.test"
	deleteUserByEmail(t, conn, email)

	failInsertsInto(t, conn, "workspaces")

	if _, err := auth.Register(context.Background(), email, "Correct-Horse-Battery-1", "Ada Lovelace"); err == nil {
		t.Fatal("Register reported success although the workspace could not be created")
	}

	if _, err := queries.GetUserByEmail(context.Background(), email); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("account survived a failed registration: err = %v", err)
	}
}

// The acceptance criterion of the audit item: a failing membership insert must not leave the
// account it was created for.
func TestCreateUserAndAddMemberLeavesNoOrphanWhenTheMembershipWriteFails(t *testing.T) {
	conn, queries := transactionTestDB(t)
	workspaces := NewWorkspaceService(queries, conn)
	workspaces.SetLicenseGate(allowAllGate{})

	workspaceID, ownerID, _ := workspaceFixture(t, conn)

	email := "orphan-" + uuid.NewString() + "@example.test"
	deleteUserByEmail(t, conn, email)

	failInsertsInto(t, conn, "workspace_members")

	if _, err := workspaces.CreateUserAndAddMember(
		context.Background(), workspaceID, ownerID, email, "Alan Turing", "Correct-Horse-Battery-1", "developer",
	); err == nil {
		t.Fatal("CreateUserAndAddMember reported success although the membership could not be written")
	}

	if _, err := queries.GetUserByEmail(context.Background(), email); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("account survived a failed membership insert: err = %v", err)
	}
}

func TestCreateUserAndAddMemberWritesBothRows(t *testing.T) {
	conn, queries := transactionTestDB(t)
	workspaces := NewWorkspaceService(queries, conn)
	workspaces.SetLicenseGate(allowAllGate{})

	workspaceID, ownerID, _ := workspaceFixture(t, conn)

	email := "member-" + uuid.NewString() + "@example.test"
	deleteUserByEmail(t, conn, email)

	created, err := workspaces.CreateUserAndAddMember(
		context.Background(), workspaceID, ownerID, email, "Ada Lovelace", "Correct-Horse-Battery-1", "developer",
	)
	if err != nil {
		t.Fatalf("CreateUserAndAddMember: %v", err)
	}

	var role string
	if err := conn.QueryRowContext(context.Background(),
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, created.UserID).Scan(&role); err != nil {
		t.Fatalf("membership row is missing: %v", err)
	}
	if role != "developer" {
		t.Errorf("role = %q, want developer", role)
	}
}

// A workspace whose owner row never landed belongs to nobody: it is not in any list and no
// screen can delete it.
func TestCreateWorkspaceLeavesNothingBehindWhenTheMembershipWriteFails(t *testing.T) {
	conn, queries := transactionTestDB(t)
	workspaces := NewWorkspaceService(queries, conn)

	_, ownerID, _ := workspaceFixture(t, conn)

	name := "orphan-workspace-" + uuid.NewString()[:8]
	failInsertsInto(t, conn, "workspace_members")

	if _, err := workspaces.Create(context.Background(), ownerID, name); err == nil {
		t.Fatal("Create reported success although the owner membership could not be written")
	}

	var count int
	if err := conn.QueryRowContext(context.Background(),
		`SELECT count(*) FROM workspaces WHERE name = $1`, name).Scan(&count); err != nil {
		t.Fatalf("count workspaces: %v", err)
	}
	if count != 0 {
		t.Errorf("workspace survived a failed owner insert (%d rows named %q)", count, name)
	}
}

// A trigger that raises is an error like any other, so the flow must not turn it into a
// silent success either.
func TestAFailedFlowReportsAnError(t *testing.T) {
	conn, queries := transactionTestDB(t)
	workspaces := NewWorkspaceService(queries, conn)
	workspaces.SetLicenseGate(allowAllGate{})

	workspaceID, ownerID, _ := workspaceFixture(t, conn)
	email := "reported-" + uuid.NewString() + "@example.test"
	deleteUserByEmail(t, conn, email)

	failInsertsInto(t, conn, "workspace_members")

	_, err := workspaces.CreateUserAndAddMember(
		context.Background(), workspaceID, ownerID, email, "Grace Hopper", "Correct-Horse-Battery-1", "viewer",
	)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "forced failure") {
		t.Errorf("error %q does not carry the reason the insert failed", err)
	}
}
